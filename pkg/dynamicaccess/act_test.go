package dynamicaccess_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/ethersphere/bee/pkg/dynamicaccess"
	"github.com/ethersphere/bee/pkg/file/loadsave"
	"github.com/ethersphere/bee/pkg/file/pipeline"
	"github.com/ethersphere/bee/pkg/file/pipeline/builder"
	"github.com/ethersphere/bee/pkg/manifest"
	"github.com/ethersphere/bee/pkg/storage"
	"github.com/ethersphere/bee/pkg/storage/inmemchunkstore"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/ethersphere/bee/pkg/swarm"
)

func pipelineFn(s storage.Putter) func() pipeline.Interface {
	return func() pipeline.Interface {
		return builder.NewPipelineBuilder(context.Background(), s, false, 0)
	}
}

func TestAdd(t *testing.T) {
	storer := mockstorer.New()
	ls := loadsave.New(storer.ChunkStore(), storer.Cache(), pipelineFn(storer.Cache()))
	m, _ := manifest.NewDefaultManifest(ls, true)
	rootHash, err := m.Store(context.Background())
	if err != nil {
		t.Error("Store() should not return an error")
	}

	a := dynamicaccess.NewDefaultAct(m)
	// rootHash := swarm.RandAddress(t)
	// lookupkKey0 := swarm.RandAddress(t)
	lookupkKey0 := make([]byte, 64)
	_, err = rand.Read(lookupkKey0)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(lookupkKey0)
	expval := "0xff"
	_, err = a.Add(context.Background(), rootHash.String(), lookupkKey0, expval) // Assign the value to err
	if err != nil {
		t.Error("Add() should not return an error")
	}

	val, err := a.Get(context.Background(), rootHash.Bytes(), lookupkKey0)
	if err != nil {
		t.Error("Get() should not return an error")
	}
	if val != expval {
		t.Errorf("Get() value is not the expected %s != %s", val, expval)
	}
}

func TestGet(t *testing.T) {
	store := inmemchunkstore.New()
	ls := loadsave.New(store, store, pipelineFn(store))
	m, _ := manifest.NewDefaultManifest(ls, true)
	a := dynamicaccess.NewDefaultAct(m)
	rootHash := swarm.RandAddress(t)
	lookupkKey0 := swarm.RandAddress(t)
	val, err := a.Get(context.Background(), rootHash.Bytes(), lookupkKey0.Bytes())
	if err != nil {
		t.Error("Get() should not return an error")
	}
	if val != "" {
		t.Error("Get() should not return a value")
	}
}
