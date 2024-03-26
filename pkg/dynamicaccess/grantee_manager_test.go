package dynamicaccess_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/ethersphere/bee/pkg/dynamicaccess"
	kvsmock "github.com/ethersphere/bee/pkg/kvs/mock"
	"github.com/ethersphere/bee/pkg/swarm"
)

func setupAccessLogic(privateKey *ecdsa.PrivateKey) dynamicaccess.ActLogic {
	si := dynamicaccess.NewDefaultSession(privateKey)
	al := dynamicaccess.NewLogic(si)

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
	s := kvsmock.New(createLs(), swarm.ZeroAddress)
	m.Publish(s, &pub.PublicKey, "topic")
	fmt.Println("")
}
