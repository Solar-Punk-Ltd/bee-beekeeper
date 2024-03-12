package mock

import (
	"context"

	"github.com/ethersphere/bee/pkg/crypto"
	"github.com/ethersphere/bee/pkg/feeds"
	"github.com/ethersphere/bee/pkg/storage"
	"github.com/ethersphere/bee/pkg/swarm"
)

type finder struct {
	getter *feeds.Getter
}

type updater struct {
	*feeds.Putter
	next uint64
}

func (f *finder) At(ctx context.Context, at int64, after uint64) (chunk swarm.Chunk, currentIndex, nextIndex feeds.Index, err error) {
	return nil, nil, nil, nil
}

func HistoryFinder(getter storage.Getter, feed *feeds.Feed) feeds.Lookup {
	return &finder{feeds.NewGetter(getter, feed)}
}

func (u *updater) Update(ctx context.Context, at int64, payload []byte) error {
	return nil
}

func (u *updater) Feed() *feeds.Feed {
	return nil
}

func HistoryUpdater(putter storage.Putter, signer crypto.Signer, topic []byte) (feeds.Updater, error) {
	p, err := feeds.NewPutter(putter, signer, topic)
	if err != nil {
		return nil, err
	}
	return &updater{Putter: p}, nil
}
