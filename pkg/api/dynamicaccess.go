package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"net/http"

	"github.com/ethersphere/bee/v2/pkg/swarm"
)

func deserializeBytes(data []byte) *ecdsa.PublicKey {
	curve := elliptic.P256()
	x, y := elliptic.Unmarshal(curve, data)
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}
}

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
			// TODO: use int and swarm.Address,bool... in header, test should convert
			headers := struct {
				Act            bool             `map:"Swarm-Act"`
				Timestamp      int64            `map:"Swarm-Act-Timestamp"`
				Publisher      *ecdsa.PublicKey `map:"Swarm-Act-Publisher"`
				HistoryAddress swarm.Address    `map:"Swarm-Act-History-Address"`
			}{}
			if response := s.mapStructure(r.Header, &headers); response != nil {
				response("invalid path params", logger, w)
				return
			}
			if !headers.Act {
				h.ServeHTTP(w, r)
				return
			}
			ctx := r.Context()
			ref, err := s.dac.DownloadHandler(ctx, headers.Timestamp, paths.Address, headers.Publisher, headers.HistoryAddress)
			if err != nil {
				return
			}
			w.Header().Set("address", ref.String())
			h.ServeHTTP(w, r)
		})
	}

}

// func (s *Service) actEncryptionHandler(publisher *ecdsa.PublicKey, history, reference swarm.Address) (swarm.Address, swarm.Address, error) {
// 	ctx := context.Background()
// 	return s.dac.UploadHandler(ctx, reference, publisher, history)
// }
