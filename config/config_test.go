package config

import (
	"strings"
	"testing"
)

func TestConfigParser(t *testing.T) {
	config, err := ParseConfig(strings.NewReader(`
		adapter "deconz" {
			type = "deconz"
			config {
				test = 1
			}
		}

		light "RandomLight" {
			source = "deconz.lights.1"
		}

		sensor "SomeSensor" {
			source = "${deconz.sensors.1}"
			trigger {
				event = [5001, 5003]
				target = "deconz.light.1.state.on"
				value = true
			}
		}
	`))
	if err != nil {
		t.Error(err)
	}

	t.Errorf("%+v\n", config)

}
