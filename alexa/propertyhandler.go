package alexa

import (
	"github.com/orktes/homeautomation/config"
)

type propertyHandler struct {
	alexa *Alexa
	conf  config.AlexaDeviceCapabilityProperty
}

func (ph *propertyHandler) GetValue() (interface{}, error) {
	gval, err := ph.alexa.exec(ph.conf.Get, nil)
	if err != nil {
		return nil, err
	}

	val := gval.Export()

	return val, nil
}

func (ph *propertyHandler) SetValue(val interface{}) error {
	_, err := ph.alexa.exec(ph.conf.Set, map[string]interface{}{"value": val})
	if err != nil {
		return err
	}
	return nil
}

func (ph *propertyHandler) UpdateChannel() <-chan interface{} {
	// TODO
	return nil
}
