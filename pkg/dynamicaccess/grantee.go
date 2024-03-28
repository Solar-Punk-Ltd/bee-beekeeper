package dynamicaccess

import (
	"crypto/ecdsa"
	"crypto/elliptic"

	"github.com/ethersphere/bee/pkg/swarm"
)

type GranteeList interface {
	Add(topic string, addList []*ecdsa.PublicKey) error
	Remove(topic string, removeList []*ecdsa.PublicKey) error
	Get(topic string) []*ecdsa.PublicKey
	Save() (swarm.Address, error)
}

type GranteeListStruct struct {
	grantees map[string][]*ecdsa.PublicKey
}

func (g *GranteeListStruct) Get(topic string) []*ecdsa.PublicKey {
	grantees := g.grantees[topic]
	keys := make([]*ecdsa.PublicKey, len(grantees))
	copy(keys, grantees)
	return keys
}

func (g *GranteeListStruct) Serialize(addList []*ecdsa.PublicKey) []byte {
	b := make([]byte, 0, len(addList))
	// d := byte(',')
	for _, key := range addList {
		b = append(b, g.SerializePublicKey(key)...)
	}
	// grantees = b
	return b
}

func (g *GranteeListStruct) DeSerialize(data []byte) []*ecdsa.PublicKey {
	p := make([]*ecdsa.PublicKey, 0, len(data)/65)
	// d := byte(',')
	for i := 0; i < len(data); i += 65 {
		p = append(p, g.DeSerializeBytes(data[i:i+65]))
	}
	// grantees = b
	return p
}

func (g *GranteeListStruct) SerializePublicKey(pub *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(pub.Curve, pub.X, pub.Y)
}

func (g *GranteeListStruct) DeSerializeBytes(data []byte) *ecdsa.PublicKey {
	curve := elliptic.P256()
	// TODO: use not depreceted ecdh stuff
	// pub, err := ecdh.P256().NewPublicKey(data)
	// if err != nil {
	// 	return nil
	// }
	x, y := elliptic.Unmarshal(curve, data)
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}
	// return pub
}

func (g *GranteeListStruct) Save() (swarm.Address, error) {
	return swarm.EmptyAddress, nil
}

// == 0x11132454/..soc...soc...
// file upload -> address -> key/ lookupkey
//
//	roothash      / pubkey(lookupkey) /   accesskey / mtdt (empty map) (application/json)
//
// manifest: swarmAddress, path :(hash/address) / address/metadata (map[string]string)
// manifes.RootPath == / -> soc hash
func (g *GranteeListStruct) Add(topic string, addList []*ecdsa.PublicKey) error {
	// addList -> serialize to []byte -> string -> manifestkey
	// b, err := hex.DecodeString(topic)
	// if err != nil {
	// 	return err
	// }
	// g.grantees.Put(context.Background(), b, addList)
	g.grantees[topic] = append(g.grantees[topic], addList...)
	return nil
}

func (g *GranteeListStruct) Remove(topic string, removeList []*ecdsa.PublicKey) error {
	for _, remove := range removeList {
		for i, grantee := range g.grantees[topic] {
			if *grantee == *remove {
				g.grantees[topic][i] = g.grantees[topic][len(g.grantees[topic])-1]
				g.grantees[topic] = g.grantees[topic][:len(g.grantees[topic])-1]
			}
		}
	}

	return nil
}

func NewGrantee() *GranteeListStruct {
	return &GranteeListStruct{grantees: make(map[string][]*ecdsa.PublicKey)}
}

func (g *GranteeListStruct) Store() (swarm.Address, error) {
	return swarm.EmptyAddress, nil
}
