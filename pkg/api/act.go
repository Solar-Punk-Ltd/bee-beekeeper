package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"net/http"

	"github.com/ethersphere/bee/v2/pkg/swarm"
	"github.com/gorilla/mux"
)

func deserializeBytes(data []byte) *ecdsa.PublicKey {
	curve := elliptic.P256()
	x, y := elliptic.Unmarshal(curve, data)
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}
}

func (s *Service) actDownHandler() func(h http.Handler) http.Handler {
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
				Timestamp      int64         `map:"Swarm-Act-Timestamp"`
				Publisher      []byte        `map:"Swarm-Act-Publisher"`
				HistoryAddress swarm.Address `map:"Swarm-Act-History-Address"`
			}{}

			if response := s.mapStructure(mux.Vars(r), &headers); response != nil {
				response("invalid path params", logger, w)
				return
			}
			// TODO: historyRootHah
			// TODO: refactor DownloadHandler to accept a context
			// TODO: if ACT?
			ctx := r.Context()
			ref, err := s.dac.DownloadHandler(ctx, headers.Timestamp, paths.Address, deserializeBytes(headers.Publisher), headers.HistoryAddress)
			if err != nil {
				return
			}
			w.Header().Set("address", ref.String())
			h.ServeHTTP(w, r)
		})
	}

}

func (s *Service) actUpHandler() func(h http.Handler) http.Handler {
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
				Timestamp      int64         `map:"Swarm-Act-Timestamp"`
				Publisher      []byte        `map:"Swarm-Act-Publisher"`
				HistoryAddress swarm.Address `map:"Swarm-Act-History-Address"`
			}{}

			if response := s.mapStructure(mux.Vars(r), &headers); response != nil {
				response("invalid path params", logger, w)
				return
			}
			// TODO: historyRootHah
			// TODO: refactor DownloadHandler to accept a context
			// TODO: if ACT?
			ctx := r.Context()
			_, encryptedRef, err := s.dac.UploadHandler(ctx, paths.Address, deserializeBytes(headers.Publisher), headers.HistoryAddress)
			if err != nil {
				return
			}
			w.Header().Set("address", encryptedRef.String())
			h.ServeHTTP(w, r)
		})
	}

}
