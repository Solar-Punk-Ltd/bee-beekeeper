package dynamicaccess

import (
	"testing"

	"github.com/ethersphere/bee/pkg/crypto"
	"github.com/ethersphere/bee/pkg/storage/inmemchunkstore"
)

func TestHistoryFirstAdd(t *testing.T) {
	_ = inmemchunkstore.New()

	topicStr := "testtopic"
	topic, err := crypto.LegacyKeccak256([]byte(topicStr))
	if err != nil {
		t.Fatal(err)
	}

	pk, _ := crypto.GenerateSecp256k1Key()
	signer := crypto.NewDefaultSigner(pk)
	owner, _ := signer.EthereumAddress()

	// updater, err := mock.HistoryUpdater(storer, signer, topic)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// finder := mock.HistoryFinder(storer, updater.Feed())

	history := NewHistory(topic, owner)

	history.Add(topic, owner, 0, []byte("payload"), []byte("sig"))

}
