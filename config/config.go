package config

import (
	"io"
	"io/ioutil"

	"github.com/hashicorp/hcl"
)

// Adapter represents a single adapter config
type Adapter struct {
	ID     string                 `hcl:"id,key"`
	Type   string                 `hcl:"type"`
	Config map[string]interface{} `hcl:"config"`
}

// Frontend represents a single frontend config
type Frontend struct {
	ID     string                 `hcl:"id,key"`
	Type   string                 `hcl:"type"`
	Config map[string]interface{} `hcl:"config"`
}

// LightState represent configurable values in a light
type LightState struct {
	Brightness string `hcl:"bri"`
	On         string `hcl:"on"`
}

// Light represents a single light (or a virtual light) config
type Light struct {
	Name  string     `hcl:"name,key"`
	Read  LightState `hcl:"read"`
	Write LightState `hcl:"write"`
}

// SwitchState represents a switch state
type SwitchState struct {
	On string `hcl:"ok"`
}

// Switch represents a switch
type Switch struct {
	Name  string      `hcl:"name,key"`
	Read  SwitchState `hcl:"read"`
	Write SwitchState `hcl:"write"`
}

// Amplifier represent a single amplifier
type Amplifier struct {
	Name   string `hcl:"name,key"`
	Source string `hcl:"source"`
}

// MediaController represent a single amplifier
type MediaController struct {
	Name   string `hcl:"name,key"`
	Source string `hcl:"source"`
}

// Television represents a single television
type Television struct {
	Name   string `hcl:"name,key"`
	Source string `hcl:"source"`
}

// Trigger trigger something
type Trigger struct {
	Name         string `hcl:"name,key"`
	Key          string `hcl:"key"`
	Condition    string `hcl:"condition"`
	EndCondition string `hcl:"end_condition"`
	Action       string `hcl:"action"`
	Interval     int    `hcl:"interval"`
	Delay        int    `hcl:"delay"`
}

// Sensor defines a single sensor and the triggers for it
type Sensor struct {
	Name   string `hcl:"name,key"`
	Source string `hcl:"source"`
}

// Config represents homeautomation config
type Config struct {
	Adapters         []Adapter         `hcl:"adapter"`
	Frontends        []Frontend        `hcl:"frontend"`
	Sensors          []Sensor          `hcl:"sensor"`
	Trigger          []Trigger         `hcl:"trigger"`
	Lights           []Light           `hcl:"light"`
	Switchs          []Switch          `hcl:"switch"`
	Amplifiers       []Amplifier       `hcl:"amplifier"`
	MediaControllers []MediaController `hcl:"mediacontroller"`
	Televisions      []Television      `hcl:"television"`
}

// Parse config returns a Config struct pointer parsed from a given reader
func ParseConfig(reader io.Reader) (*Config, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	err = hcl.Decode(conf, string(data))

	return conf, err
}
