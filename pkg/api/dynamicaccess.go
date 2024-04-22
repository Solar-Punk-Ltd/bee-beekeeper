package api

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ethersphere/bee/v2/pkg/crypto"
	"github.com/ethersphere/bee/v2/pkg/jsonhttp"
	storer "github.com/ethersphere/bee/v2/pkg/storer"
	"github.com/ethersphere/bee/v2/pkg/swarm"
	"github.com/gorilla/mux"
)

type addressKey struct{}

// getAddressFromContext is a helper function to extract the address from the context
func getAddressFromContext(ctx context.Context) swarm.Address {
	v, ok := ctx.Value(addressKey{}).(swarm.Address)
	if ok {
		return v
	}
	return swarm.ZeroAddress
}

// setAddress sets the swarm address in the context
func setAddressInContext(ctx context.Context, address swarm.Address) context.Context {
	return context.WithValue(ctx, addressKey{}, address)
}

type GranteesPatchRequest struct {
	Addlist    []string `json:"add"`
	Revokelist []string `json:"revoke"`
}

type GranteesPatchResponse struct {
	Reference        swarm.Address `json:"ref"`
	HistoryReference swarm.Address `json:"historyref"`
}

type GranteesPostRequest struct {
	GranteeList []string `json:"grantees"`
}

type GranteesPostResponse struct {
	Reference        swarm.Address `json:"ref"`
	HistoryReference swarm.Address `json:"historyref"`
}
type GranteesPatch struct {
	Addlist    []ecdsa.PublicKey
	Revokelist []ecdsa.PublicKey
}

// actDecryptionHandler is a middleware that looks up and decrypts the given address,
// if the act headers are present
func (s *Service) actDecryptionHandler() func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := s.logger.WithName("act_decryption_handler").Build()
			paths := struct {
				Address swarm.Address `map:"address,resolve" validate:"required"`
			}{}
			if response := s.mapStructure(mux.Vars(r), &paths); response != nil {
				response("invalid path params", logger, w)
				return
			}

			headers := struct {
				Timestamp      *int64           `map:"Swarm-Act-Timestamp"`
				Publisher      *ecdsa.PublicKey `map:"Swarm-Act-Publisher"`
				HistoryAddress *swarm.Address   `map:"Swarm-Act-History-Address"`
			}{}
			if response := s.mapStructure(r.Header, &headers); response != nil {
				response("invalid header params", logger, w)
				return
			}

			// Try to download the file wihtout decryption, if the act headers are not present
			if headers.Publisher == nil || headers.Timestamp == nil || headers.HistoryAddress == nil {
				h.ServeHTTP(w, r)
				return
			}
			ctx := r.Context()
			reference, err := s.dac.DownloadHandler(ctx, paths.Address, headers.Publisher, *headers.HistoryAddress, *headers.Timestamp)
			if err != nil {
				jsonhttp.InternalServerError(w, errActDownload)
				return
			}
			h.ServeHTTP(w, r.WithContext(setAddressInContext(ctx, reference)))
		})
	}

}

// actEncryptionHandler is a middleware that encrypts the given address using the publisher's public key
// Uploads the encrypted reference, history and kvs to the store
func (s *Service) actEncryptionHandler(
	ctx context.Context,
	w http.ResponseWriter,
	putter storer.PutterSession,
	reference swarm.Address,
	historyRootHash swarm.Address,
) (swarm.Address, error) {
	logger := s.logger.WithName("act_encryption_handler").Build()
	publisherPublicKey := &s.publicKey
	storageReference, historyReference, encryptedReference, err := s.dac.UploadHandler(ctx, reference, publisherPublicKey, historyRootHash)
	if err != nil {
		logger.Debug("act failed to encrypt reference", "error", err)
		logger.Error(nil, "act failed to encrypt reference")
		return swarm.ZeroAddress, err
	}
	// only need to upload history and kvs if a new history is created,
	// meaning that the publsher uploaded to the history for the first time
	if !historyReference.Equal(historyRootHash) {
		err = putter.Done(storageReference)
		if err != nil {
			logger.Debug("done split keyvaluestore failed", "error", err)
			logger.Error(nil, "done split keyvaluestore failed")
			return swarm.ZeroAddress, err
		}
		err = putter.Done(historyReference)
		if err != nil {
			logger.Debug("done split history failed", "error", err)
			logger.Error(nil, "done split history failed")
			return swarm.ZeroAddress, err
		}
	}
	err = putter.Done(encryptedReference)
	if err != nil {
		logger.Debug("done split encrypted reference failed", "error", err)
		logger.Error(nil, "done split encrypted reference failed")
		return swarm.ZeroAddress, err
	}

	w.Header().Set(SwarmActHistoryAddressHeader, historyReference.String())

	return encryptedReference, nil
}

func (s *Service) actListGranteesHandler(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithName("act_list_grantees_handler").Build()
	paths := struct {
		GranteesAddress swarm.Address `map:"address,resolve" validate:"required"`
	}{}
	if response := s.mapStructure(mux.Vars(r), &paths); response != nil {
		response("invalid path params", logger, w)
		return
	}
	grantees, err := s.dac.GetGrantees(r.Context(), paths.GranteesAddress)
	if err != nil {
		jsonhttp.NotFound(w, "grantee list not found")
		return
	}
	granteeSlice := make([]string, len(grantees))
	for i, grantee := range grantees {
		granteeSlice[i] = hex.EncodeToString(crypto.EncodeSecp256k1PublicKey(grantee))
	}
	jsonhttp.OK(w, granteeSlice)
}

func (s *Service) actGrantRevokeHandler(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithName("act_grant_revoke_handler").Build()

	if r.Body == http.NoBody {
		logger.Error(nil, "request has no body")
		jsonhttp.BadRequest(w, errInvalidRequest)
		return
	}

	paths := struct {
		GranteesAddress swarm.Address `map:"address,resolve" validate:"required"`
	}{}
	if response := s.mapStructure(mux.Vars(r), &paths); response != nil {
		response("invalid path params", logger, w)
		return
	}

	headers := struct {
		BatchID        []byte         `map:"Swarm-Postage-Batch-Id" validate:"required"`
		HistoryAddress *swarm.Address `map:"Swarm-Act-History-Address" validate:"required"`
	}{}
	if response := s.mapStructure(r.Header, &headers); response != nil {
		response("invalid header params", logger, w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		if jsonhttp.HandleBodyReadError(err, w) {
			return
		}
		logger.Debug("read request body failed", "error", err)
		logger.Error(nil, "read request body failed")
		jsonhttp.InternalServerError(w, "cannot read request")
		return
	}

	gpr := GranteesPatchRequest{}
	if len(body) > 0 {
		err = json.Unmarshal(body, &gpr)
		if err != nil {
			logger.Debug("unmarshal body failed", "error", err)
			logger.Error(nil, "unmarshal body failed")
			jsonhttp.InternalServerError(w, "error unmarshaling request body")
			return
		}
	}

	grantees := GranteesPatch{}
	for _, g := range gpr.Addlist {
		h, _ := hex.DecodeString(g)
		k, _ := btcec.ParsePubKey(h)
		grantees.Addlist = append(grantees.Addlist, *k.ToECDSA())
	}
	for _, g := range gpr.Revokelist {
		h, _ := hex.DecodeString(g)
		k, _ := btcec.ParsePubKey(h)
		grantees.Revokelist = append(grantees.Revokelist, *k.ToECDSA())
	}

	tag, _ := s.getOrCreateSessionID(0)

	ctx := r.Context()
	putter, _ := s.newStamperPutter(ctx, putterOptions{
		BatchID:  headers.BatchID,
		TagID:    tag,
		Pin:      false,
		Deferred: true,
	})

	granteeref := paths.GranteesAddress
	granteeref, historyref, _ := s.dac.HandleGrantees(ctx, granteeref, *headers.HistoryAddress, &s.publicKey, convertToPointerSlice(grantees.Addlist), convertToPointerSlice(grantees.Revokelist))
	putter.Done(granteeref)
	putter.Done(historyref)
	jsonhttp.OK(w, GranteesPatchResponse{
		Reference:        granteeref,
		HistoryReference: historyref,
	})
}

func (s *Service) actCreateGranteesHandler(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithName("acthandler").Build()

	if r.Body == http.NoBody {
		logger.Error(nil, "request has no body")
		jsonhttp.BadRequest(w, errInvalidRequest)
		return
	}

	headers := struct {
		BatchID []byte `map:"Swarm-Postage-Batch-Id" validate:"required"`
	}{}
	if response := s.mapStructure(r.Header, &headers); response != nil {
		response("invalid header params", logger, w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		if jsonhttp.HandleBodyReadError(err, w) {
			return
		}
		logger.Debug("read request body failed", "error", err)
		logger.Error(nil, "read request body failed")
		jsonhttp.InternalServerError(w, "cannot read request")
		return
	}

	gpr := GranteesPostRequest{}
	if len(body) > 0 {
		err = json.Unmarshal(body, &gpr)
		if err != nil {
			logger.Debug("unmarshal body failed", "error", err)
			logger.Error(nil, "unmarshal body failed")
			jsonhttp.InternalServerError(w, "error unmarshaling request body")
			return
		}
	}

	list := make([]ecdsa.PublicKey, len(gpr.GranteeList))
	for _, g := range gpr.GranteeList {
		h, _ := hex.DecodeString(g)
		k, _ := btcec.ParsePubKey(h)
		list = append(list, *k.ToECDSA())
	}
	tag, _ := s.getOrCreateSessionID(0)

	ctx := r.Context()
	putter, _ := s.newStamperPutter(ctx, putterOptions{
		BatchID:  headers.BatchID,
		TagID:    tag,
		Pin:      false,
		Deferred: true,
	})

	granteeref, historyref, _ := s.dac.HandleGrantees(ctx, swarm.ZeroAddress, swarm.ZeroAddress, &s.publicKey, convertToPointerSlice(list), nil)
	putter.Done(granteeref)
	putter.Done(historyref)
	jsonhttp.Created(w, GranteesPostResponse{
		Reference:        granteeref,
		HistoryReference: historyref,
	})
}

func convertToPointerSlice(slice []ecdsa.PublicKey) []*ecdsa.PublicKey {
	pointerSlice := make([]*ecdsa.PublicKey, len(slice))
	for i, key := range slice {
		pointerSlice[i] = &key
	}
	return pointerSlice
}
