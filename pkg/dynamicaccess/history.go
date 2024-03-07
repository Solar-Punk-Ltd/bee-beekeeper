package dynamicaccess

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethersphere/bee/pkg/feeds"
)

type History interface {
	Add(topic []byte, owner common.Address, timestamp int64, payload, sig []byte) error
	Get(topic []byte, owner common.Address, timestamp int64) error
}

var _ History = (*history)(nil)

type history struct {
	Feed *feeds.Feed
}

func NewHistory(topic []byte, owner common.Address) *history {
	return &history{Feed: &feeds.Feed{Topic: topic, Owner: owner}}
}

func (h *history) Add(topic []byte, owner common.Address, timestamp int64, payload, sig []byte) error {
	// use timestamp based feed update
	return nil
}

func (h *history) Get(topic []byte, owner common.Address, timestamp int64) error {
	// get the feed
	return nil
}
