package mqtt

import (
	"errors"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/orktes/homeautomation/config"
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

func TestMQTTBridgeGet(t *testing.T) {
	ma := &mockAdapter{
		id:   "adid",
		vals: map[string]interface{}{},
	}
	ma.Set("foo", "bar")

	bridge := New(config.Config{}, ma)

	subs := make(chan struct {
		topic    string
		callback mqtt.MessageHandler
	})
	pubs := make(chan struct {
		topic   string
		payload []byte
	})
	bridge.c = &mockClient{subs, pubs}

	t.Run("publish statuses", func(t *testing.T) {
		go bridge.publishStatuses()

		stus := <-pubs
		if stus.topic != "adid/connected" || string(stus.payload) != "2" {
			t.Error("Wrong publish received", stus.topic, stus.payload)
		}

		stus = <-pubs
		if stus.topic != "adid/status/foo" || string(stus.payload) != "\"bar\"" {
			t.Error("Wrong publish received", stus.topic, string(stus.payload))
		}

	})

	t.Run("subscribe", func(t *testing.T) {
		go bridge.subscribeToTopics()
		sub := <-subs
		if sub.topic != "adid/set#" {
			t.Error("Wrong topic subs", sub.topic)
		}

		sub = <-subs
		if sub.topic != "adid/get#" {
			t.Error("Wrong topic subs", sub.topic)
		}

	})

	t.Run("value update", func(t *testing.T) {
		// Try updating a value

		go ma.Set("foo", "biz")

		stus := <-pubs
		if stus.topic != "adid/status/foo" || string(stus.payload) != "\"biz\"" {
			t.Error("Wrong publish received", stus.topic, string(stus.payload))
		}
	})

	t.Run("value get", func(t *testing.T) {
		// try retrieving back a value
		go bridge.defaultHandler(bridge.c, &mockMessage{topic: "adid/get/foo", payload: []byte{}})
		stus := <-pubs
		if stus.topic != "adid/status/foo" || string(stus.payload) != "\"biz\"" {
			t.Error("Wrong publish received", stus.topic, string(stus.payload))
		}
	})

	t.Run("Root ID", func(t *testing.T) {
		// Try with root
		bridge.conf.Root = "bridgeroot"

		t.Run("publish statuses", func(t *testing.T) {

			go bridge.publishStatuses()

			stus := <-pubs
			if stus.topic != "bridgeroot/connected" || string(stus.payload) != "2" {
				t.Error("Wrong publish received", stus.topic, stus.payload)
			}

			stus = <-pubs
			if stus.topic != "bridgeroot/status/adid/foo" || string(stus.payload) != "\"biz\"" {
				t.Error("Wrong publish received", stus.topic, string(stus.payload))
			}
		})

		t.Run("subscribe", func(t *testing.T) {
			go bridge.subscribeToTopics()
			sub := <-subs
			if sub.topic != "bridgeroot/set#" {
				t.Error("Wrong topic subs", sub.topic)
			}

			sub = <-subs
			if sub.topic != "bridgeroot/get#" {
				t.Error("Wrong topic subs", sub.topic)
			}

		})

		t.Run("value update", func(t *testing.T) {
			// Try updating a value

			go ma.Set("foo", "biz")

			stus := <-pubs
			if stus.topic != "bridgeroot/status/adid/foo" || string(stus.payload) != "\"biz\"" {
				t.Error("Wrong publish received", stus.topic, string(stus.payload))
			}
		})

		t.Run("value get", func(t *testing.T) {
			// try retrieving back a value
			go bridge.defaultHandler(bridge.c, &mockMessage{topic: "bridgeroot/get/adid/foo", payload: []byte{}})
			stus := <-pubs
			if stus.topic != "bridgeroot/status/adid/foo" || string(stus.payload) != "\"biz\"" {
				t.Error("Wrong publish received", stus.topic, string(stus.payload))
			}
		})
	})

	t.Run("Subroot", func(t *testing.T) {
		bridge.conf.Root = "bridgeroot/subroot"
		t.Run("publish statuses", func(t *testing.T) {

			go bridge.publishStatuses()

			stus := <-pubs
			if stus.topic != "bridgeroot/connected" || string(stus.payload) != "2" {
				t.Error("Wrong publish received", stus.topic, stus.payload)
			}

			stus = <-pubs
			if stus.topic != "bridgeroot/status/subroot/adid/foo" || string(stus.payload) != "\"biz\"" {
				t.Error("Wrong publish received", stus.topic, string(stus.payload))
			}
		})

		t.Run("subscribe", func(t *testing.T) {
			go bridge.subscribeToTopics()
			sub := <-subs
			if sub.topic != "bridgeroot/set#" {
				t.Error("Wrong topic subs", sub.topic)
			}

			sub = <-subs
			if sub.topic != "bridgeroot/get#" {
				t.Error("Wrong topic subs", sub.topic)
			}

		})

		t.Run("value update", func(t *testing.T) {
			// Try updating a value

			go ma.Set("foo", "biz")

			stus := <-pubs
			if stus.topic != "bridgeroot/status/subroot/adid/foo" || string(stus.payload) != "\"biz\"" {
				t.Error("Wrong publish received", stus.topic, string(stus.payload))
			}
		})

		t.Run("value get", func(t *testing.T) {
			// try retrieving back a value
			go bridge.defaultHandler(bridge.c, &mockMessage{topic: "bridgeroot/get/subroot/adid/foo", payload: []byte{}})
			stus := <-pubs
			if stus.topic != "bridgeroot/status/subroot/adid/foo" || string(stus.payload) != "\"biz\"" {
				t.Error("Wrong publish received", stus.topic, string(stus.payload))
			}
		})
	})

	t.Run("new value", func(t *testing.T) {
		bridge.conf.Root = ""
		go ma.Set("biz", "foz")

		stus := <-pubs
		if stus.topic != "adid/status/biz" || string(stus.payload) != "\"foz\"" {
			t.Error("Wrong publish received", stus.topic, string(stus.payload))
		}

		t.Run("multi value get", func(t *testing.T) {
			go bridge.defaultHandler(bridge.c, &mockMessage{topic: "adid/get", payload: []byte{}})
			stus := <-pubs
			if stus.topic != "adid/status/biz" && stus.topic != "adid/status/foo" {
				t.Error("Wrong topic received", stus.topic)
			}

			stus2 := <-pubs
			if stus2.topic != "adid/status/biz" && stus2.topic != "adid/status/foo" {
				t.Error("Wrong topic received", stus2.topic)
			}

			if stus.topic == stus2.topic {
				t.Error("Same message received twice")
			}
		})

		t.Run("multi value get with root", func(t *testing.T) {
			bridge.conf.Root = "rootpath"
			go bridge.defaultHandler(bridge.c, &mockMessage{topic: "rootpath/get/adid", payload: []byte{}})
			stus := <-pubs
			if stus.topic != "rootpath/status/adid/biz" && stus.topic != "rootpath/status/adid/foo" {
				t.Error("Wrong topic received", stus.topic)
			}

			stus2 := <-pubs
			if stus.topic != "rootpath/status/adid/biz" && stus.topic != "rootpath/status/adid/foo" {
				t.Error("Wrong topic received", stus2.topic)
			}

			if stus.topic == stus2.topic {
				t.Error("Same message received twice")
			}
		})

		t.Run("multi value get with root & subroot", func(t *testing.T) {
			bridge.conf.Root = "rootpath/something"
			check := func() {
				stus := <-pubs
				if stus.topic != "rootpath/status/something/adid/biz" && stus.topic != "rootpath/status/something/adid/foo" {
					t.Error("Wrong topic received", stus.topic)
				}

				stus2 := <-pubs
				if stus.topic != "rootpath/status/something/adid/biz" && stus.topic != "rootpath/status/something/adid/foo" {
					t.Error("Wrong topic received", stus2.topic)
				}

				if stus.topic == stus2.topic {
					t.Error("Same message received twice")
				}
			}

			go bridge.defaultHandler(bridge.c, &mockMessage{topic: "rootpath/get", payload: []byte{}})
			check()

			go bridge.defaultHandler(bridge.c, &mockMessage{topic: "rootpath/get/something", payload: []byte{}})
			check()
		})
	})

}

func TestMQTTBridgeSet(t *testing.T) {
	ma := &mockAdapter{
		id:   "adid",
		vals: map[string]interface{}{},
	}

	bridge := New(config.Config{}, ma)

	subs := make(chan struct {
		topic    string
		callback mqtt.MessageHandler
	})
	pubs := make(chan struct {
		topic   string
		payload []byte
	})
	bridge.c = &mockClient{subs, pubs}

	go bridge.defaultHandler(bridge.c, &mockMessage{topic: "adid/set/bar", payload: []byte("\"foo\"")})
	stus := <-pubs
	if stus.topic != "adid/status/bar" {
		t.Error("Wrong topic received", stus.topic)
	}

	if ma.vals["bar"] != "foo" {
		t.Error("Wrong val received", ma.vals)
	}
}
