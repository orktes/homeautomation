package deconz

import (
	"reflect"

	"github.com/orktes/homeautomation/bridge/adapter"
)

type updatedKey struct {
	key string
	val interface{}
}

func getAllFromStruct(a interface{}, vc adapter.ValueContainer) map[string]interface{} {
	vals := map[string]interface{}{}
	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		json := field.Tag.Get("json")
		if json != "" {
			val, _ := vc.Get(json)
			vals[json] = val
		}
	}

	return vals
}

func getStructValueByName(a interface{}, key string) interface{} {
	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		valueField := v.Field(i)
		if field.Tag.Get("json") == key {
			if valueField.Kind() == reflect.Ptr {
				if valueField.IsNil() {
					return nil
				}
				return valueField.Elem().Interface()
			}
			return valueField.Interface()
		}
	}

	return nil
}

func mergeStructs(a interface{}, b interface{}) []updatedKey {
	keys := []updatedKey{}

	aV := reflect.ValueOf(a)
	v := reflect.ValueOf(b)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if aV.Kind() == reflect.Ptr {
		aV = aV.Elem()
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		valueField := v.Field(i)
		if !valueField.IsNil() {
			var val interface{}
			if valueField.Kind() == reflect.Ptr {
				val = valueField.Elem().Interface()
			} else {
				val = valueField.Interface()
			}
			keys = append(keys, updatedKey{
				key: field.Tag.Get("json"),
				val: val,
			})

			if v.Field(i).Kind() == reflect.Ptr {
				aV.Field(i).Elem().Set(v.Field(i).Elem())
			} else {
				aV.Field(1).Set(v.Field(1))
			}
		}
	}

	return keys
}

type groupValueContainer struct {
	deconz *Deconz
	group  string
}

func (gcf *groupValueContainer) Get(id string) (interface{}, error) {
	return gcf.deconz.Get(gcf.group + "." + id)
}

func (gcf *groupValueContainer) Set(id string, val interface{}) error {
	return gcf.deconz.Set(gcf.group+"."+id, val)
}

func (gcf *groupValueContainer) GetAll() (map[string]interface{}, error) {
	vals := map[string]interface{}{}
	switch gcf.group {
	case "lights":
		for k, v := range gcf.deconz.lights {
			vals[k] = v
		}
	case "groups":
		for k, v := range gcf.deconz.groups {
			vals[k] = v
		}
	case "sensors":
		for k, v := range gcf.deconz.sensors {
			vals[k] = v
		}
	}

	return vals, nil
}
