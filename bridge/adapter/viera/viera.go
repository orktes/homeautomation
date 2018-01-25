package viera

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/huin/goupnp"
	"github.com/orktes/homeautomation/bridge/adapter"
)

type VieraDiscovery struct {
	adapter.Updater

	id string

	// TODO make the mac address a map from uuid to mac
	mac string
	tvs []*VieraTV

	sync.Mutex
}

func (vd *VieraDiscovery) initialize() error {
	responses, err := goupnp.DiscoverDevices("urn:panasonic-com:service:p00NetworkControl:1")
	if err != nil {
		return err
	}

	for i, info := range responses {
		tv := &VieraTV{id: fmt.Sprintf("%s/%d", vd.id, i+1), host: info.Root.URLBase.Host, mac: vd.mac}
		vd.pipeUpdates(tv)
		if err := tv.init(); err != nil {
			return err
		}

		vd.tvs = append(vd.tvs, tv)
	}

	return nil
}

func (vd *VieraDiscovery) pipeUpdates(d adapter.Device) {
	go func() {
		ch := d.UpdateChannel()
		for u := range ch {
			vd.Updater.SendUpdate(u)
		}
	}()
}

func (vd *VieraDiscovery) Get(id string) (interface{}, error) {
	vd.Lock()
	defer vd.Unlock()

	if id == "" {
		return vd, nil
	}

	parts := strings.Split(id, "/")
	intID, _ := strconv.ParseInt(parts[0], 10, 64)
	if intID >= 1 && len(vd.tvs) >= int(intID) {
		tv := vd.tvs[int(intID)-1]
		if len(parts) == 1 {
			return tv, nil
		}

		return tv.Get(strings.Join(parts[1:], "/"))
	}

	return nil, nil
}

func (vd *VieraDiscovery) Set(id string, val interface{}) error {
	if id == "" {
		return nil
	}

	parts := strings.Split(id, "/")
	c, err := vd.Get(parts[0])
	if err != nil {
		return err
	}
	vc, ok := c.(adapter.ValueContainer)
	if !ok {
		return fmt.Errorf("Parent %s is not a value container", parts[0])
	}

	return vc.Set(strings.Join(parts[1:], "/"), val)
}

func (vd *VieraDiscovery) GetAll() (map[string]interface{}, error) {
	vd.Lock()
	l := len(vd.tvs)
	vd.Unlock()

	vals := map[string]interface{}{}
	for i := 1; i <= l; i++ {
		id := fmt.Sprintf("%d", i)
		c, err := vd.Get(id)
		if err != nil {
			return nil, err
		}
		vals[id] = c
	}

	return vals, nil
}

func (vd *VieraDiscovery) ID() string {
	return vd.id
}

func (vd *VieraDiscovery) UpdateChannel() <-chan adapter.Update {
	return vd.Updater.UpdateChannel()
}

func (vd *VieraDiscovery) Close() error {
	panic("Not yet implemented")
}

// Create returns a new denon dra instance
func Create(id string, config map[string]interface{}) (adapter.Adapter, error) {

	mac, _ := config["mac"].(string)

	viera := &VieraDiscovery{id: id, mac: mac}
	return viera, viera.initialize()
}
