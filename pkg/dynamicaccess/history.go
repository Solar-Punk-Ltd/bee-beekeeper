package dynamicaccess

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/ethersphere/bee/pkg/file"
	"github.com/ethersphere/bee/pkg/manifest"
	"github.com/ethersphere/bee/pkg/manifest/mantaray"
	"github.com/ethersphere/bee/pkg/swarm"
)

type History interface {
	Add(ctx context.Context, actRef swarm.Address, timestamp *int64) error
	Lookup(ctx context.Context, timestamp int64, ls file.LoadSaver) (swarm.Address, error)
	Store(ctx context.Context) (swarm.Address, error)
}

var _ History = (*history)(nil)

type history struct {
	manifest *manifest.MantarayManifest
}

func NewHistory(ls file.LoadSaver, ref *swarm.Address) (*history, error) {
	var err error
	var m manifest.Interface

	if ref != nil {
		m, err = manifest.NewDefaultManifestReference(*ref, ls)
	} else {
		m, err = manifest.NewDefaultManifest(ls, false)
	}
	if err != nil {
		return nil, err
	}

	mm, ok := m.(*manifest.MantarayManifest)
	if !ok {
		return nil, fmt.Errorf("expected MantarayManifest, got %T", m)
	}

	return &history{manifest: mm}, nil
}

func (h *history) Manifest() *manifest.MantarayManifest {
	return h.manifest
}

func (h *history) Add(ctx context.Context, actRef swarm.Address, timestamp *int64) error {
	// Do we need any extra meta/act?
	meta := map[string]string{}
	// add timestamps transformed so that the latests timestamp becomes the smallest key
	var unixTime int64
	if timestamp != nil {
		unixTime = *timestamp
	} else {
		unixTime = time.Now().Unix()
	}

	key := strconv.FormatInt(math.MaxInt64-unixTime, 10)
	return h.manifest.Add(ctx, key, manifest.NewEntry(actRef, meta))
}

// Lookup finds the entry for a path or returns error if not found
func (h *history) Lookup(ctx context.Context, timestamp int64, ls file.LoadSaver) (swarm.Address, error) {
	reversedTimestamp := math.MaxInt64 - timestamp
	node, err := h.LookupNode(ctx, reversedTimestamp, ls)
	if err != nil {
		return swarm.Address{}, err
	}

	if node != nil {
		return swarm.NewAddress(node.Entry()), nil
	}

	return swarm.Address{}, nil
}

func (h *history) LookupNode(ctx context.Context, searchedTimestamp int64, ls file.LoadSaver) (*mantaray.Node, error) {
	var node *mantaray.Node

	walker := func(pathTimestamp []byte, currNode *mantaray.Node, err error) error {
		if err != nil {
			return err
		}

		if currNode.IsValueType() && len(currNode.Entry()) > 0 {
			match, err := isMatch(pathTimestamp, searchedTimestamp)
			if match && node == nil {
				node = currNode
				// return error to stop the walk, this is how WalkNode works...
				return errors.New("end iteration")
			}

			return err
		}

		return nil
	}

	rootNode := h.manifest.Root()
	err := rootNode.WalkNode(ctx, []byte{}, ls, walker)

	if node == nil && err != nil {
		return nil, fmt.Errorf("history lookup node error: %w", err)
	}

	return node, nil
}

func (h *history) Store(ctx context.Context) (swarm.Address, error) {
	return h.manifest.Store(ctx)
}

func bytesToInt64(b []byte) (int64, error) {
	if len(b) == 0 {
		return math.MaxInt, nil

	}

	num, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return -1, err
	}

	return num, nil
}

func isMatch(pathTimestamp []byte, searchedTimestamp int64) (bool, error) {
	targetTimestamp, err := bytesToInt64(pathTimestamp)
	if err != nil {
		return false, err
	}
	return searchedTimestamp <= targetTimestamp, nil
}
