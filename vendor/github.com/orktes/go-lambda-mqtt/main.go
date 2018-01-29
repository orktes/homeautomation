package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/orktes/go-lambda-mqtt/structs"
)

const timeout = time.Second * 5

var errorTimeout = errors.New("MQTT response not received during specified time frame")
var outTopic string = os.Getenv("MQTT_OUT_TOPIC")
var client mqtt.Client
var connected bool

func handler(in json.RawMessage) (out json.RawMessage, err error) {
	if outTopic == "" {
		panic("Output topic should be defined")
	}

	fmt.Printf("[REQUEST] %s\n", string(in))
	defer func() {
		if err != nil {
			fmt.Printf("[RESPONSE] Error: %s\n", err)
			return
		}
		fmt.Printf("[RESPONSE] %s\n", string(out))
	}()

	if !connected {
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			err = token.Error()
			return
		}
		connected = true
	}

	reqID := uuid.New()

	ch := make(chan json.RawMessage)
	errChan := make(chan error)

	inTopic := fmt.Sprintf("%s/response/%s", outTopic, reqID)
	fmt.Printf("[MQTT-SUBSCRIBE] %s\n", inTopic)
	token := client.Subscribe(inTopic, 2, func(client mqtt.Client, message mqtt.Message) {
		payload := message.Payload()
		ch <- payload
	})

	if token.Wait() && token.Error() != nil {
		err = token.Error()
		return
	}
	defer client.Unsubscribe(inTopic)

	req := structs.Request{
		ID:      reqID.String(),
		Topic:   inTopic,
		Payload: in,
	}

	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[MQTT-PUBLISH] %s\n", outTopic)
	if token := client.Publish(outTopic, 1, false, b); token.Wait() && token.Error() != nil {
		err = token.Error()
		return
	}

	select {
	case err = <-errChan:
	case out = <-ch:
	case <-time.After(timeout):
		err = errorTimeout
	}

	return
}

func init() {
	clientID := os.Getenv("MQTT_CLIENT_ID")
	broker := os.Getenv("MQTT_BROKER")
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")

	opts := mqtt.NewClientOptions()
	opts = opts.AddBroker(broker)
	if clientID != "" {
		opts = opts.SetClientID(clientID)
	}
	if username != "" {
		opts = opts.SetUsername(username)
	}
	if password != "" {
		opts = opts.SetPassword(password)
	}
	client = mqtt.NewClient(opts)
}

func main() {
	lambda.Start(handler)
}
