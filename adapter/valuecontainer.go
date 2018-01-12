package adapter

type ValueContainer interface {
	Get(id string) (interface{}, error)
	Set(id string, val interface{}) error
	GetAll() (map[string]interface{}, error)
}

type LeafValueContainer struct{ Value interface{} }

func (lvc LeafValueContainer) Get(id string) (interface{}, error) {
	panic("This is a leaf value")
}
func (lvc LeafValueContainer) Set(id string, val interface{}) error {
	panic("This is a leaf value")
}
