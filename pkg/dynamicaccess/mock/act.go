package mock

import "github.com/ethersphere/bee/pkg/manifest"

type ActMock struct {
	AddFunc   func(lookupKey []byte, encryptedAccessKey []byte) *ActMock
	GetFunc   func(lookupKey []byte) string // TODO: return []byte
	LoadFunc  func(lookupKey []byte) manifest.Entry
	StoreFunc func(me manifest.Entry)
}

func (act *ActMock) Add(lookupKey []byte, encryptedAccessKey []byte) *ActMock {
	if act.AddFunc == nil {
		return act
	}
	return act.AddFunc(lookupKey, encryptedAccessKey)
}

func (act *ActMock) Get(rootHash string, lookupKey []byte) string {
	if act.GetFunc == nil {
		return ""
	}
	return act.GetFunc(lookupKey)
}

func (act *ActMock) Load(lookupKey []byte) manifest.Entry {
	if act.LoadFunc == nil {
		return nil
	}
	return act.LoadFunc(lookupKey)
}

func (act *ActMock) Store(me manifest.Entry) {
	if act.StoreFunc == nil {
		return
	}
	act.StoreFunc(me)
}

func NewActMock() *ActMock {
	return &ActMock{}
}
