package dynamicaccess

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"io"
	"testing"

	"github.com/ethersphere/bee/pkg/crypto"
)

func TestSharedSecret(t *testing.T) {
	_, err := NewDiffieHellman(nil).SharedSecret(&ecdsa.PublicKey{}, "", nil)
	if err != nil {
		t.Errorf("Error generating shared secret: %v", err)
	}
}

func TestECDHCorrect(t *testing.T) {
	t.Parallel()

	key1, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}
	dh1 := NewDiffieHellman(key1)

	key2, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}
	dh2 := NewDiffieHellman(key2)

	moment := make([]byte, 1)
	if _, err := io.ReadFull(rand.Reader, moment); err != nil {
		t.Fatal(err)
	}

	shared1, err := dh1.SharedSecret(&key2.PublicKey, "", moment)
	if err != nil {
		t.Fatal(err)
	}
	shared2, err := dh2.SharedSecret(&key1.PublicKey, "", moment)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(shared1, shared2) {
		t.Fatal("shared secrets do not match")
	}
}
