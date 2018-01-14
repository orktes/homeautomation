package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/orktes/homeautomation/config"
	"github.com/orktes/homeautomation/hub"
	"github.com/orktes/homeautomation/registry"

	// Adapters
	_ "github.com/orktes/homeautomation/adapter/deconz"
	_ "github.com/orktes/homeautomation/adapter/dra"
	_ "github.com/orktes/homeautomation/adapter/viera"

	// Frontends
	_ "github.com/orktes/homeautomation/frontend/ssh"
)

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

	h := hub.New()
	for _, adapterConfig := range conf.Adapters {
		ad, err := registry.CreateAdapter(adapterConfig, h)
		if err != nil {
			fmt.Printf("Could not init %s (%s)\n", adapterConfig.ID, adapterConfig.Type)
			panic(err)
		}
		h.AddAdapter(adapterConfig.ID, ad)
	}

	for _, lightConfig := range conf.Lights {
		h.CreateLight(lightConfig)
	}

	for _, trigger := range conf.Trigger {
		h.CreateTrigger(trigger)
	}

	for _, frontendConf := range conf.Frontends {
		_, err := registry.CreateFrontend(frontendConf, h)
		if err != nil {
			fmt.Printf("Could not init %s (%s)\n", frontendConf.ID, frontendConf.Type)
			panic(err)
		}
	}

	fmt.Printf("Configuration done!\n")

	go func() {
		ch := h.UpdateChannel()
		for u := range ch {
			for _, ukv := range u.Updates {
				fmt.Printf("Update %s = %v\n", ukv.Key, ukv.Value)
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	<-c

	h.Close()
}
