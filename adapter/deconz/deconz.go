package deconz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/orktes/homeautomation/adapter"
	"github.com/orktes/homeautomation/hub"
	"github.com/orktes/homeautomation/registry"
)

var (
	ErrorNotFound = errors.New("device not found")
)

var instances = map[string]*Deconz{}

type Deconz struct {
	id  string
	key string

	hostname string
	port     int

	updateChannels []chan adapter.Update

	lights  map[string]*lightDevice
	groups  map[string]*groupDevice
	sensors map[string]*sensorDevice

	sync.RWMutex
}

func (deconz *Deconz) Set(id string, val interface{}) error {
	if id == "" {
		return errors.New("Value can't be assigned to the deconz router")
	}

	parts := strings.Split(id, ".")

	parent := strings.Join(parts[0:len(parts)-1], ".")
	c, err := deconz.Get(parent)
	if err != nil {
		return err
	}

	vc, ok := c.(adapter.ValueContainer)
	if !ok {
		return fmt.Errorf("Parent %s is not a value container", parent)
	}

	return vc.Set(parts[len(parts)-1], val)
}

func (deconz *Deconz) Get(id string) (interface{}, error) {
	deconz.RLock()
	defer deconz.RUnlock()

	if id == "" {
		return deconz, nil
	}

	parts := strings.Split(id, ".")

	if len(parts) == 1 {
		return &groupValueContainer{group: parts[0], deconz: deconz}, nil
	}

	var res adapter.ValueContainer

	switch parts[0] {
	case "lights":
		if light, ok := deconz.lights[parts[1]]; ok {
			res = light
		}
	case "groups":
		if grp, ok := deconz.groups[parts[1]]; ok {
			res = grp
		}
	case "sensors":
		if sens, ok := deconz.sensors[parts[1]]; ok {
			res = sens
		}
	}

	if res != nil {
		if len(parts) > 2 {
			return res.Get(strings.Join(parts[2:], "."))
		}

		return res, nil
	}

	return nil, ErrorNotFound
}

func (deconz *Deconz) GetAll() (map[string]interface{}, error) {
	devices := map[string]interface{}{}
	for _, key := range []string{"lights", "groups", "sensors"} {
		d, err := deconz.Get(key)
		if err != nil {
			return nil, err
		}
		devices[key] = d
	}

	return devices, nil
}

func (deconz *Deconz) UpdateChannel() <-chan adapter.Update {
	deconz.RLock()
	defer deconz.RUnlock()

	ch := make(chan adapter.Update)
	deconz.updateChannels = append(deconz.updateChannels, ch)
	return ch
}

func (deconz *Deconz) init() {
	deconz.register()
	deconz.setupWSConnection()
	deconz.fetchInitialState()

}

func (deconz *Deconz) get(path string, in interface{}) error {
	res, err := http.Get(
		fmt.Sprintf("http://%s:%d/api/%s/%s", deconz.hostname, deconz.port, deconz.key, path),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(in)
}

func (deconz *Deconz) post(path string, out interface{}, in interface{}) error {
	return deconz.request("POST", path, out, in)
}

func (deconz *Deconz) put(path string, out interface{}, in interface{}) error {
	return deconz.request("PUT", path, out, in)
}

func (deconz *Deconz) request(method string, path string, out interface{}, in interface{}) error {
	var r io.Reader

	if in != nil {
		data, err := json.Marshal(out)
		if err != nil {
			return err
		}
		r = bytes.NewReader(data)
	}
	client := &http.Client{}
	req, err := http.NewRequest(
		method,
		fmt.Sprintf("http://%s:%d/api/%s/%s", deconz.hostname, deconz.port, deconz.key, path),
		r,
	)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(in)
}

func (deconz *Deconz) setupWSConnection() {
start:
	configRes := &configResponse{}
	err := deconz.get("config", configRes)
	if err != nil {
		fmt.Printf("Unabled establish websocket connection: %s", err.Error())
		fmt.Print("Retrying in 5 seconds")
		time.Sleep(5 * time.Second)
		goto start
	}
	go deconz.initWebsocketConnection(configRes.IPAddress, configRes.WebsocketPort)
}

func (deconz *Deconz) initWebsocketConnection(host string, port int) {
	defer func() {
		// Keep the connection up no matter what happens
		fmt.Print("Retrying in 5 seconds")
		time.Sleep(5 * time.Second)
		go deconz.initWebsocketConnection(host, port)
	}()

	url := fmt.Sprintf("ws://%s:%d", host, port)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		panic(err)
	}

	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			fmt.Printf("Error occured while reading data from the websocket: %s", err.Error())
			return
		}
		if messageType == websocket.CloseMessage {
			return
		}
		ev := &event{}
		err = json.Unmarshal(message, ev)
		if err != nil {
			fmt.Printf("Error occured while parsing event: %s", err.Error())
			continue
		}
		if ev.Event == "changed" {
			device, err := deconz.Get(ev.Route + "." + ev.ID)
			if err != nil {
				fmt.Printf("Error occured while updating device with event event: %s", err.Error())
				continue
			}

			switch d := device.(type) {
			case *lightDevice:
				s := &lightState{}
				ev.unmarshalState(s)
				d.updateState(s)
			case *groupDevice:
				s := &groupState{}
				ev.unmarshalState(s)
				d.updateState(s)
			case *sensorDevice:
				s := &sensorState{}
				ev.unmarshalState(s)
				d.updateState(s)
			}

		}
	}
}

func (deconz *Deconz) fetchInitialState() {
	err := deconz.getLights()
	if err != nil {
		panic(err)
	}

	err = deconz.getGroups()
	if err != nil {
		panic(err)
	}

	err = deconz.getSensors()
	if err != nil {
		panic(err)
	}
}

func (deconz *Deconz) getLights() error {
	lights := map[string]light{}
	if err := deconz.get("lights", &lights); err != nil {
		return err
	}

	deconz.Lock()
	defer deconz.Unlock()
	for id, light := range lights {
		ld := &lightDevice{id: id, deconz: deconz, data: light}
		deconz.pipeUpdates(ld)
		deconz.lights[id] = ld
	}
	return nil
}

func (deconz *Deconz) getGroups() error {
	groups := map[string]group{}
	if err := deconz.get("groups", &groups); err != nil {
		return err
	}

	deconz.Lock()
	defer deconz.Unlock()
	for id, group := range groups {
		gd := &groupDevice{id: id, deconz: deconz, data: group}
		deconz.pipeUpdates(gd)
		deconz.groups[id] = gd
	}
	return nil
}

func (deconz *Deconz) getSensors() error {
	sensors := map[string]sensor{}
	if err := deconz.get("sensors", &sensors); err != nil {
		return err
	}

	deconz.Lock()
	defer deconz.Unlock()
	for id, sensor := range sensors {
		sd := &sensorDevice{id: id, deconz: deconz, data: sensor}
		deconz.pipeUpdates(sd)
		deconz.sensors[id] = sd
	}
	return nil
}

func (deconz *Deconz) pipeUpdates(d adapter.Device) {
	go func() {
		ch := d.UpdateChannel()
		for u := range ch {
			deconz.sendUpdate(u)
		}
	}()
}

func (deconz *Deconz) sendUpdate(update adapter.Update) {
	deconz.Lock()
	defer deconz.Unlock()

	for _, ch := range deconz.updateChannels {
		ch <- update
	}
}

func (deconz *Deconz) register() {
	if deconz.key != "" {
		return
	}
	// TODO handler register
	panic("Not yet implemented")
}

// Create returns a new Deconz instance
func Create(id string, config map[string]interface{}, hub *hub.Hub) (adapter.Adapter, error) {
	if deconz, ok := instances[id]; ok {
		return deconz, nil
	}

	hostname, _ := config["hostname"].(string)
	port, _ := config["port"].(int)
	key, _ := config["key"].(string)

	deconz := &Deconz{
		id:       id,
		hostname: hostname,
		port:     port,
		key:      key,

		lights:  map[string]*lightDevice{},
		groups:  map[string]*groupDevice{},
		sensors: map[string]*sensorDevice{},
	}

	deconz.init()

	instances[id] = deconz

	return deconz, nil
}

func init() {
	registry.RegisterAdapter("deconz", Create)
}
