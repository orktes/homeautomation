package adapter

import (
	"errors"
	"strings"
)

var (
	NoSuchAdapterError = errors.New("no such adapter")
)

type MultiAdapter struct {
	id       string
	adapters map[string]Adapter

	Updater
}

func NewMultiAdapter(id string, adapters ...Adapter) Adapter {
	adapterMap := map[string]Adapter{}
	ma := &MultiAdapter{id: id, adapters: adapterMap}

	for _, adapter := range adapters {
		adapterMap[adapter.ID()] = adapter
		go func(adapter Adapter) {
			ch := adapter.UpdateChannel()
			for u := range ch {
				proxyU := Update{
					ValueContainer: u.ValueContainer,
					Updates:        make([]ValueUpdate, 0, len(u.Updates)),
				}

				for _, kvu := range u.Updates {
					proxyU.Updates = append(proxyU.Updates, ValueUpdate{
						Key:   id + "/" + kvu.Key,
						Value: kvu.Value,
					})
				}

				ma.Updater.SendUpdate(proxyU)

			}
		}(adapter)
	}

	return ma
}

func (ma *MultiAdapter) Get(id string) (interface{}, error) {
	if id == "" {
		return ma, nil
	}

	parts := strings.Split(id, "/")
	adapter, ok := ma.adapters[parts[0]]
	if !ok {
		return nil, NoSuchAdapterError
	}

	return adapter.Get(strings.Join(parts[1:], "/"))
}

func (ma *MultiAdapter) Set(id string, val interface{}) error {
	if id == "" {
		return nil
	}

	parts := strings.Split(id, "/")
	adapter, ok := ma.adapters[parts[0]]
	if !ok {
		return NoSuchAdapterError
	}

	return adapter.Set(strings.Join(parts[1:], "/"), val)
}

func (ma *MultiAdapter) GetAll() (map[string]interface{}, error) {
	vals := map[string]interface{}{}

	for key, adapter := range ma.adapters {
		vals[key] = adapter
	}

	return vals, nil
}

func (ma *MultiAdapter) ID() string {
	return ma.id
}

func (ma *MultiAdapter) UpdateChannel() <-chan Update {
	return ma.Updater.UpdateChannel()
}

func (ma *MultiAdapter) Close() error {
	// TODO process errors for Close
	for _, adapter := range ma.adapters {
		adapter.Close()
	}

	return nil
}
