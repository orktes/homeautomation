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
		ad, err := registry.Create(adapterConfig, h)
		if err != nil {
			fmt.Printf("Could not init %s (%s)\n", adapterConfig.ID, adapterConfig.Type)
			panic(err)
		}
		h.AddAdapter(adapterConfig.ID, ad)
	}

	for _, trigger := range conf.Trigger {
		h.CreateTrigger(trigger)
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

	/*
		val, err := h.RunScript(`
			tv[1].power = on
		`)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%+v\n", val)
	*/

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	<-c

	h.Close()
}
