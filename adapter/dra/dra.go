package dra

import (
	"fmt"
	"strings"
	"time"

	denondra "github.com/orktes/go-dra"
	"github.com/orktes/homeautomation/adapter"
	"github.com/orktes/homeautomation/hub"
	"github.com/orktes/homeautomation/registry"
)

type DRA struct {
	id   string
	addr string
	*denondra.DRA

	adapter.Updater
}

func (dra *DRA) connect() {
start:
	d, err := denondra.NewFromAddr(dra.addr)
	if err != nil {
		fmt.Printf("Unable to connect to DRA: %s\n", err.Error())
		time.Sleep(5 * time.Second)
		goto start
	}

	d.OnUpdate = make(chan string)
	dra.DRA = d

	for u := range d.OnUpdate {
		switch {
		case strings.HasPrefix(u, "MV"):
			dra.SendUpdate(adapter.Update{
				ValueContainer: dra,
				Updates: []adapter.ValueUpdate{
					adapter.ValueUpdate{
						Key:   dra.id + ".master_volume",
						Value: dra.GetMasterVolume(),
					},
				},
			})
		case strings.HasPrefix(u, "MU"):
			dra.SendUpdate(adapter.Update{
				ValueContainer: dra,
				Updates: []adapter.ValueUpdate{
					adapter.ValueUpdate{
						Key:   dra.id + ".mute",
						Value: dra.GetMute(),
					},
				},
			})
		case strings.HasPrefix(u, "SI"):
			dra.SendUpdate(adapter.Update{
				ValueContainer: dra,
				Updates: []adapter.ValueUpdate{
					adapter.ValueUpdate{
						Key:   dra.id + ".input",
						Value: dra.GetInput(),
					},
				},
			})
		case strings.HasPrefix(u, "PW"):
			dra.SendUpdate(adapter.Update{
				ValueContainer: dra,
				Updates: []adapter.ValueUpdate{
					adapter.ValueUpdate{
						Key:   dra.id + ".power",
						Value: dra.GetPower(),
					},
				},
			})
		}
	}

	time.Sleep(5 * time.Second)
	goto start
}

func (dra *DRA) ID() string {
	return dra.id
}

func (dra *DRA) Get(id string) (interface{}, error) {
	if dra.DRA == nil {
		return nil, nil
	}

	switch id {
	case "master_volume":
		return dra.DRA.GetMasterVolume(), nil
	case "mute":
		return dra.DRA.GetMute(), nil
	case "power":
		return dra.DRA.GetPower(), nil
	case "input":
		return dra.DRA.GetInput(), nil
	}

	return nil, nil
}

func (dra *DRA) Set(id string, val interface{}) error {
	if dra.DRA == nil {
		return nil
	}

	switch id {
	case "master_volume":
		if intval, ok := val.(int); ok {
			return dra.DRA.SetMasterVolume(intval)
		} else if strval, ok := val.(string); ok {
			if strval == "UP" {
				return dra.DRA.Send("MVUP")
			} else if strval == "DOWN" {
				return dra.DRA.Send("MVDOWN")
			}
		}

	case "mute":
		if boolval, ok := val.(bool); ok {
			return dra.DRA.SetMute(boolval)
		}
	case "power":
		if boolval, ok := val.(bool); ok {
			return dra.DRA.SetPower(boolval)
		}
	case "input":
		if strval, ok := val.(string); ok {
			return dra.DRA.SetInput(strval)
		}
	}
	return nil
}

func (dra *DRA) GetAll() (map[string]interface{}, error) {
	vals := map[string]interface{}{}

	for _, key := range []string{"master_volume", "mute", "power", "input"} {
		val, err := dra.Get(key)
		if err != nil {
			return nil, err
		}
		vals[key] = val
	}

	return vals, nil
}

func (dra *DRA) UpdateChannel() <-chan adapter.Update {
	return dra.Updater.UpdateChannel()
}

// Create returns a new denon dra instance
func Create(id string, config map[string]interface{}, hub *hub.Hub) (adapter.Adapter, error) {
	dra := &DRA{
		id:   id,
		addr: config["address"].(string),
	}

	go dra.connect()

	return dra, nil

}

func init() {
	registry.RegisterAdapter("dra", Create)
}
