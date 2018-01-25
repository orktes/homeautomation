package adapter

type ValueContainer interface {
	Get(id string) (interface{}, error)
	Set(id string, val interface{}) error
	GetAll() (map[string]interface{}, error)
}
