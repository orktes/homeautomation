package alexa

import (
	"math"

	"github.com/orktes/homeautomation/config"
	"github.com/orktes/homeautomation/util"
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

	switch v := val.(type) {
	case int64:
		val = float64(v)
	case int:
		val = float64(v)
	}

	if v, ok := val.(float64); ok {
		if rangeVal, err := util.ConvertFloatValueToRange(ph.conf.OutputRange, ph.conf.InputRange, v); err == nil {
			val = rangeVal
			if ph.conf.Type == "int" {
				val = int(rangeVal + math.Copysign(0.5, rangeVal))
			}
		}
	}

	return val, nil
}

func (ph *propertyHandler) SetValue(val interface{}) error {
	switch v := val.(type) {
	case int64:
		val = float64(v)
	case int:
		val = float64(v)
	}

	if v, ok := val.(float64); ok {
		if rangeVal, err := util.ConvertFloatValueToRange(ph.conf.InputRange, ph.conf.OutputRange, v); err == nil {
			val = rangeVal
			if ph.conf.Type == "int" {
				val = int(rangeVal + math.Copysign(0.5, rangeVal))
			}
		}
	}

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
