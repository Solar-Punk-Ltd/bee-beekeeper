package dynamicaccess

import "testing"

func TestXxx(t *testing.T) {
	//key encryption.Key, padding int, initCtr uint32, hashFunc func() hash.Hash
	al := NewAccessLogic(nil, 0, 0, nil)
	if al == nil {
		t.Errorf("Error creating access logic")
	}
	newObj,err:=al.Get("rootKey", "encryped_ref", "publisher", "tag")
	if err!=nil {
	println(newObj)
	}
}
