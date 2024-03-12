package dynamicaccess

import (
	"encoding/hex"

	"github.com/ethersphere/bee/pkg/manifest"
	"github.com/ethersphere/bee/pkg/swarm"
)

type Act interface {
	Add(lookupKey []byte, encryptedAccessKey []byte) *defaultAct
	Get(lookupKey []byte) string // TODO: return []byte
	Load(lookupKey []byte) manifest.Entry
	Store(me manifest.Entry)
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

// to manifestEntry
func (act *defaultAct) Load(lookupKey []byte) manifest.Entry {
	return manifest.NewEntry(swarm.NewAddress(lookupKey), act.container)
}

// from manifestEntry
func (act *defaultAct) Store(me manifest.Entry) {
	if act.container == nil {
		act.container = make(map[string]string)
	}
	act.container = me.Metadata()
}

func NewDefaultAct() Act {
	return &defaultAct{
		container: make(map[string]string),
	}
}
