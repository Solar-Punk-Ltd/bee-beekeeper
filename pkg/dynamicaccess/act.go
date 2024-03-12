package dynamicaccess

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
)

type Act interface {
	Add(lookupKey []byte, encryptedAccessKey []byte) *defaultAct
	Get(lookupKey []byte) string
	Load(data string) error
	Store() (string, error)
}

var _ Act = (*defaultAct)(nil)

type defaultAct struct {
	container map[string]string
}

func (act *defaultAct) Add(lookupKey []byte, encryptedAccessKey []byte) *defaultAct {
	act.container[hex.EncodeToString(lookupKey)] = hex.EncodeToString(encryptedAccessKey)
	return act
}

func (act *defaultAct) Get(lookupKey []byte) string {
	if key, ok := act.container[hex.EncodeToString(lookupKey)]; ok {
		return key
	}
	return ""
}

func (act *defaultAct) Load(data string) error {
	b := new(bytes.Buffer)
	b.WriteString(data)
	d := gob.NewDecoder(b)

	// Decoding the serialized data
	err := d.Decode(&act.container)
	if err != nil {
		return err
	}
	return nil
}

func (act *defaultAct) Store() (string, error) {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)

	// Encoding the map
	err := e.Encode(act.container)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func NewDefaultAct() Act {
	return &defaultAct{
		container: make(map[string]string),
	}
}
