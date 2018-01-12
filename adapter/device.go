package adapter

// Device interface is implemented by all devices returned by the adapters
type Device interface {
	ValueContainer
	ID() string
	UpdateChannel() <-chan Update
}
