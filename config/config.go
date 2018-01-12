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
	Type   string                 `hcl:"type,key"`
	Config map[string]interface{} `hcl:"config"`
}

// Light represents a single light (or a virtual light) config
type Light struct {
	Name   string `hcl:"name,key"`
	Source string `hcl:"source"`
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
	Name        string       `hcl:"name,key"`
	Key         string       `hcl:"key"`
	Value       interface{}  `hcl:"value"`
	EndValue    *interface{} `hcl:"end_value"`
	Target      string       `hcl:"target"`
	TargetValue interface{}  `hcl:"target_value"`
	Interval    int          `hcl:"interval"`
	Delay       int          `hcl:"delay"`
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
