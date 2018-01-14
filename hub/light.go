package hub

import (
	"errors"
	"reflect"
	"sync"

	"github.com/orktes/homeautomation/adapter"
	"github.com/orktes/homeautomation/config"
)

var (
	InvalidReturnTypeError = errors.New("invalid return type")
)

// LightUpdate represent a single light update
type LightUpdate struct {
	Property string
	Value    interface{}
}

// Light represent a hub light
type Light struct {
	hub  *Hub
	id   string
	conf config.Light

	updateChannels []chan LightUpdate

	sync.Mutex
}

func (l *Light) init() {
	go l.listenToUpdates()
}

func (l *Light) ID() string {
	return l.id
}

func (l *Light) GetName() string {
	return l.conf.Name
}

func (l *Light) GetBrightness() (int, error) {
	if l.conf.Read.Brightness == "" {
		return 0, nil
	}

	val, err := l.hub.Get(l.conf.Read.Brightness)
	if err != nil {
		return 0, nil
	}

	if val == nil {
		return 0, nil
	}

	i, ok := val.(int)
	if !ok {
		return 0, InvalidReturnTypeError
	}

	return i, nil
}

func (l *Light) SetBrightness(val int) error {
	if l.conf.Write.Brightness == "" {
		return nil
	}

	return l.hub.Set(l.conf.Write.Brightness, val)
}

func (l *Light) GetON() (bool, error) {
	if l.conf.Read.On == "" {
		return false, nil
	}

	val, err := l.hub.Get(l.conf.Read.On)
	if err != nil {
		return false, err
	}

	i, ok := val.(bool)
	if !ok {
		return false, InvalidReturnTypeError
	}

	return i, nil
}

func (l *Light) SetOn(val bool) error {
	if l.conf.Write.On == "" {
		return nil
	}

	return l.hub.Set(l.conf.Write.On, val)
}

func (l *Light) readStructProps(in interface{}) []string {
	keys := []string{}
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		val := v.Field(i).Interface()
		strVal, _ := val.(string)
		if strVal != "" {
			keys = append(keys, t.Field(i).Name)
		}
	}

	return keys
}

func (l *Light) emitUpdate(updatedKeyIndex int, val interface{}) {
	l.Lock()
	defer l.Unlock()

	key := l.ReadableProperties()[updatedKeyIndex]
	u := LightUpdate{Property: key, Value: val}

	for _, ch := range l.updateChannels {
		ch <- u
	}
}

func (l *Light) getConfiguredSourceKeys() []string {
	keys := []string{}
	v := reflect.ValueOf(l.conf.Read)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		val := v.Field(i).Interface()
		strVal, _ := val.(string)
		if strVal != "" {
			keys = append(keys, strVal)
		}
	}

	return keys
}

func (l *Light) listenToUpdates() {
	ch := l.hub.UpdateChannel()
	keys := l.getConfiguredSourceKeys()

	for ue := range ch {
	updateLoop:
		for _, u := range ue.Updates {
			for i, k := range keys {
				if u.Key == k {
					l.emitUpdate(i, u.Value)
					break updateLoop
				}
			}
		}
	}
}

func (l *Light) ReadableProperties() []string {
	return l.readStructProps(l.conf.Read)
}

func (l *Light) WritableProperties() []string {
	return l.readStructProps(l.conf.Write)
}

func (l *Light) UpdateChannel() <-chan LightUpdate {
	l.Lock()
	defer l.Unlock()

	ch := make(chan LightUpdate)
	l.updateChannels = append(l.updateChannels, ch)

	return ch
}

type lightsValueContainer map[string]*Light

func (lvc lightsValueContainer) Get(id string) (interface{}, error) {
	light, ok := lvc[id]
	if ok {
		return adapter.NewWrapper(light), nil
	}
	return nil, nil
}

func (lvc lightsValueContainer) Set(id string, val interface{}) error {
	light, ok := lvc[id]
	if ok {
		return adapter.NewWrapper(light).Set(id, val)
	}
	return nil
}

func (lvc lightsValueContainer) GetAll() (map[string]interface{}, error) {
	all := map[string]interface{}{}

	for key, light := range lvc {
		all[key] = adapter.NewWrapper(light)
	}

	return all, nil
}
