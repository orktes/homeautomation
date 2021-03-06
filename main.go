package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/orktes/homeautomation/alexa"
	"github.com/orktes/homeautomation/bridge/adapter"
	"github.com/orktes/homeautomation/trigger"

	"github.com/orktes/homeautomation/config"

	// Adapters
	"github.com/orktes/homeautomation/bridge/adapter/bolt"
	"github.com/orktes/homeautomation/bridge/adapter/deconz"
	"github.com/orktes/homeautomation/bridge/adapter/dra"
	"github.com/orktes/homeautomation/bridge/adapter/viera"

	"github.com/orktes/homeautomation/bridge/mqtt"
)

var NoopCloser = func() error { return nil }

func configureBridge(conf config.Config) func() error {
	bridgeConf := conf.Bridge
	if bridgeConf == nil {
		return NoopCloser
	}

	if len(bridgeConf.Adapters) > 1 && bridgeConf.Root == "" {
		fmt.Println("root path must be defined when defining multiple adapters")
		os.Exit(1)
		return NoopCloser
	}

	adapters := make([]adapter.Adapter, 0, len(bridgeConf.Adapters))

	for _, adapterConf := range bridgeConf.Adapters {
		var createFunc func(id string, config map[string]interface{}) (adapter.Adapter, error)
		switch adapterConf.Type {
		case "deconz":
			createFunc = deconz.Create
		case "dra":
			createFunc = dra.Create
		case "viera":
			createFunc = viera.Create
		case "bolt":
			createFunc = bolt.Create
		default:
			fmt.Printf("No such adapter %s\n", adapterConf.Type)
			os.Exit(1)
			return NoopCloser
		}

		adapter, err := createFunc(adapterConf.ID, adapterConf.Config)
		if err != nil {
			fmt.Printf("Error creating adapter %s: %s\n", adapterConf.Type, err.Error())
			os.Exit(1)
		}

		adapters = append(adapters, adapter)
	}

	var mainAdapter adapter.Adapter
	if len(adapters) == 0 {
		mainAdapter = adapters[0]
	} else {
		mainAdapter = adapter.NewMultiAdapter(bridgeConf.Root, adapters...)
		// Root key now comes from multi adapter
		bridgeConf.Root = ""
	}

	mqttBridge := mqtt.New(conf, mainAdapter)

	if err := mqttBridge.Connect(); err != nil {
		fmt.Printf("Error connecting to mqtt brokers %s\n", err.Error())
		os.Exit(1)
		return NoopCloser
	}

	return func() error {
		if err := mqttBridge.Disconnect(0); err != nil {
			fmt.Printf("Error disconnecting from mqtt brokers %s\n", err.Error())
			return err
		}

		return nil
	}

}

func configureTriggerSystem(conf config.Config) func() error {
	if len(conf.Triggers) == 0 {
		return NoopCloser
	}

	ts := trigger.New(conf)
	if err := ts.Connect(); err != nil {
		fmt.Printf("Error connecting to mqtt brokers %s\n", err.Error())
		os.Exit(1)
		return NoopCloser
	}
	return func() error {
		if err := ts.Disconnect(0); err != nil {
			fmt.Printf("Error disconnecting from mqtt brokers %s\n", err.Error())
			return err
		}

		return nil
	}
}

func configureAlexa(conf config.Config) func() error {
	if conf.Alexa == nil {
		return NoopCloser
	}

	a := alexa.New(conf)
	if err := a.Connect(); err != nil {
		fmt.Printf("Error connecting to mqtt brokers %s\n", err.Error())
		os.Exit(1)
		return NoopCloser
	}

	return func() error {
		if err := a.Disconnect(0); err != nil {
			fmt.Printf("Error disconnecting from mqtt brokers %s\n", err.Error())
			return err
		}

		return nil
	}
}

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(file)
	conf, err := config.ParseConfig(reader)
	if err != nil {
		panic(err)
	}

	closeBridge := configureBridge(conf)
	closeTriggerSystem := configureTriggerSystem(conf)
	closeAlexa := configureAlexa(conf)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	<-c

	closeBridge()
	closeTriggerSystem()
	closeAlexa()

}
