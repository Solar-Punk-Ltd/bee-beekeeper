package api

import (
	"crypto/ecdsa"
	"net/http"

	"github.com/ethersphere/bee/v2/pkg/file/redundancy"
	"github.com/ethersphere/bee/v2/pkg/swarm"
)

func (s *Service) actDecrpytionHandler() func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := s.logger.WithName("acthandler").Build()
			paths := struct {
				Address swarm.Address `map:"address,resolve" validate:"required"`
			}{}
			if response := s.mapStructure(r.Header, &paths); response != nil {
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
			if headers.Publisher == nil || headers.Timestamp == nil || headers.HistoryAddress == nil {
				h.ServeHTTP(w, r)
				return
			}

			reference, err := s.dac.DownloadHandler(r.Context(), *headers.Timestamp, paths.Address, headers.Publisher, *headers.HistoryAddress, headers.Encrypt, headers.RLevel)
			if err != nil {
				return
			}
			w.Header().Set("address", reference.String())
			h.ServeHTTP(w, r)
		})
	}

}

// func (s *Service) actEncryptionHandler(publisher *ecdsa.PublicKey, history, reference swarm.Address) (swarm.Address, swarm.Address, error) {
// 	ctx := context.Background()
// 	return s.dac.UploadHandler(ctx, reference, publisher, history)
// }
