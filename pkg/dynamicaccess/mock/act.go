package mock

type ActMock struct {
	AddFunc   func(lookupKey []byte, encryptedAccessKey []byte) *ActMock
	GetFunc   func(lookupKey []byte) string
	LoadFunc  func(data string) error
	StoreFunc func() (string, error)
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

func (act *ActMock) Load(data string) error {
	if act.LoadFunc == nil {
		return nil
	}
	return act.LoadFunc(data)
}

func (act *ActMock) Store() (string, error) {
	if act.StoreFunc == nil {
		return "", nil
	}
	return act.StoreFunc()
}

func NewActMock() *ActMock {
	return &ActMock{}
}
