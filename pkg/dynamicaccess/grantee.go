package dynamicaccess

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ethersphere/bee/v2/pkg/file"
	"github.com/ethersphere/bee/v2/pkg/swarm"
)

const (
	publicKeyLen = 65
)

type GranteeList interface {
	Add(addList []*ecdsa.PublicKey) error
	Remove(removeList []*ecdsa.PublicKey) error
	Get() []*ecdsa.PublicKey
	Save(ctx context.Context) (swarm.Address, error)
}

type GranteeListStruct struct {
	grantees []*ecdsa.PublicKey
	loadSave file.LoadSaver
}

var _ GranteeList = (*GranteeListStruct)(nil)

func (g *GranteeListStruct) Get() []*ecdsa.PublicKey {
	return g.grantees
}

func (g *GranteeListStruct) Add(addList []*ecdsa.PublicKey) error {
	if len(addList) == 0 {
		return fmt.Errorf("no public key provided")
	}
	g.grantees = append(g.grantees, addList...)

	return nil
}

func (g *GranteeListStruct) Save(ctx context.Context) (swarm.Address, error) {
	data := serialize(g.grantees)
	refBytes, err := g.loadSave.Save(ctx, data)
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("grantee save error: %w", err)
	}

	return swarm.NewAddress(refBytes), nil
}

func (g *GranteeListStruct) Remove(keysToRemove []*ecdsa.PublicKey) error {
	if len(keysToRemove) == 0 {
		return fmt.Errorf("nothing to remove")
	}

	if len(g.grantees) == 0 {
		return fmt.Errorf("no grantee found")
	}
	grantees := g.grantees

	for _, remove := range keysToRemove {
		for i := 0; i < len(grantees); i++ {
			if grantees[i].Equal(remove) {
				grantees[i] = grantees[len(grantees)-1]
				grantees = grantees[:len(grantees)-1]
			}
		}
	}
	g.grantees = grantees

	return nil
}

func NewGranteeList(ls file.LoadSaver) GranteeList {
	return &GranteeListStruct{
		grantees: []*ecdsa.PublicKey{},
		loadSave: ls,
	}
}

func NewGranteeListReference(ls file.LoadSaver, reference swarm.Address) GranteeList {
	data, err := ls.Load(context.Background(), reference.Bytes())
	if err != nil {
		return nil
	}
	grantees := deserialize(data)

	return &GranteeListStruct{
		grantees: grantees,
		loadSave: ls,
	}
}

func serialize(publicKeys []*ecdsa.PublicKey) []byte {
	b := make([]byte, 0, len(publicKeys)*publicKeyLen)
	for _, key := range publicKeys {
		b = append(b, serializePublicKey(key)...)
	}
	return b
}

func serializePublicKey(pub *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(pub.Curve, pub.X, pub.Y)
}

func deserialize(data []byte) []*ecdsa.PublicKey {
	if len(data) == 0 {
		return []*ecdsa.PublicKey{}
	}

	p := make([]*ecdsa.PublicKey, 0, len(data)/publicKeyLen)
	for i := 0; i < len(data); i += publicKeyLen {
		pubKey := deserializeBytes(data[i : i+publicKeyLen])
		if pubKey == nil {
			return []*ecdsa.PublicKey{}
		}
		p = append(p, pubKey)
	}
	return p
}

func deserializeBytes(data []byte) *ecdsa.PublicKey {
	key, err := btcec.ParsePubKey(data)
	if err != nil {
		return nil
	}
	return key.ToECDSA()
}
