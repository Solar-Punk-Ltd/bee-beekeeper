package kvs

import (
	"context"
	"encoding/hex"
	"sync"

	"github.com/ethersphere/bee/pkg/api"
	"github.com/ethersphere/bee/pkg/file/loadsave"
	"github.com/ethersphere/bee/pkg/file/pipeline"
	"github.com/ethersphere/bee/pkg/file/pipeline/builder"
	"github.com/ethersphere/bee/pkg/file/redundancy"
	"github.com/ethersphere/bee/pkg/manifest"
	"github.com/ethersphere/bee/pkg/storage"
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

type KeyValueStore interface {
	Get(rootHash swarm.Address, key []byte) ([]byte, error)
	Put(rootHash swarm.Address, key, value []byte) (swarm.Address, error)
}

type memoryKeyValueStore struct {
	rootHash swarm.Address
}

type manifestKeyValueStore struct {
	// m manifest.Interface
	// rootHash swarm.Address
	storer api.Storer
}

/*
	type potKeyValueStore struct {
		rootHash swarm.Address
		p        ProximityOrderTrie
	}
*/
func (m *memoryKeyValueStore) Get(rootHash swarm.Address, key []byte) ([]byte, error) {
	mem := getMemory()
	val := mem[hex.EncodeToString(key)]
	return val, nil
}

func (m *memoryKeyValueStore) Put(rootHash swarm.Address, key []byte, value []byte) (swarm.Address, error) {
	mem := getMemory()
	mem[hex.EncodeToString(key)] = value
	return swarm.EmptyAddress, nil
}

func NewmemoryKeyValueStore(rootHash swarm.Address) KeyValueStore {
	return &memoryKeyValueStore{
		rootHash: rootHash,
	}
}

// TODO: pass context as dep.
func (m *manifestKeyValueStore) Get(rootHash swarm.Address, key []byte) ([]byte, error) {
	ls := loadsave.NewReadonly(m.storer.ChunkStore())
	// existing manif
	manif, err := manifest.NewSimpleManifestReference(rootHash, ls)
	if err != nil {
		return nil, err
	}
	entry, err := manif.Lookup(context.Background(), hex.EncodeToString(key))
	if err != nil {
		return nil, err
	}
	ref := entry.Reference()
	return ref.Bytes(), nil
}

func (m *manifestKeyValueStore) Put(rootHash swarm.Address, key []byte, value []byte) (swarm.Address, error) {
	factory := requestPipelineFactory(context.Background(), m.storer.Cache(), false, redundancy.NONE)
	ls := loadsave.New(m.storer.ChunkStore(), m.storer.Cache(), factory)
	// existing manif
	manif, err := manifest.NewSimpleManifestReference(rootHash, ls)
	if err != nil {
		// new manif
		manif, err = manifest.NewSimpleManifest(ls)
		if err != nil {
			return swarm.EmptyAddress, err
		}
	}
	err = manif.Add(context.Background(), hex.EncodeToString(key), manifest.NewEntry(swarm.NewAddress(value), map[string]string{}))
	if err != nil {
		return swarm.EmptyAddress, err
	}
	manifRef, err := manif.Store(context.Background())
	if err != nil {
		return swarm.EmptyAddress, err
	}

	session := m.storer.DirectUpload()
	err = session.Done(manifRef)
	if err != nil {
		return swarm.EmptyAddress, err
	}
	return manifRef, nil
}

func requestPipelineFactory(ctx context.Context, s storage.Putter, encrypt bool, rLevel redundancy.Level) func() pipeline.Interface {
	return func() pipeline.Interface {
		return builder.NewPipelineBuilder(ctx, s, encrypt, rLevel)
	}
}

func NewManifestKeyValueStore( /*rootHash swarm.Address,*/ storer api.Storer) KeyValueStore {
	return &manifestKeyValueStore{
		// rootHash: rootHash,
		storer: storer,
	}
}

/*
func (p *ProximityOrderTrie) Get(key []byte) ([]byte, error) {
	manifest, err := p.p.Lookup(context.Background(), key)
	ref := manifest.Reference()
	return ref.Bytes(), err
}

func (p *ProximityOrderTrie) Put(key []byte, value []byte) error {
	p.p.Add(context.Background(), key, manifest.NewEntry(swarm.NewAddress([]byte(value)), nil))
	return nil
}

func NewPotKeyValueStore(rootHash swarm.Address) KeyValueStore {
	m, _ := pot.NewPot(rootHash, nil)
	return &potKeyValueStore{
		p: m,
	}
}

type ProximityOrderTrie struct {
}

func (*ProximityOrderTrie) Add(context.Context, string, manifest.Entry) error
func (*ProximityOrderTrie) Lookup(context.Context, string) (manifest.Entry, error)
*/
