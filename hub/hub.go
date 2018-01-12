package hub

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/orktes/homeautomation/adapter"
	"github.com/orktes/homeautomation/config"
)

type Hub struct {
	adapters       map[string]adapter.Adapter
	updateChannels []chan adapter.Update

	sync.Mutex
}

func New() *Hub {
	hub := &Hub{adapters: map[string]adapter.Adapter{}}
	return hub
}

func (hub *Hub) UpdateChannel() <-chan adapter.Update {
	hub.Lock()
	defer hub.Unlock()

	ch := make(chan adapter.Update)
	hub.updateChannels = append(hub.updateChannels, ch)
	return ch
}

func (hub *Hub) Get(id string) (interface{}, error) {
	hub.Lock()
	defer hub.Unlock()

	parts := strings.Split(id, ".")
	c, ok := hub.adapters[parts[0]]
	if !ok {
		return nil, errors.New("Not found")
	}

	if len(parts) == 1 {
		return c, nil
	}

	return c.Get(strings.Join(parts[1:], "."))
}

func (hub *Hub) Set(id string, val interface{}) error {
	parts := strings.Split(id, ".")
	c, err := hub.Get(parts[0])
	if err != nil {
		return err
	}

	vc, ok := c.(adapter.ValueContainer)
	if !ok {
		return fmt.Errorf("Can't be set as %s (%s) is not a value container", parts[0], reflect.TypeOf(vc))
	}

	return vc.Set(strings.Join(parts[1:], "."), val)
}

func (hub *Hub) GetAll() (map[string]interface{}, error) {
	hub.Lock()
	defer hub.Unlock()

	nodes := map[string]interface{}{}

	for key, ad := range hub.adapters {
		nodes[key] = ad
	}

	return nodes, nil
}

func (hub *Hub) AddAdapter(id string, ad adapter.Adapter) {
	hub.Lock()
	defer hub.Unlock()

	hub.adapters[id] = ad
	go func() {
		ch := ad.UpdateChannel()
		for u := range ch {
			hub.sendUpdate(u)
		}
	}()
}

func (hub *Hub) CreateTrigger(trigger config.Trigger) {
	go func() {
		ch := hub.UpdateChannel()
		for u := range ch {
			for _, kvpu := range u.Updates {
				// TODO support intervalled triggers
				if kvpu.Key == trigger.Key && kvpu.Value == trigger.Value {
					go func(trigger config.Trigger) {
						if trigger.Delay > 0 {
							time.Sleep(time.Duration(trigger.Delay) * time.Millisecond)
						}
						err := hub.Set(trigger.Target, trigger.TargetValue)
						if err != nil {
							fmt.Printf("Error occured when setting %s = %+v : %s\n", trigger.Target, trigger.TargetValue, err.Error())
						}
					}(trigger)
					break
				}
			}
		}
	}()
}

func (hub *Hub) GetSettingStore() *SettingsStore {
	return &SettingsStore{}
}

func (hub *Hub) Close() error {
	// TODO close adapters and frontends
	return nil
}

func (hub *Hub) sendUpdate(update adapter.Update) {
	hub.Lock()
	defer hub.Unlock()

	for _, ch := range hub.updateChannels {
		ch <- update
	}
}
