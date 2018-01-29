package config

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"

	"text/template"

	"github.com/gosimple/slug"
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

// AlexaDeviceCapabilityProperty device property
type AlexaDeviceCapabilityProperty struct {
	Name string `hcl:"name,key"`
	Get  string `hcl:"get"`
	Set  string `hcl:"set"`
}

// AlexaDeviceCapability device capability
type AlexaDeviceCapability struct {
	Interface  string                          `hcl:"interface,key"`
	Properties []AlexaDeviceCapabilityProperty `hcl:"property"`
}

// AlexaDevice alexa device config
type AlexaDevice struct {
	ID                string   `hcl:"id,key"`
	Name              string   `hcl:"name"`
	Description       string   `hcl:"description"`
	Manafacturer      string   `hcl:"manafacturer"`
	DisplayCategories []string `hcl:"display_categories"`

	Capabilities []AlexaDeviceCapability `hcl:"capability"`
}

// Alexa mqtt config
type Alexa struct {
	Topic   string        `hcl:"topic"`
	Devices []AlexaDevice `hcl:"device"`
}

// Config represents homeautomation config
type Config struct {
	Servers  []string      `hcl:"servers"`
	Username string        `hcl:"username"`
	Password string        `hcl:"password"`
	ClientID string        `hcl:"client_id"`
	Bridge   *BridgeConfig `hcl:"bridge"`
	Triggers []Trigger     `hcl:"trigger"`
	Alexa    *Alexa        `hcl:"alexa"`
}

// Parse config returns a Config struct pointer parsed from a given reader
func ParseConfig(reader io.Reader) (Config, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return Config{}, err
	}

	tmpl, err := template.New("config").Funcs(map[string]interface{}{
		"toLower": strings.ToLower,
		"toUpper": strings.ToUpper,
		"slugify": slug.Make,
		"array": func(vals ...interface{}) []interface{} {
			return vals
		},
	}).Parse(string(data))
	if err != nil {
		return Config{}, err
	}

	b := &bytes.Buffer{}

	if err := tmpl.Execute(b, map[string]interface{}{}); err != nil {
		return Config{}, err
	}

	conf := &Config{}
	err = hcl.Decode(conf, string(b.Bytes()))

	return *conf, err
}
