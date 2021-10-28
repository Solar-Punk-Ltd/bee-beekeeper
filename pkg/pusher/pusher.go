// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package pusher provides protocol-orchestrating functionality
// over the pushsync protocol. It makes sure that chunks meant
// to be distributed over the network are sent used using the
// pushsync protocol.
package pusher

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ethersphere/bee/pkg/crypto"
	"github.com/ethersphere/bee/pkg/logging"
	"github.com/ethersphere/bee/pkg/postage"
	"github.com/ethersphere/bee/pkg/pushsync"
	"github.com/ethersphere/bee/pkg/storage"
	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/ethersphere/bee/pkg/tags"
	"github.com/ethersphere/bee/pkg/topology"
	"github.com/ethersphere/bee/pkg/tracing"
	"github.com/opentracing/opentracing-go"

	"github.com/sirupsen/logrus"
)

type Service struct {
	networkID         uint64
	storer            storage.Storer
	pushSyncer        pushsync.PushSyncer
	validStamp        postage.ValidStampFn
	depther           topology.NeighborhoodDepther
	logger            logging.Logger
	tag               *tags.Tags
	tracer            *tracing.Tracer
	metrics           metrics
	quit              chan struct{}
	chunksWorkerQuitC chan struct{}
}

var (
	retryInterval  = 5 * time.Second // time interval between retries
	concurrentJobs = 10              // how many chunks to push simultaneously
	retryCount     = 6
)

var (
	ErrInvalidAddress = errors.New("invalid address")
	ErrShallowReceipt = errors.New("shallow recipt")
)

func New(networkID uint64, storer storage.Storer, depther topology.NeighborhoodDepther, pushSyncer pushsync.PushSyncer, validStamp postage.ValidStampFn, tagger *tags.Tags, logger logging.Logger, tracer *tracing.Tracer, warmupTime time.Duration) *Service {
	service := &Service{
		networkID:         networkID,
		storer:            storer,
		pushSyncer:        pushSyncer,
		validStamp:        validStamp,
		depther:           depther,
		tag:               tagger,
		logger:            logger,
		tracer:            tracer,
		metrics:           newMetrics(),
		quit:              make(chan struct{}),
		chunksWorkerQuitC: make(chan struct{}),
	}
	go service.chunksWorker(warmupTime)
	return service
}

// chunksWorker is a loop that keeps looking for chunks that are locally uploaded ( by monitoring pushIndex )
// and pushes them to the closest peer and get a receipt.
func (s *Service) chunksWorker(warmupTime time.Duration) {
	var (
		chunks        <-chan swarm.Chunk
		repeat        func()
		unsubscribe   func()
		timer         = time.NewTimer(0) // timer, initially set to 0 to fall through select case on timer.C for initialisation
		chunksInBatch = -1
		cctx, cancel  = context.WithCancel(context.Background())
		ctx           = cctx
		sem           = make(chan struct{}, concurrentJobs)
		inflight      = make(map[string]struct{})
		mtx           sync.Mutex
		span          opentracing.Span
		logger        *logrus.Entry
		retryCounter  = make(map[string]int)
	)
	// and start iterating on Push index from the beginning
	chunks, repeat, unsubscribe = s.storer.SubscribePush(ctx, func([]byte) bool { return false })

	defer timer.Stop()
	defer close(s.chunksWorkerQuitC)
	go func() {
		<-s.quit
		cancel()
	}()

	// wait for warmup duration to complete
	select {
	case <-time.After(warmupTime):
	case <-s.quit:
		return
	}

	s.logger.Info("pusher: warmup period complete, worker starting.")

LOOP:
	for {
		select {
		// handle incoming chunks
		case ch, more := <-chunks:
			// if no more, set to nil, reset timer to finalise batch
			if !more {
				chunks = nil
				var dur time.Duration
				if chunksInBatch == 0 {
					dur = 500 * time.Millisecond
				}
				timer.Reset(dur)
				break
			}

			// If the stamp is invalid, the chunk is not synced with the network
			// since other nodes would reject the chunk, so the chunk is marked as
			// synced which makes it available to the node but not to the network
			stampBytes, err := ch.Stamp().MarshalBinary()
			if err != nil {
				s.logger.Errorf("pusher: stamp marshal: %w", err)
				if err = s.storer.Set(ctx, storage.ModeSetSync, ch.Address()); err != nil {
					s.logger.Errorf("pusher: set sync: %w", err)
				}
				continue
			}

			_, err = s.validStamp(ch, stampBytes)
			if err != nil {
				s.logger.Warningf("pusher: stamp with batch ID %x is no longer valid, skipping syncing for chunk %s: %v", ch.Stamp().BatchID(), ch.Address().String(), err)
				if err = s.storer.Set(ctx, storage.ModeSetSync, ch.Address()); err != nil {
					s.logger.Errorf("pusher: set sync: %w", err)
				}
				continue
			}

			if span == nil {
				mtx.Lock()
				span, logger, ctx = s.tracer.StartSpanFromContext(cctx, "pusher-sync-batch", s.logger)
				mtx.Unlock()
			}

			// postpone a retry only after we've finished processing everything in index
			timer.Reset(retryInterval)
			chunksInBatch++
			s.metrics.TotalToPush.Inc()

			select {
			case sem <- struct{}{}:
			case <-s.quit:
				if unsubscribe != nil {
					unsubscribe()
				}
				if span != nil {
					span.Finish()
				}

				return
			}
			mtx.Lock()
			if _, ok := inflight[ch.Address().String()]; ok {
				mtx.Unlock()
				<-sem
				continue
			}

			inflight[ch.Address().String()] = struct{}{}
			mtx.Unlock()

			go func(ctx context.Context, ch swarm.Chunk) {
				var (
					err        error
					startTime  = time.Now()
					t          *tags.Tag
					wantSelf   bool
					storerPeer swarm.Address
				)
				defer func() {
					mtx.Lock()
					if err == nil {
						s.metrics.TotalSynced.Inc()
						s.metrics.SyncTime.Observe(time.Since(startTime).Seconds())
						if wantSelf {
							logger.Tracef("pusher: chunk %s stays here, i'm the closest node", ch.Address().String())
						} else {
							po := swarm.Proximity(ch.Address().Bytes(), storerPeer.Bytes())
							logger.Tracef("pusher: pushed chunk %s to node %s, receipt depth %d", ch.Address().String(), storerPeer.String(), po)
							s.metrics.ReceiptDepth.WithLabelValues(strconv.Itoa(int(po))).Inc()
						}
						delete(retryCounter, ch.Address().ByteString())
					} else {
						repeat()
						s.metrics.TotalErrors.Inc()
						s.metrics.ErrorTime.Observe(time.Since(startTime).Seconds())
						logger.Tracef("pusher: cannot push chunk %s: %v", ch.Address().String(), err)
					}
					delete(inflight, ch.Address().String())
					mtx.Unlock()
					<-sem
				}()

				// Later when we process receipt, get the receipt and process it
				// for now ignoring the receipt and checking only for error
				receipt, err := s.pushSyncer.PushChunkToClosest(ctx, ch)
				if err != nil {
					if errors.Is(err, topology.ErrWantSelf) {
						// we are the closest ones - this is fine
						// this is to make sure that the sent number does not diverge from the synced counter
						// the edge case is on the uploader node, in the case where the uploader node is
						// connected to other nodes, but is the closest one to the chunk.
						wantSelf = true
					} else {
						return
					}
				}

				if receipt != nil {
					var publicKey *ecdsa.PublicKey
					publicKey, err = crypto.Recover(receipt.Signature, receipt.Address.Bytes())
					if err != nil {
						err = fmt.Errorf("pusher: receipt recover: %w", err)
						return
					}

					storerPeer, err = crypto.NewOverlayAddress(*publicKey, s.networkID, receipt.BlockHash)
					if err != nil {
						err = fmt.Errorf("pusher: receipt storer address: %w", err)
						return
					}

					po := swarm.Proximity(ch.Address().Bytes(), storerPeer.Bytes())
					d := s.depther.NeighborhoodDepth()
					if po < d {
						mtx.Lock()
						retryCounter[ch.Address().ByteString()]++
						if retryCounter[ch.Address().ByteString()] < retryCount {
							mtx.Unlock()
							err = fmt.Errorf("pusher: shallow receipt depth %d, want at least %d", po, d)
							s.metrics.ShallowReceiptDepth.WithLabelValues(strconv.Itoa(int(po))).Inc()
							return
						}
						mtx.Unlock()
					} else {
						s.metrics.ReceiptDepth.WithLabelValues(strconv.Itoa(int(po))).Inc()
					}
				}

				if err = s.storer.Set(ctx, storage.ModeSetSync, ch.Address()); err != nil {
					err = fmt.Errorf("pusher: set sync: %w", err)
					return
				}
				if ch.TagID() > 0 {
					// for individual chunks uploaded using the
					// /chunks api endpoint the tag will be missing
					// by default, unless the api consumer specifies one
					t, err = s.tag.Get(ch.TagID())
					if err == nil && t != nil {
						err = t.Inc(tags.StateSynced)
						if err != nil {
							err = fmt.Errorf("increment synced: %w", err)
							logger.Tracef("pusher: chunk %s tag error: %v", ch.Address().String(), err)
							err = nil
							return
						}
						if wantSelf {
							err = t.Inc(tags.StateSent)
							if err != nil {
								err = fmt.Errorf("increment sent: %w", err)
								logger.Tracef("pusher: chunk %s tag error: %v", ch.Address().String(), err)
								err = nil
								return
							}
						}
					} else {
						err = nil // don't fail because of a tag error
					}
				}

			}(ctx, ch)
		case <-timer.C:
			// initially timer is set to go off as well as every time we hit the end of push index
			startTime := time.Now()

			chunksInBatch = 0

			// reset timer to go off after retryInterval
			timer.Reset(retryInterval)
			s.metrics.MarkAndSweepTime.Observe(time.Since(startTime).Seconds())

			if span != nil {
				span.Finish()
				span = nil
			}

		case <-s.quit:
			if unsubscribe != nil {
				unsubscribe()
			}
			if span != nil {
				span.Finish()
			}

			break LOOP
		}
	}

	// wait for all pending push operations to terminate
	closeC := make(chan struct{})
	go func() {
		defer func() { close(closeC) }()
		for i := 0; i < cap(sem); i++ {
			sem <- struct{}{}
		}
	}()

	select {
	case <-closeC:
	case <-time.After(5 * time.Second):
		s.logger.Warning("pusher shutting down with pending operations")
	}
}

func (s *Service) Close() error {
	s.logger.Info("pusher shutting down")
	close(s.quit)

	// Wait for chunks worker to finish
	select {
	case <-s.chunksWorkerQuitC:
	case <-time.After(6 * time.Second):
	}
	return nil
}
