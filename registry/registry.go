package registry

import (
	"errors"

	"github.com/orktes/homeautomation/adapter"
	"github.com/orktes/homeautomation/config"
	"github.com/orktes/homeautomation/hub"
)

var (
	AdapterNotFoundError = errors.New("Adapter not found from registry")
)

var factories = map[string]func(id string, conf map[string]interface{}, hub *hub.Hub) (adapter.Adapter, error){}

func Register(typ string, factory func(id string, conf map[string]interface{}, hub *hub.Hub) (adapter.Adapter, error)) {
	factories[typ] = factory
}

func Create(config config.Adapter, hub *hub.Hub) (adapter.Adapter, error) {
	factory, ok := factories[config.Type]
	if !ok {
		return nil, AdapterNotFoundError
	}

	return factory(config.ID, config.Config, hub)
}
