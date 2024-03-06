package dynamicaccess_test

import (
	"testing"

	"github.com/ethersphere/bee/pkg/dynamicaccess"
	"github.com/ethersphere/bee/pkg/swarm"
)

func TestAdd(t *testing.T) {
	a := dynamicaccess.NewDefaultAct()
	addr := swarm.RandAddress(t)
	expval := "0xff"
	err := a.Add(addr, expval)
	if err != nil {
		t.Error("Add() should not return an error")
	}

	val, err := a.Get(addr)
	if err != nil {
		t.Error("Get() should not return an error")
	}
	if val != expval {
		t.Errorf("Get() value is not the expected %s != %s", val, expval)
	}
}

func TestGet(t *testing.T) {
	a := dynamicaccess.NewDefaultAct()
	val, err := a.Get(swarm.RandAddress(t))
	if err != nil {
		t.Error("Get() should not return an error")
	}
	if val != "" {
		t.Error("Get() should not return a value")
	}
}
