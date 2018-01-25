package mqtt

import (
	"encoding/json"
	"fmt"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/orktes/homeautomation/bridge/adapter"
	"github.com/orktes/homeautomation/bridge/util"
	"github.com/orktes/homeautomation/config"
)

type MQTTBridge struct {
	adapter adapter.Adapter
	conf    config.BridgeConfig
	c       mqtt.Client
}

func New(conf config.BridgeConfig, adapter adapter.Adapter) *MQTTBridge {
	bri := &MQTTBridge{conf: conf, adapter: adapter}
	bri.subscribeToAdapter()
	return bri
}

func (bridge *MQTTBridge) Connect() error {
	conf := bridge.conf
	opts := mqtt.NewClientOptions()
	for _, server := range conf.Servers {

		opts = opts.AddBroker(server)
	}
	if conf.ClientID != "" {
		opts = opts.SetClientID(conf.ClientID)
	}
	if conf.Username != "" {
		opts = opts.SetUsername(conf.Username)
	}
	if conf.Password != "" {
		opts = opts.SetPassword(conf.Password)
	}
	opts = opts.SetDefaultPublishHandler(bridge.defaultHandler)
	c := mqtt.NewClient(opts)

	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	bridge.c = c

	if err := bridge.publishStatuses(); err != nil {
		return err
	}

	if err := bridge.subscribeToTopics(); err != nil {
		return err
	}

	return nil
}

func (bridge *MQTTBridge) Disconnect(wait uint) error {
	bridge.c.Disconnect(wait)
	return nil
}

func (bridge *MQTTBridge) buildTopic(key string, function string) string {
	parts := strings.Split(key, "/")
	if bridge.conf.Root != "" {
		parts = append(strings.Split(bridge.conf.Root, "/"), parts...)
	}

	toplevel := parts[0]
	return fmt.Sprintf("%s/%s/%s", toplevel, function, strings.Join(parts[1:], "/"))
}

func (bridge *MQTTBridge) getRoot() string {
	root := bridge.conf.Root
	if root == "" {
		root = bridge.adapter.ID()
	}
	return strings.Split(root, "/")[0]
}

func (bridge *MQTTBridge) publishBridgeStatus() error {
	root := bridge.getRoot()

	// TODO add way for adaptor to tell state
	if token := bridge.c.Publish(root+"/connected", 2, false, []byte("2")); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (bridge *MQTTBridge) publishStatus(key string, val interface{}) error {
	topic := bridge.buildTopic(key, "status")
	if b, err := json.Marshal(val); err == nil {
		fmt.Printf("MQTT publish %s %s\n", topic, string(b))
		if token := bridge.c.Publish(topic, 1, false, b); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}

	return nil
}

func (bridge *MQTTBridge) publishStatuses() error {
	if err := bridge.publishBridgeStatus(); err != nil {
		return err
	}
	return util.Traverse(bridge.adapter, bridge.publishStatus, true)
}

func (bridge *MQTTBridge) subscribeToTopics() error {
	root := bridge.getRoot()

	if token := bridge.c.Subscribe(root+"/set/#", 2, bridge.defaultHandler); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	if token := bridge.c.Subscribe(root+"/get/#", 2, bridge.defaultHandler); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	// TODO adapter command

	return nil
}

func (bridge *MQTTBridge) subscribeToAdapter() {
	go func() {
		ch := bridge.adapter.UpdateChannel()
		for u := range ch {
			for _, kvu := range u.Updates {
				bridge.publishStatus(kvu.Key, kvu.Value)
			}
		}
	}()

	return
}

func (bridge *MQTTBridge) defaultHandler(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := msg.Payload()

	fmt.Printf("MQTT received %s %s\n", topic, string(payload))

	parts := strings.Split(topic, "/")
	root := bridge.getRoot()

	if parts[0] != root {
		return
	}

	if len(parts) < 2 {
		return
	}

	var val interface{}
	json.Unmarshal(payload, &val)

	function := parts[1]
	path := parts[2:]

	path = append([]string{root}, path...)

	if bridge.conf.Root != "" {
		start := len(strings.Split(bridge.conf.Root, "/"))

		if start > len(path)-1 {
			path = []string{}
		} else {
			path = path[start:]
		}
	}

	var id, pathString string
	if len(path) > 1 {
		id = strings.Join(path[1:], "/")
	}
	if len(path) > 0 {
		pathString = strings.Join(path, "/") // path contains adapter id also
	} else {
		pathString = bridge.adapter.ID()
	}

	switch function {
	case "set":
		if err := bridge.adapter.Set(id, val); err != nil {
			fmt.Printf("Error occured while reading key %s %s\n", pathString, err.Error())
		}
	case "get":
		if val, err := bridge.adapter.Get(id); err != nil {
			fmt.Printf("Error occured while writing key %s %s\n", pathString, err.Error())
		} else {
			switch val := val.(type) {
			case adapter.ValueContainer:
				util.Traverse(val, func(key string, val interface{}) error {
					return bridge.publishStatus(pathString+"/"+key, val)
				}, false)
			default:
				bridge.publishStatus(pathString, val)
			}
		}
	}
}
