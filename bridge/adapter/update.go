package adapter

import (
	"sync"
)

// ValueUpdate reperesents a single updated value
type ValueUpdate struct {
	Key   string
	Value interface{}
}

// Update a device update event
type Update struct {
	ValueContainer ValueContainer
	Updates        []ValueUpdate
}

// Updater is a helper struct for implementing UpdateChannels
type Updater struct {
	updateChannels []chan Update
	sync.Mutex
}

func (ld *Updater) UpdateChannel() <-chan Update {
	ld.Lock()
	defer ld.Unlock()
	ch := make(chan Update)
	ld.updateChannels = append(ld.updateChannels, ch)
	return ch
}

func (ld *Updater) SendUpdate(u Update) {
	ld.Lock()
	defer ld.Unlock()
	for _, ch := range ld.updateChannels {
		ch <- u
	}
}
