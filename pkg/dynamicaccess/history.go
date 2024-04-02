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
	Add(ctx context.Context, actRef swarm.Address) error
	Lookup(timestamp string) (swarm.Address, error)
}

// var _ History = (*history)(nil)

type history struct {
	manifest *manifest.MantarayManifest
}

func NewHistory(ls file.LoadSaver) (*history, error) {
	m, err := manifest.NewDefaultManifest(ls, false)
	if err != nil {
		return nil, err
	}

	mm, ok := m.(*manifest.MantarayManifest)
	if !ok {
		return nil, fmt.Errorf("Expected MantarayManifest, got %T", m)
	}

	return &history{manifest: mm}, nil
}

func (h *history) Add(ctx context.Context, actRef swarm.Address) error {
	// Do we need any extra meta/act?
	meta := map[string]string{}
	// add timestamps transformed so that the latests timestamp becomes the smallest key
	unixTime := time.Now().Unix()
	key := strconv.FormatInt(math.MaxInt64-unixTime, 10)
	return h.manifest.Add(ctx, key, manifest.NewEntry(actRef, meta))
}

// Lookup finds the entry for a path or returns error if not found
func (h *history) Lookup(ctx context.Context, timestamp int64, ls file.LoadSaver) swarm.Address {
	node := h.LookupNode(ctx, timestamp, ls)
	if node != nil {
		return swarm.NewAddress(node.Entry())
	}
	return swarm.Address{}
}

func (h *history) LookupNode(ctx context.Context, searchedTimestamp int64, ls file.LoadSaver) *mantaray.Node {
	var node *mantaray.Node

	walker := func(pathTimestamp []byte, walkedNode *mantaray.Node, err error) error {
		if err != nil {
			return err
		}

		if walkedNode != nil {
			match, err := isMatch(pathTimestamp, searchedTimestamp)
			if match {
				node = walkedNode
				// return error to stop the walk, this is how WalkNode works...
				return errors.New("End iteration!")
			}
			return err
		}
		return nil
	}

	rootNode := h.manifest.Root()
	node, err := rootNode.WalkNode(ctx, []byte{}, ls, walker)
	if err != nil {
		fmt.Errorf("History lookup node error: %w", err)
		return nil
	}

	return node
}

func int64ToBytes(num int64) []byte {
	return []byte(strconv.FormatInt(num, 10))
}

func bytesToInt64(b []byte) (int64, error) {
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
	return searchedTimestamp >= targetTimestamp, nil
}
