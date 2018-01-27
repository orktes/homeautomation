package trigger

import "github.com/orktes/goja"

type runtime struct {
	workChannel chan func(r *runtime)
	*goja.Runtime
}

func newRuntime() *runtime {
	gr := goja.New()
	ch := make(chan func(*runtime))

	r := &runtime{
		Runtime:     gr,
		workChannel: ch,
	}

	go func() {
		for w := range ch {
			w(r)
		}
	}()

	return r
}

func (r *runtime) Work(cb func(*runtime)) {
	r.workChannel <- cb
}
