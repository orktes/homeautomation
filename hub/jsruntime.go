package hub

import (
	"sync"

	"github.com/orktes/homeautomation/adapter"

	"github.com/orktes/goja"
)

type jsRuntime struct {
	*goja.Runtime
	sync.Mutex
}

func (js *jsRuntime) RunString(str string) (goja.Value, error) {
	js.Lock()
	defer js.Unlock()

	return js.Runtime.RunString(str)
}

func newJSRuntime() *jsRuntime {
	return &jsRuntime{Runtime: goja.New()}
}

type objectLikeValueContainer struct {
	adapter.ValueContainer
}

func (olvl *objectLikeValueContainer) GetObjectValue(key string) (val interface{}, exists bool) {
	v, err := olvl.Get(key)
	if err != nil {
		return nil, false
	}

	if c, ok := v.(adapter.ValueContainer); ok {
		return &objectLikeValueContainer{c}, true
	}

	return v, true
}

func (olvl *objectLikeValueContainer) SetObjectValue(key string, val interface{}) {
	olvl.Set(key, val)
}

func (olvl *objectLikeValueContainer) GetObjectKeys() []string {
	all, _ := olvl.GetAll()

	keys := make([]string, 0, len(all))

	for key := range all {
		keys = append(keys, key)
	}

	return keys
}

func (olvl *objectLikeValueContainer) GetObjectLength() int {
	all, _ := olvl.GetAll()
	return len(all)
}

func (olvl *objectLikeValueContainer) DeleteObjectValue(key string) {
	// TODO nothing to do here
}
