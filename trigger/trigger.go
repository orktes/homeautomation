package trigger

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/orktes/goja"
	"github.com/orktes/homeautomation/config"
	"github.com/orktes/homeautomation/util"
)

type TriggerSystem struct {
	conf config.Config
	c    mqtt.Client
	sync.Mutex

	subscriptionID int
	subscriptions  map[string]map[int]mqtt.MessageHandler
	data           map[string]interface{}

	runtimes []*runtime
}

func New(conf config.Config) *TriggerSystem {

	ts := &TriggerSystem{
		conf:          conf,
		subscriptions: map[string]map[int]mqtt.MessageHandler{},
		data:          map[string]interface{}{},
	}

	return ts
}

func (trigger *TriggerSystem) initTriggers() {
	for _, triggerConf := range trigger.conf.Triggers {
		trigger.runtimes = append(trigger.runtimes, trigger.getRuntime(triggerConf))
	}
}

func (trigger *TriggerSystem) handler(client mqtt.Client, msg mqtt.Message) {
	trigger.Lock()
	defer trigger.Unlock()

	topic := msg.Topic()
	topicParts := strings.Split(topic, "/")
	if len(topicParts) >= 2 && topicParts[1] == "status" {
		valKey := strings.Join(append([]string{topicParts[0]}, topicParts[2:]...), "/")
		var val interface{}
		json.Unmarshal(msg.Payload(), &val)
		trigger.data[valKey] = val
	}

	for subTopic, subs := range trigger.subscriptions {
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

func (trigger *TriggerSystem) subscribe(topic string, handler mqtt.MessageHandler) int {
	trigger.Lock()
	defer trigger.Unlock()

	id := trigger.subscriptionID
	trigger.subscriptionID++

	if len(trigger.subscriptions[topic]) == 0 {
		// TODO figure out qos
		trigger.c.Subscribe(topic, 1, nil)
		trigger.subscriptions[topic] = map[int]mqtt.MessageHandler{}
	}

	trigger.subscriptions[topic][id] = handler

	return id
}

func (trigger *TriggerSystem) unsubscribe(topic string, id int) {
	trigger.Lock()
	defer trigger.Unlock()

	if subs, ok := trigger.subscriptions[topic]; ok {
		delete(subs, id)
		if len(subs) == 0 {
			// TODO figure out a proper way to unsubscribe
		}
	}
}

func (trigger *TriggerSystem) getRuntime(triggerConf config.Trigger) *runtime {
	runtime := newRuntime()

	runtime.Set("get", trigger.get(runtime))
	runtime.Set("set", trigger.set(runtime))
	runtime.Set("subscribe", trigger.jsSubscribe(runtime))
	runtime.Set("unsubscribe", trigger.jsUnsubscribe(runtime))
	runtime.Set("publish", trigger.publish(runtime))
	runtime.Set("topic", trigger.topic(runtime))
	runtime.Set("sleep", trigger.sleep(runtime))
	runtime.Set("print", trigger.print(runtime))
	_, err := runtime.RunString(`
		function listen(key, cb) {
			return subscribe(topic(key, "status"), cb);
		}
		function unlisten(key, id) {
			return unsubscribe(topic(key, "status"), id);
		}
	`)
	if err != nil {
		panic(err)
	}

	_, err = runtime.RunScript("trigger.Script", triggerConf.Script)
	if err != nil {
		// TODO proper error prosessing
		panic(err)
	}

	// fmt.Printf("Eval trigger \n%s\n", triggerConf.Script)

	return runtime
}

func (trigger *TriggerSystem) get(r *runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()

	getVal:
		trigger.Lock()
		val, ok := trigger.data[key]
		trigger.Unlock()

		if !ok {
			ch := make(chan struct{})

			statusTopic := util.ConvertValueToTopic(key, "status")
			id := trigger.subscribe(statusTopic, func(client mqtt.Client, msg mqtt.Message) {
				ch <- struct{}{}
			})

			// TODO figure out right qos and retain
			trigger.c.Publish(util.ConvertValueToTopic(key, "get"), 0, false, []byte{})

			<-ch // TODO timeout etc

			trigger.unsubscribe(statusTopic, id)

			goto getVal
		}

		return r.ToValue(val)
	}

}

func (trigger *TriggerSystem) set(r *runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		val := call.Argument(1).Export()

		topic := util.ConvertValueToTopic(key, "set")

		if b, err := json.Marshal(val); err == nil {
			if token := trigger.c.Publish(topic, 1, false, b); token.Wait() && token.Error() != nil {
				panic(token.Error())
			}
		}

		return goja.Undefined()
	}

}
func (trigger *TriggerSystem) jsSubscribe(r *runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		topic := call.Argument(0).String()
		if fn, ok := goja.AssertFunction(call.Argument(1)); ok {
			id := trigger.subscribe(topic, func(client mqtt.Client, msg mqtt.Message) {
				go r.Work(func(r *runtime) {
					defer func() {
						err := recover()
						if err != nil {
							fmt.Printf("Error: %s\n", err)
						}
					}()

					fn(nil, r.ToValue(msg.Topic()), r.ToValue(string(msg.Payload())))
				})
			})

			return r.ToValue(id)
		}

		panic("Invalid arguments passed to listen")
	}

}

func (trigger *TriggerSystem) jsUnsubscribe(r *runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		topic := call.Argument(0).String()
		id := call.Argument(0).ToInteger()
		trigger.unsubscribe(topic, int(id))

		return goja.Undefined()
	}

}

func (trigger *TriggerSystem) publish(r *runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		topic := call.Argument(0).String()
		qos := byte(call.Argument(1).ToInteger())
		retained := call.Argument(2).ToBoolean()
		payload := call.Argument(3).String()

		trigger.c.Publish(topic, qos, retained, []byte(payload))

		return goja.Undefined()
	}
}

func (trigger *TriggerSystem) topic(r *runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		return r.ToValue(util.ConvertValueToTopic(call.Argument(0).String(), call.Argument(1).String()))
	}

}
func (trigger *TriggerSystem) sleep(r *runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		timeInMS := call.Argument(0).ToInteger()
		time.Sleep(time.Duration(timeInMS) * time.Millisecond)
		return goja.Undefined()
	}

}

func (trigger *TriggerSystem) print(r *runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		// TODO proper indicator on which of the riggers this is

		strs := make([]string, len(call.Arguments))

		for i, arg := range call.Arguments {
			strs[i] = arg.String()
		}

		fmt.Printf("[TRIGGER]: %s\n", strings.Join(strs, " "))

		return goja.Undefined()
	}

}

func (trigger *TriggerSystem) Connect() error {
	conf := trigger.conf
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

	opts = opts.SetDefaultPublishHandler(trigger.handler)
	c := mqtt.NewClient(opts)

	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	trigger.c = c

	trigger.initTriggers()

	return nil
}

func (trigger *TriggerSystem) Disconnect(wait uint) error {
	trigger.c.Disconnect(wait)
	return nil
}
