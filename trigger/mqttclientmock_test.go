package trigger

import (
	"errors"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type mockMessage struct {
	topic   string
	payload []byte
}

func (mm mockMessage) Duplicate() bool {
	panic("not implemented")
}

func (mm mockMessage) Qos() byte {
	panic("not implemented")
}

func (mm mockMessage) Retained() bool {
	panic("not implemented")
}

func (mm mockMessage) Topic() string {
	return mm.topic
}

func (mm mockMessage) MessageID() uint16 {
	panic("not implemented")
}

func (mm mockMessage) Payload() []byte {
	return mm.payload
}

type mockToken bool

func (mt mockToken) Wait() bool {
	return true
}

func (mt mockToken) WaitTimeout(time.Duration) bool {
	return true
}

func (mt mockToken) Error() error {
	if mt {
		return errors.New("Some error")
	}
	return nil
}

type mockClient struct {
	subs chan struct {
		topic    string
		callback mqtt.MessageHandler
	}
	pubs chan struct {
		topic   string
		payload []byte
	}
}

func (mc *mockClient) IsConnected() bool {
	panic("not implemented")
}

func (mc *mockClient) Connect() mqtt.Token {
	panic("not implemented")
}

func (mc *mockClient) Disconnect(quiesce uint) {
	panic("not implemented")
}

func (mc *mockClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	mc.pubs <- struct {
		topic   string
		payload []byte
	}{topic, payload.([]byte)}
	return mockToken(false)
}

func (mc *mockClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	mc.subs <- struct {
		topic    string
		callback mqtt.MessageHandler
	}{topic, callback}

	return mockToken(false)
}

func (mc *mockClient) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	panic("not implemented")
}

func (mc *mockClient) Unsubscribe(topics ...string) mqtt.Token {
	panic("not implemented")
}

func (mc *mockClient) AddRoute(topic string, callback mqtt.MessageHandler) {
	panic("not implemented")
}
