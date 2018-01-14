package adapter

import (
	"fmt"
	"reflect"
	"strings"
)

// Wrappers allows you to wrap anything to be a value container
type Wrapper struct {
	getMapping map[string]func() (interface{}, error)
	setMapping map[string]func(val interface{}) error
}

func NewWrapper(in interface{}) ValueContainer {
	wrapper := &Wrapper{
		getMapping: map[string]func() (interface{}, error){},
		setMapping: map[string]func(val interface{}) error{},
	}

	v := reflect.ValueOf(in)
	t := v.Type()
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		valMethod := v.Method(i)
		name := method.Name

		if strings.HasPrefix(name, "Set") {
			// TODO process different cases
			wrapper.setMapping[strings.ToLower(name[3:])] = func(val interface{}) error {
				res := valMethod.Call([]reflect.Value{reflect.ValueOf(val)})
				if len(res) == 1 {
					resVal := res[0].Interface()
					if errVal, ok := resVal.(error); ok {
						return errVal
					}
				}
				return nil
			}
		}

		if strings.HasPrefix(name, "Get") {
			wrapper.getMapping[strings.ToLower(name[3:])] = func() (interface{}, error) {
				res := valMethod.Call([]reflect.Value{})
				if len(res) == 2 {
					resVal := res[1].Interface()
					errVal, _ := resVal.(error)
					return res[0].Interface(), errVal
				} else if len(res) == 1 {
					return res[0].Interface(), nil
				}
				return nil, nil
			}
		}
	}

	return wrapper
}

func (w *Wrapper) Get(id string) (interface{}, error) {
	fn, ok := w.getMapping[id]
	if !ok {
		return nil, nil
	}
	return fn()
}

func (w *Wrapper) Set(id string, val interface{}) error {
	fn, ok := w.setMapping[id]
	if !ok {
		return nil
	}
	return fn(val)
}

func (w *Wrapper) GetAll() (map[string]interface{}, error) {
	fmt.Printf("Get mappign %+v\n", w.getMapping)

	all := map[string]interface{}{}

	for key, fn := range w.getMapping {
		val, err := fn()

		if err != nil {
			fmt.Printf("Error: %s", err.Error())
			return nil, err
		}

		all[key] = val

	}

	return all, nil
}
