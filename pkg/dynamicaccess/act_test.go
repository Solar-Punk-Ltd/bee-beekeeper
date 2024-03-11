package dynamicaccess_test

import (
	"encoding/hex"
	"testing"

	"github.com/ethersphere/bee/pkg/dynamicaccess"
	"github.com/ethersphere/bee/pkg/swarm"
)

func TestAddGet(t *testing.T) {
	act := dynamicaccess.NewDefaultAct()
	lookupKey := swarm.RandAddress(t).Bytes()
	encryptedAccesskey := swarm.RandAddress(t).Bytes()
	act2 := act.Add(lookupKey, encryptedAccesskey)
	if act2 == nil {
		t.Error("Add() should return an act")
	}

	key := act.Get(lookupKey)
	if key != hex.EncodeToString(encryptedAccesskey) {
		t.Errorf("Get() value is not the expected %s != %s", key, encryptedAccesskey)
	}
}
