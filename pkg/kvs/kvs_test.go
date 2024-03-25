// Copyright 2024 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kvs_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"testing"

	"github.com/ethersphere/bee/pkg/file"
	"github.com/ethersphere/bee/pkg/file/loadsave"
	"github.com/ethersphere/bee/pkg/file/pipeline"
	"github.com/ethersphere/bee/pkg/file/pipeline/builder"
	"github.com/ethersphere/bee/pkg/file/redundancy"
	"github.com/ethersphere/bee/pkg/kvs"
	"github.com/ethersphere/bee/pkg/storage"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/ethersphere/bee/pkg/swarm"
)

var mockStorer = mockstorer.New()

func requestPipelineFactory(ctx context.Context, s storage.Putter, encrypt bool, rLevel redundancy.Level) func() pipeline.Interface {
	return func() pipeline.Interface {
		return builder.NewPipelineBuilder(ctx, s, encrypt, rLevel)
	}
}

func createLs() file.LoadSaver {
	return loadsave.New(mockStorer.ChunkStore(), mockStorer.Cache(), requestPipelineFactory(context.Background(), mockStorer.Cache(), false, redundancy.NONE))
}

func TestActAddLookup(t *testing.T) {
	ls := createLs()

	s := kvs.New(ls, swarm.ZeroAddress)

	lookupKey := swarm.RandAddress(t).Bytes()
	encryptedAccesskey := swarm.RandAddress(t).Bytes()

	err := s.Put(lookupKey, encryptedAccesskey)
	if err != nil {
		t.Errorf("Add() should not return an error: %v", err)
	}

	key, err := s.Get(lookupKey)
	if err != nil {
		t.Errorf("Lookup() should not return an error: %v", err)
	}

	if !bytes.Equal(key, encryptedAccesskey) {
		t.Errorf("Get() value is not the expected %s != %s", hex.EncodeToString(key), hex.EncodeToString(encryptedAccesskey))
	}

}

func TestActAddLookupWithNew(t *testing.T) {
	ls := createLs()
	s1 := kvs.New(ls, swarm.ZeroAddress)
	lookupKey := swarm.RandAddress(t).Bytes()
	encryptedAccesskey := swarm.RandAddress(t).Bytes()

	err := s1.Put(lookupKey, encryptedAccesskey)
	if err != nil {
		t.Fatalf("Add() should not return an error: %v", err)
	}

	s2 := kvs.New(ls, swarm.ZeroAddress)
	key, err := s2.Get(lookupKey)
	if err != nil {
		t.Fatalf("Lookup() should not return an error: %v", err)
	}

	if !bytes.Equal(key, encryptedAccesskey) {
		t.Errorf("Get() value is not the expected %s != %s", hex.EncodeToString(key), hex.EncodeToString(encryptedAccesskey))
	}

}

/*
func TestActStoreLoad(t *testing.T) {

	act := dynamicaccess.NewInMemoryAct()
	lookupKey := swarm.RandAddress(t).Bytes()
	encryptedAccesskey := swarm.RandAddress(t).Bytes()
	err := act.Add(lookupKey, encryptedAccesskey)
	if err != nil {
		t.Error("Add() should not return an error")
	}

	swarm_ref, err := act.Store()
	if err != nil {
		t.Error("Store() should not return an error")
	}

	actualAct := dynamicaccess.NewInMemoryAct()
	actualAct.Load(swarm_ref)
	actualEak, _ := actualAct.Lookup(lookupKey)
	if !bytes.Equal(actualEak, encryptedAccesskey) {
		t.Errorf("actualAct.Load() value is not the expected %s != %s", hex.EncodeToString(actualEak), hex.EncodeToString(encryptedAccesskey))
	}
}
*/
