package hub

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/orktes/homeautomation/adapter"
	"github.com/orktes/homeautomation/config"
	"github.com/orktes/homeautomation/util"
)

type Hub struct {
	adapters       map[string]adapter.Adapter
	lights         map[string]*Light
	updateChannels []chan adapter.Update

	jsRuntime *jsRuntime

	sync.Mutex
}

func New() *Hub {
	hub := &Hub{
		adapters:  map[string]adapter.Adapter{},
		lights:    map[string]*Light{},
		jsRuntime: newJSRuntime(),
	}

	hub.jsRuntime.Set("hub", &objectLikeValueContainer{hub})
	hub.jsRuntime.Set("on", true)
	hub.jsRuntime.Set("off", true)

	return hub
}

func (hub *Hub) RunScript(src string) (interface{}, error) {
	v, err := hub.jsRuntime.RunString(fmt.Sprintf("with(hub) {%s}", src))
	if err != nil {
		return nil, err
	}

	val := v.Export()

	if olcv, ok := val.(*objectLikeValueContainer); ok {
		val = olcv.ValueContainer
	}

	return val, nil
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

	parts := util.SplitID(id)
	root := parts[0]

	var c adapter.ValueContainer
	var ok bool
	switch root {
	case "lights":
		// TODO only create once
		c, ok = lightsValueContainer(hub.lights), true
	default:
		c, ok = hub.adapters[root]
	}

	if !ok {
		return nil, errors.New("Not found")
	}

	if len(parts) == 1 {
		return c, nil
	}

	return c.Get(strings.Join(parts[1:], "."))
}

func (hub *Hub) Set(id string, val interface{}) error {
	parts := util.SplitID(id)
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

	nodes := map[string]interface{}{
		// TODO check if lightsValueContainer needs hub mutex
		// TODO only create lightsValueContainer one
		"lights": lightsValueContainer(hub.lights),
	}

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

func (hub *Hub) CreateLight(lightConfig config.Light) *Light {
	hub.Lock()
	defer hub.Unlock()

	id := fmt.Sprintf("%d", len(hub.lights)+1)

	light := &Light{id: id, hub: hub, conf: lightConfig}
	light.init()

	hub.lights[id] = light

	return light
}

func (hub *Hub) GetLights() []*Light {
	hub.Lock()
	defer hub.Unlock()

	lights := make([]*Light, 0, len(hub.lights))

	for _, light := range hub.lights {
		lights = append(lights, light)
	}

	return lights
}

func (hub *Hub) CreateTrigger(trigger config.Trigger) {
	go func() {
		key := util.ConvertJSIDToDotID(trigger.Key)
		ch := hub.UpdateChannel()

		var stopInterval func()

		action := func() {
			time.Sleep(time.Duration(trigger.Delay) * time.Millisecond)
			_, err := hub.RunScript(trigger.Action)
			if err != nil {
				log.Printf("Trigger:%s action returned an error: %s\n", trigger.Name, err.Error())
			}
		}

		for u := range ch {
			for _, kvpu := range u.Updates {
				if kvpu.Key == key {
					cond, err := hub.RunScript(trigger.Condition)
					if err != nil {
						log.Printf("Error occured while executing trigger:%s condition: %s", trigger.Name, err.Error())
						break
					}

					condBoolValue, ok := cond.(bool)
					if !ok {
						log.Printf("Trigger:%s end condition returned a non bool result %+v\n", trigger.Name, cond)
						break
					}

					if condBoolValue {
						if trigger.EndCondition == "" {
							action()
						} else {
							if stopInterval != nil {
								stopInterval()
							}
							if trigger.Interval == 0 {
								log.Printf("Trigger:%s doesnt have an interval defined", trigger.Name)
								break
							}
							action()
							stopInterval = util.Interval(action, time.Duration(trigger.Interval)*time.Millisecond)
						}
					}

					if trigger.EndCondition != "" {
						cond, err := hub.RunScript(trigger.EndCondition)
						if err != nil {
							log.Printf("Error occured while executing trigger:%s condition: %s", trigger.Name, err.Error())
							break
						}

						condBoolValue, ok = cond.(bool)
						if !ok {
							log.Printf("Trigger:%s end condition returned a non bool result %+v\n", trigger.Name, cond)
							break
						}

						if condBoolValue && stopInterval != nil {
							stopInterval()
							stopInterval = nil
						}
					}

					break
				}

			}
		}
	}()
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
