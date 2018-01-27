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

// BridgeConfig represents homeautomation config
type BridgeConfig struct {
	Adapters []Adapter `hcl:"adapter"`
	Root     string    `hcl:"root"`
}

// Trigger represents a single toggle
type Trigger struct {
	Script string `hcl:"script"`
}

// Config represents homeautomation config
type Config struct {
	Servers  []string      `hcl:"servers"`
	Username string        `hcl:"username"`
	Password string        `hcl:"password"`
	ClientID string        `hcl:"client_id"`
	Bridge   *BridgeConfig `hcl:"bridge"`
	Triggers []Trigger     `hcl:"trigger"`
}

// Parse config returns a Config struct pointer parsed from a given reader
func ParseConfig(reader io.Reader) (Config, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return Config{}, err
	}

	conf := &Config{}
	err = hcl.Decode(conf, string(data))

	return *conf, err
}
