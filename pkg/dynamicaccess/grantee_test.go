package dynamicaccess_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"reflect"
	"testing"
)

func TestGranteeAddGrantees(t *testing.T) {
	grantee := NewGrantee()

	key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	key2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	addList := []ecdsa.PublicKey{key1.PublicKey, key2.PublicKey}
	grantees, err := grantee.AddGrantees(addList)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !reflect.DeepEqual(grantees, addList) {
		t.Errorf("Expected grantees %v, got %v", addList, grantees)
	}
}

func TestRemoveGrantees(t *testing.T) {
	grantee := NewGrantee()

	key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	key2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	addList := []ecdsa.PublicKey{key1.PublicKey, key2.PublicKey}
	grantee.AddGrantees(addList)

	removeList := []ecdsa.PublicKey{key1.PublicKey}
	grantees, err := grantee.RemoveGrantees(removeList)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedGrantees := []ecdsa.PublicKey{key2.PublicKey}
	if !reflect.DeepEqual(grantees, expectedGrantees) {
		t.Errorf("Expected grantees %v, got %v", expectedGrantees, grantees)
	}
}

func TestGetGrantees(t *testing.T) {
	grantee := NewGrantee()

	key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	key2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	addList := []ecdsa.PublicKey{key1.PublicKey, key2.PublicKey}
	grantee.AddGrantees(addList)

	grantees := grantee.GetGrantees()

	if !reflect.DeepEqual(grantees, addList) {
		t.Errorf("Expected grantees %v, got %v", addList, grantees)
	}
}
