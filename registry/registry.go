package registry

import (
	"errors"

	"github.com/orktes/homeautomation/adapter"
	"github.com/orktes/homeautomation/config"
	"github.com/orktes/homeautomation/frontend"
	"github.com/orktes/homeautomation/hub"
)

var (
	AdapterNotFoundError  = errors.New("Adapter not found from registry")
	FrontendNotFoundError = errors.New("Frontend not found from registry")
)

var adapters = map[string]func(id string, conf map[string]interface{}, hub *hub.Hub) (adapter.Adapter, error){}
var frontends = map[string]func(id string, conf map[string]interface{}, hub *hub.Hub) (frontend.Frontend, error){}

func RegisterAdapter(typ string, factory func(id string, conf map[string]interface{}, hub *hub.Hub) (adapter.Adapter, error)) {
	adapters[typ] = factory
}

func RegisterFrontend(typ string, factory func(id string, conf map[string]interface{}, hub *hub.Hub) (frontend.Frontend, error)) {
	frontends[typ] = factory
}

func CreateAdapter(config config.Adapter, hub *hub.Hub) (adapter.Adapter, error) {
	factory, ok := adapters[config.Type]
	if !ok {
		return nil, AdapterNotFoundError
	}

	return factory(config.ID, config.Config, hub)
}

func CreateFrontend(config config.Frontend, hub *hub.Hub) (frontend.Frontend, error) {
	factory, ok := frontends[config.Type]
	if !ok {
		return nil, FrontendNotFoundError
	}

	return factory(config.ID, config.Config, hub)
}
