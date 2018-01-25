package mqtt

import "github.com/orktes/homeautomation/bridge/adapter"

type mockAdapter struct {
	id   string
	vals map[string]interface{}
	adapter.Updater
}

func (ma *mockAdapter) Get(id string) (interface{}, error) {
	if id == "" {
		return ma, nil
	}
	val := ma.vals[id]
	return val, nil
}

func (ma *mockAdapter) Set(id string, val interface{}) error {
	ma.vals[id] = val
	go ma.Updater.SendUpdate(adapter.Update{
		ValueContainer: ma,
		Updates: []adapter.ValueUpdate{
			adapter.ValueUpdate{
				Key:   ma.ID() + "/" + id,
				Value: val,
			},
		},
	})
	return nil
}

func (ma *mockAdapter) GetAll() (map[string]interface{}, error) {
	return ma.vals, nil
}

func (ma *mockAdapter) ID() string {
	return ma.id
}

func (ma *mockAdapter) UpdateChannel() <-chan adapter.Update {
	return ma.Updater.UpdateChannel()
}

func (ma *mockAdapter) Close() error {
	return nil
}
