package dynamicaccess

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"

	"github.com/ethersphere/bee/pkg/file"
	"github.com/ethersphere/bee/pkg/storer"
	"github.com/ethersphere/bee/pkg/swarm"
)

const (
	publicKeyLen = 65
)

// TODO: maybe rename to "List", simply
type GranteeList interface {
	Add(publicKeys []*ecdsa.PublicKey) error
	Remove(removeList []*ecdsa.PublicKey) error
	Get() []*ecdsa.PublicKey
	Save() (swarm.Address, error)
}

type GranteeListStruct struct {
	grantees []byte
	loadSave file.LoadSaver
	putter   storer.PutterSession
}

var _ GranteeList = (*GranteeListStruct)(nil)

func (g *GranteeListStruct) Get() []*ecdsa.PublicKey {
	return g.Deserialize(g.grantees)
}

func (g *GranteeListStruct) Serialize(publicKeys []*ecdsa.PublicKey) []byte {
	b := make([]byte, 0, len(publicKeys))
	for _, key := range publicKeys {
		b = append(b, g.SerializePublicKey(key)...)
	}
	return b
}

func (g *GranteeListStruct) SerializePublicKey(pub *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(pub.Curve, pub.X, pub.Y)
}

func (g *GranteeListStruct) Deserialize(data []byte) []*ecdsa.PublicKey {
	if len(data) == 0 {
		return nil
	}

	p := make([]*ecdsa.PublicKey, 0, len(data)/publicKeyLen)
	for i := 0; i < len(data); i += publicKeyLen {
		pubKey := g.DeserializeBytes(data[i : i+publicKeyLen])
		if pubKey == nil {
			return nil
		}
		p = append(p, pubKey)
	}
	return p
}

func (g *GranteeListStruct) DeserializeBytes(data []byte) *ecdsa.PublicKey {
	curve := elliptic.P256()
	// TODO: use not deprecated ecdsa, ecdh instead
	// pub, err := ecdh.P256().NewPublicKey(data)
	// if err != nil {
	// 	return nil
	// }
	// return pub
	x, y := elliptic.Unmarshal(curve, data)
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}
}

func (g *GranteeListStruct) Add(publicKeys []*ecdsa.PublicKey) error {
	if len(publicKeys) == 0 {
		return fmt.Errorf("no public key provided")
	}

	data := g.Serialize(publicKeys)
	g.grantees = append(g.grantees, data...)
	return nil
}

func (g *GranteeListStruct) Save() (swarm.Address, error) {
	refBytes, err := g.loadSave.Save(context.Background(), g.grantees)
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("grantee save error: %w", err)
	}
	address := swarm.NewAddress(refBytes)
	err = g.putter.Done(address)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	return address, nil
}

func (g *GranteeListStruct) Remove(keysToRemove []*ecdsa.PublicKey) error {
	grantees := g.Deserialize(g.grantees)
	if grantees == nil {
		return fmt.Errorf("no grantee found")
	}

	for _, remove := range keysToRemove {
		for i, grantee := range grantees {
			if grantee.Equal(remove) {
				grantees[i] = grantees[len(grantees)-1]
				grantees = grantees[:len(grantees)-1]
			}
		}
	}
	g.grantees = g.Serialize(grantees)
	return nil
}

// TODO: retrun GrantList IF instead of GranteeListStructMock
func NewGranteeList(ls file.LoadSaver, putter storer.PutterSession, reference swarm.Address) GranteeList {
	var (
		data []byte
		err  error
	)
	if swarm.ZeroAddress.Equal(reference) || swarm.EmptyAddress.Equal(reference) {
		data = []byte{}
	} else {
		data, err = ls.Load(context.Background(), reference.Bytes())
	}
	if err != nil {
		return nil
	}

	return &GranteeListStruct{
		grantees: data,
		loadSave: ls,
		putter:   putter,
	}
}
