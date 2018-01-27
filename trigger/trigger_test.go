package trigger

import (
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/orktes/homeautomation/config"
)

func TestTriggerListenAndGet(t *testing.T) {
	ts := New(config.Config{
		Triggers: []config.Trigger{
			config.Trigger{
				Script: `listen("haaga/foo/bar", function () {
					var val = get("haaga/foo/bar");
					var val2 = get("haaga/foo/foz");
					print(val, val2);
					var now = Date.now();
					sleep(10);
					set("haaga/foo/diz", val + val2);
				})`,
			},
		},
	})

	subs := make(chan struct {
		topic    string
		callback mqtt.MessageHandler
	})
	pubs := make(chan struct {
		topic   string
		payload []byte
	})
	ts.c = &mockClient{subs, pubs}

	go ts.initTriggers()

	s := <-subs
	if s.topic != "haaga/status/foo/bar" {
		t.Error("Wrong topic subscription")
	}

	go ts.handler(nil, &mockMessage{
		topic:   "haaga/status/foo/bar",
		payload: []byte(`"biz"`),
	})

	s = <-subs
	if s.topic != "haaga/status/foo/foz" {
		t.Error("Wrong topic subscription")
	}

	p := <-pubs
	if p.topic != "haaga/get/foo/foz" {
		t.Error("Wrong topic publish", p.topic)
	}

	go ts.handler(nil, &mockMessage{
		topic:   "haaga/status/foo/foz",
		payload: []byte(`"goz"`),
	})

	p = <-pubs
	if p.topic != "haaga/set/foo/diz" {
		t.Error("Wrong topic publish", p.topic)
	}

	if string(p.payload) != `"bizgoz"` {
		t.Error("Wrong payload received", string(p.payload))
	}
}
