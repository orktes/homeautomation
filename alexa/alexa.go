package alexa

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	smarthome "github.com/orktes/go-alexa-smarthome"
	"github.com/orktes/go-lambda-mqtt/structs"
	"github.com/orktes/goja"
	"github.com/orktes/homeautomation/config"
	"github.com/orktes/homeautomation/util"
)

type Alexa struct {
	conf      config.Config
	smarthome *smarthome.Smarthome
	c         mqtt.Client

	subscriptionID int
	subscriptions  map[string]map[int]mqtt.MessageHandler
	data           map[string]interface{}

	runtime      *goja.Runtime
	runtimeMutex sync.Mutex

	sync.Mutex
}

func New(conf config.Config) *Alexa {
	a := &Alexa{
		conf:          conf,
		subscriptions: map[string]map[int]mqtt.MessageHandler{},
		data:          map[string]interface{}{},
		runtime:       goja.New(),
	}

	a.runtime.Set("get", a.get)
	a.runtime.Set("set", a.set)

	sm := smarthome.New(smarthome.AuthorizationFunc(a.auth))

	for _, device := range conf.Alexa.Devices {
		id := device.ID
		name := device.Name
		manafacturerName := device.Manafacturer
		description := device.Description

		if manafacturerName == "" {
			manafacturerName = "Homeautomation"
		}

		dev := smarthome.NewAbstractDevice(
			id,
			name,
			manafacturerName,
			description,
		)

		for _, category := range device.DisplayCategories {
			dev.AddDisplayCategory(category)
		}

		for _, capabilityConfig := range device.Capabilities {
			capability := dev.NewCapability(capabilityConfig.Interface)

			for _, conf := range capabilityConfig.Properties {
				capability.AddPropertyHandler(conf.Name, a.getMQTTPropertyHandler(conf))
			}

			for _, conf := range capabilityConfig.Actions {
				capability.AddAction(conf.Name, a.getMQTTActionHandler(conf))
			}
		}

		sm.AddDevice(dev)

	}

	a.smarthome = sm
	return a
}

func (a *Alexa) exec(str string, context map[string]interface{}) (val goja.Value, err error) {
	a.runtimeMutex.Lock()
	defer a.runtimeMutex.Unlock()

	contextData := []byte("{}")

	if context != nil {
		contextData, err = json.Marshal(context)
		if err != nil {
			return nil, err
		}
	}

	script := fmt.Sprintf(`
		with(%s) {
			%s
		}
	`, string(contextData), str)

	return a.runtime.RunString(script)
}

func (a *Alexa) getMQTTPropertyHandler(conf config.AlexaDeviceCapabilityProperty) smarthome.PropertyHandler {
	return &propertyHandler{alexa: a, conf: conf}
}

func (a *Alexa) getMQTTActionHandler(conf config.AlexaDeviceCapabilityAction) func(interface{}) (interface{}, error) {
	return func(arg interface{}) (interface{}, error) {
		script := conf.Script
		gval, err := a.exec(script, map[string]interface{}{"value": arg})
		if err != nil {
			return nil, err
		}

		val := gval.Export()

		switch v := val.(type) {
		case int64:
			val = float64(v)
		case int:
			val = float64(v)
		}
		if v, ok := val.(float64); ok {
			if rangeVal, err := util.ConvertFloatValueToRange(conf.OutputRange, conf.InputRange, v); err == nil {
				val = rangeVal
				if conf.Type == "int" {
					val = int(rangeVal + math.Copysign(0.5, rangeVal))
				}
			}
		}

		return val, nil
	}
}

func (a *Alexa) auth(req smarthome.AcceptGrantRequest) error {
	// TODO
	return nil
}

func (a *Alexa) handler(client mqtt.Client, msg mqtt.Message) {
	a.Lock()
	defer a.Unlock()

	topic := msg.Topic()
	topicParts := strings.Split(topic, "/")
	if len(topicParts) >= 2 && topicParts[1] == "status" {
		valKey := strings.Join(append([]string{topicParts[0]}, topicParts[2:]...), "/")
		var val interface{}
		json.Unmarshal(msg.Payload(), &val)
		a.data[valKey] = val
	}

	for subTopic, subs := range a.subscriptions {
		if subTopic != topic {
			if strings.HasSuffix(subTopic, "#") {
				if !strings.HasPrefix(topic, subTopic[:len(subTopic)-1]) {
					continue
				}
			} else {
				continue
			}
		}
		for _, sub := range subs {
			go sub(client, msg)
		}
	}
}

func (a *Alexa) subscribe(topic string, handler mqtt.MessageHandler) int {
	a.Lock()
	defer a.Unlock()

	id := a.subscriptionID
	a.subscriptionID++

	if len(a.subscriptions[topic]) == 0 {
		if token := a.c.Subscribe(topic, 1, a.handler); token.Wait() && token.Error() != nil {
			// TODO handle in a proper way
			panic(token.Error())
		}
		a.subscriptions[topic] = map[int]mqtt.MessageHandler{}
	}

	a.subscriptions[topic][id] = handler

	return id
}

func (a *Alexa) unsubscribe(topic string, id int) {
	a.Lock()
	defer a.Unlock()

	if subs, ok := a.subscriptions[topic]; ok {
		delete(subs, id)
		if len(subs) == 0 {
			// TODO figure out a proper way to unsubscribe
		}
	}
}

func (a *Alexa) get(call goja.FunctionCall) goja.Value {
	key := call.Argument(0).String()

getVal:
	a.Lock()
	val, ok := a.data[key]
	a.Unlock()

	if !ok {
		ch := make(chan struct{})

		statusTopic := util.ConvertValueToTopic(key, "status")
		id := a.subscribe(statusTopic, func(client mqtt.Client, msg mqtt.Message) {
			ch <- struct{}{}
		})
		defer a.unsubscribe(statusTopic, id)

		// TODO figure out right qos and retain
		if token := a.c.Publish(util.ConvertValueToTopic(key, "get"), 0, false, []byte{}); token.Wait() && token.Error() != nil {
			// TODO handle error
		}

		<-ch // TODO timeout etc

		goto getVal
	}

	return a.runtime.ToValue(val)

}

func (a *Alexa) set(call goja.FunctionCall) goja.Value {
	key := call.Argument(0).String()
	val := call.Argument(1).Export()

	topic := util.ConvertValueToTopic(key, "set")

	if b, err := json.Marshal(val); err == nil {
		if token := a.c.Publish(topic, 1, false, b); token.Wait() && token.Error() != nil {
			// TODO handle error
		}
	}

	return goja.Undefined()
}

func (a *Alexa) handleLambdaMessage(client mqtt.Client, msg mqtt.Message) {
	req := &structs.Request{}
	err := json.Unmarshal(msg.Payload(), req)
	if err != nil {
		// TODO log error
		return
	}

	if req.Topic == "" {
		return
	}

	alexaReq := &smarthome.Request{}
	err = json.Unmarshal(req.Payload, alexaReq)
	if err != nil {
		// TODO log error
		return
	}

	go func() {
		res := a.smarthome.Handle(alexaReq)

		resb, err := json.Marshal(res)
		if err != nil {
			// TODO log error
			return
		}

		if token := a.c.Publish(req.Topic, 2, false, resb); token.Wait() && token.Error() != nil {
			return
		}

	}()
}

func (a *Alexa) subscribeToLamdaTopic() error {
	c := a.c
	if token := c.Subscribe(a.conf.Alexa.Topic, 2, a.handleLambdaMessage); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (a *Alexa) Connect() error {
	conf := a.conf
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

	opts = opts.SetDefaultPublishHandler(a.handler)

	c := mqtt.NewClient(opts)

	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	a.c = c

	a.subscribeToLamdaTopic()

	return nil
}

func (a *Alexa) Disconnect(wait uint) error {
	a.c.Disconnect(wait)
	return nil
}
