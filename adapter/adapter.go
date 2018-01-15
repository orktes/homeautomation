package adapter

// Adapter defines the interface all adapters should implement
type Adapter interface {
	ValueContainer
	Close() error
	UpdateChannel() <-chan Update
}
