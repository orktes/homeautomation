package deconz

import (
	"github.com/orktes/homeautomation/bridge/adapter"
)

type lightDevice struct {
	id             string
	deconz         *Deconz
	data           light
	updateChannels []chan adapter.Update
}

func (ld *lightDevice) ID() string {
	return ld.deconz.id + "/lights/" + ld.id
}

func (ld *lightDevice) UpdateChannel() <-chan adapter.Update {
	ch := make(chan adapter.Update)
	ld.updateChannels = append(ld.updateChannels, ch)
	return ch
}

func (ld *lightDevice) Get(id string) (interface{}, error) {
	if id == "name" {
		return ld.data.Name, nil
	}
	return getStructValueByName(ld.data.State, id), nil
}

func (ld *lightDevice) GetAll() (map[string]interface{}, error) {
	all := getAllFromStruct(ld.data.State, ld)
	all["name"] = ld.data.Name
	return all, nil
}

func (ld *lightDevice) Set(id string, val interface{}) error {
	stateUpdate := map[string]interface{}{}
	stateUpdate[id] = val

	res := &lightStateChangeResponse{}
	return ld.deconz.put("/lights/"+ld.id+"/state", stateUpdate, res)
}

func (ld *lightDevice) updateState(state *lightState, sendUnchanged bool) {
	keys := mergeStructs(&ld.data.State, state)
	du := adapter.Update{}
	du.ValueContainer = ld
	for _, kp := range keys {
		if sendUnchanged || kp.changed {
			du.Updates = append(du.Updates, adapter.ValueUpdate{Key: ld.ID() + "/" + kp.key, Value: kp.val})
		}
	}

	for _, ch := range ld.updateChannels {
		ch <- du
	}
}

type groupDevice struct {
	id             string
	deconz         *Deconz
	data           group
	updateChannels []chan adapter.Update
}

func (gd *groupDevice) ID() string {
	return gd.deconz.id + "/groups/" + gd.id
}

func (gd *groupDevice) UpdateChannel() <-chan adapter.Update {
	ch := make(chan adapter.Update)
	gd.updateChannels = append(gd.updateChannels, ch)
	return ch
}

func (gd *groupDevice) Get(id string) (interface{}, error) {

	switch id {
	case "any_on":
		return *gd.data.State.AnyOn, nil
	case "name":
		return gd.data.Name, nil
	}

	return getStructValueByName(gd.data.Action, id), nil
}

func (gd *groupDevice) GetAll() (map[string]interface{}, error) {
	vals := getAllFromStruct(gd.data.Action, gd)
	vals["any_on"] = *gd.data.State.AnyOn
	vals["name"] = gd.data.Name
	return vals, nil
}

func (gd *groupDevice) Set(id string, val interface{}) error {
	stateUpdate := map[string]interface{}{}
	stateUpdate[id] = val

	res := &lightStateChangeResponse{}
	err := gd.deconz.put("/groups/"+gd.id+"/action", stateUpdate, res)

	if err == nil {
		// Groups are slow to update state in deconz. So lets fake it a little bit
		du := adapter.Update{}
		du.ValueContainer = gd
		du.Updates = []adapter.ValueUpdate{
			{
				Key:   gd.ID() + "/" + id,
				Value: val,
			},
		}

		for _, ch := range gd.updateChannels {
			ch <- du
		}
	}

	return err
}

func (gd *groupDevice) updateState(state *groupState, sendUnchanged bool) {
	keys := mergeStructs(&gd.data.State, state)
	du := adapter.Update{}
	du.ValueContainer = gd
	for _, kp := range keys {
		if sendUnchanged || kp.changed {
			du.Updates = append(du.Updates, adapter.ValueUpdate{Key: gd.ID() + "/" + kp.key, Value: kp.val})
		}
	}

	for _, ch := range gd.updateChannels {
		ch <- du
	}
}

type sensorDevice struct {
	id             string
	deconz         *Deconz
	data           sensor
	updateChannels []chan adapter.Update
}

func (sd *sensorDevice) ID() string {
	return sd.deconz.id + "/sensors/" + sd.id
}

func (sd *sensorDevice) UpdateChannel() <-chan adapter.Update {
	ch := make(chan adapter.Update)
	sd.updateChannels = append(sd.updateChannels, ch)
	return ch
}

func (sd *sensorDevice) Get(id string) (interface{}, error) {
	if id == "name" {
		return sd.data.Name, nil
	}

	val := getStructValueByName(sd.data.State, id)
	res := map[string]interface{}{
		"value":   val,
		"updated": sd.data.State.LastUpdated,
	}

	return res, nil
}

func (sd *sensorDevice) GetAll() (map[string]interface{}, error) {
	all := getAllFromStruct(sd.data.State, sd)
	all["name"] = sd.data.Name
	delete(all, "lastupdated")
	return all, nil
}

func (sd *sensorDevice) Set(id string, val interface{}) error {
	return nil
}

func (sd *sensorDevice) updateState(state *sensorState, sendUnchanged bool) {
	keys := mergeStructs(&sd.data.State, state)
	du := adapter.Update{}
	du.ValueContainer = sd
	for _, kp := range keys {
		if kp.key == "lastupdated" {
			// NO reason to re-send lastupdated
			continue
		}

		if sendUnchanged || kp.changed {
			du.Updates = append(du.Updates, adapter.ValueUpdate{
				Key: sd.ID() + "/" + kp.key,
				Value: map[string]interface{}{
					"updated": sd.data.State.LastUpdated,
					"value":   kp.val,
				},
			})
		}
	}

	for _, ch := range sd.updateChannels {
		ch <- du
	}
}
