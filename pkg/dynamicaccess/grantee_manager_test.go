package dynamicaccess_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/ethersphere/bee/pkg/dynamicaccess"
	"github.com/ethersphere/bee/pkg/swarm"
)

func setupAccessLogic(privateKey *ecdsa.PrivateKey) dynamicaccess.ActLogic {
<<<<<<< HEAD
	act := dynamicaccess.NewInMemoryAct()
	si := dynamicaccess.NewDefaultSession(privateKey)
	al := dynamicaccess.NewLogic(si, act)
=======
	// privateKey, err := crypto.GenerateSecp256k1Key()
	// if err != nil {
	// 	errors.New("error creating private key")
	// }
	si := dynamicaccess.NewDefaultSession(privateKey)
	al := dynamicaccess.NewLogic(si)
>>>>>>> origin/act

	return al
}

func TestAdd(t *testing.T) {
	m := dynamicaccess.NewGranteeManager(setupAccessLogic(getPrivateKey()))
	pub, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	id1, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	id2, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	err := m.Add("topic", []*ecdsa.PublicKey{&id1.PublicKey})
	if err != nil {
		t.Errorf("Add() returned an error")
	}
	err = m.Add("topic", []*ecdsa.PublicKey{&id2.PublicKey})
	if err != nil {
		t.Errorf("Add() returned an error")
	}
<<<<<<< HEAD
	m.Publish(swarm.EmptyAddress, &pub.PublicKey, "topic")
=======
	m.Publish(act, &pub.PublicKey, "topic")
>>>>>>> origin/act
	fmt.Println("")
}
