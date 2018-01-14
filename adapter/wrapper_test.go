package adapter

import "testing"

type wrapperTester struct {
	foo int
	bar string
}

func (w *wrapperTester) GetFoo() (int, error) {
	return w.foo, nil
}

func (w *wrapperTester) GetBar() (string, error) {
	return w.bar, nil
}

func (w *wrapperTester) SetFoo(val int) error {
	w.foo = val
	return nil
}

func (w *wrapperTester) SetBar(val string) (string, error) {
	w.bar = val
	return "", nil
}

func TestWrapper(t *testing.T) {
	valueContainer := NewWrapper(&wrapperTester{123, "bar"})

	foo, err := valueContainer.Get("foo")
	if err != nil {
		t.Error("Should no return error", err)
	}

	if foo.(int) != 123 {
		t.Error("Should return int 123")
	}

	bar, err := valueContainer.Get("bar")
	if err != nil {
		t.Error("Should no return error", err)
	}

	if bar.(string) != "bar" {
		t.Error("Should return int 123")
	}

	all, err := valueContainer.GetAll()
	if err != nil {
		t.Error("Should no return error", err)
	}

	if all["bar"].(string) != "bar" {
		t.Error("Wrong bar value")
	}
	if all["foo"].(int) != 123 {
		t.Error("Wrong bar value")
	}
}
