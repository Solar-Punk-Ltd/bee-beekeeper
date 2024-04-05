package mock

import (
	"sort"
	"time"

	"github.com/ethersphere/bee/v2/pkg/kvs"
)

type historyMock struct {
	history map[int64]kvs.KeyValueStore
}

func NewHistory() *historyMock {
	return &historyMock{history: make(map[int64]kvs.KeyValueStore)}
}

func (h *historyMock) Add(timestamp int64, act kvs.KeyValueStore) error {
	h.history[timestamp] = act
	return nil
}

func (h *historyMock) Insert(timestamp int64, act kvs.KeyValueStore) *historyMock {
	h.Add(timestamp, act)
	return h
}

func (h *historyMock) Lookup(at int64) (kvs.KeyValueStore, error) {
	keys := []int64{}
	for k := range h.history {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	timestamp := time.Now()
	if at != 0 {
		timestamp = time.Unix(at, 0)
	}

	for i := len(keys) - 1; i >= 0; i-- {
		update := time.Unix(keys[i], 0)
		if update.Before(timestamp) || update.Equal(timestamp) {
			return h.history[keys[i]], nil
		}
	}
	return nil, nil
}

func (h *historyMock) Get(timestamp int64) (kvs.KeyValueStore, error) {
	return h.history[timestamp], nil
}
