package api

import (
	"context"
	"crypto/ecdsa"
	"net/http"

	"github.com/ethersphere/bee/v2/pkg/file/redundancy"
	"github.com/ethersphere/bee/v2/pkg/jsonhttp"
	"github.com/ethersphere/bee/v2/pkg/swarm"
	"github.com/gorilla/mux"
)

type addressKey struct{}

// getAddressFromContext is a helper function to extract the address from the context
func getAddressFromContext(ctx context.Context) swarm.Address {
	v, ok := ctx.Value(addressKey{}).([]byte)
	if ok {
		return swarm.NewAddress(v)
	}
	return swarm.ZeroAddress
}

// setAddress sets the redundancy level in the context
func setAddress(ctx context.Context, address []byte) context.Context {
	return context.WithValue(ctx, addressKey{}, address)
}

func (s *Service) actDecrpytionHandler() func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := s.logger.WithName("acthandler").Build()
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
				Encrypt        bool             `map:"Swarm-Encrypt"`
				RLevel         redundancy.Level `map:"Swarm-Redundancy-Level"`
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
			reference, err := s.dac.DownloadHandler(ctx, *headers.Timestamp, paths.Address, headers.Publisher, *headers.HistoryAddress, headers.Encrypt, headers.RLevel)
			if err != nil {
				jsonhttp.InternalServerError(w, "failed to get reference from act")
				return
			}
			h.ServeHTTP(w, r.WithContext(setAddress(ctx, reference.Bytes())))
		})
	}

}
