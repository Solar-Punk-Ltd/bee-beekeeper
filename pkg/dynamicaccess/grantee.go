package dynamicaccess

type Grantee interface {
	Revoke(topic string) error
	Publish(topic string) error
	// RevokeList(topic string, removeList []string, addList []string) (string, error)
	RevokeGrantees(topic string, removeList []string) (string, error)
	AddGrantees(addList []string) ([]string, error)
}

type defaultGrantee struct {
	topic string;
	grantees []string;
}

func (g *defaultGrantee) Revoke(topic string) error {
	return nil
}

func (g *defaultGrantee) RevokeList(topic string, removeList []string, addList []string) (string, error) {
	return "", nil
}

func (g *defaultGrantee) Publish(topic string) error {
	return nil
}



func (g *defaultGrantee) AddGrantees(addList []string) ([]string, error) {
	g.grantees = append(g.grantees, addList...)
	return g.grantees, nil
}

func NewGrantee() Grantee {
	return &defaultGrantee{}
}
