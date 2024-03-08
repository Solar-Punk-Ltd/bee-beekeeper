package dynamicaccess_test

import (
	"encoding/hex"
	"testing"

	"github.com/ethersphere/bee/pkg/dynamicaccess"
	"github.com/ethersphere/bee/pkg/swarm"
)

func TestAddGet(t *testing.T) {
	act := dynamicaccess.NewDefaultAct()
	rootHashString := swarm.RandAddress(t).String()
	lookupKey := swarm.RandAddress(t).Bytes()
	encryptedAccesskey := swarm.RandAddress(t).Bytes()
	_, err := act.Add(rootHashString, lookupKey, encryptedAccesskey)
	if err != nil {
		t.Error("Add() should not return an error")
	}

	key, err := act.Get(rootHashString, lookupKey)
	if err != nil {
		t.Error("Get() should not return an error")
	}
	if key != hex.EncodeToString(encryptedAccesskey) {
		t.Errorf("Get() value is not the expected %s != %s", key, encryptedAccesskey)
	}
}
