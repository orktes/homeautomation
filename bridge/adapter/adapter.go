package adapter

// Adapter defines the interface all adapters should implement
type Adapter interface {
	Device
	Close() error
}
