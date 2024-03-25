package mock

import (
	"encoding/hex"
	"sync"

	"github.com/ethersphere/bee/pkg/kvs"
	"github.com/ethersphere/bee/pkg/swarm"
)

var lock = &sync.Mutex{}

type single struct {
	// TODO string -> []byte ?
	memoryMock map[string][]byte
}

var singleInMemorySwarm *single

func getInMemorySwarm() *single {
	if singleInMemorySwarm == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleInMemorySwarm == nil {
			singleInMemorySwarm = &single{
				memoryMock: make(map[string][]byte)}
		}
	}
	return singleInMemorySwarm
}

func getMemory() map[string][]byte {
	ch := make(chan *single)
	go func() {
		ch <- getInMemorySwarm()
	}()
	mem := <-ch
	return mem.memoryMock
}

type mockKeyValueStore struct {
}

func (m *mockKeyValueStore) Get(rootHash swarm.Address, key []byte) ([]byte, error) {
	mem := getMemory()
	val := mem[hex.EncodeToString(key)]
	return val, nil
}

func (m *mockKeyValueStore) Put(rootHash swarm.Address, key []byte, value []byte) (swarm.Address, error) {
	mem := getMemory()
	mem[hex.EncodeToString(key)] = value
	return swarm.EmptyAddress, nil
}

func New() kvs.KeyValueStore {
	return &mockKeyValueStore{}
}
