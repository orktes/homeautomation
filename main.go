package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/orktes/homeautomation/adapter/deconz"
	"github.com/orktes/homeautomation/adapter/dra"
	"github.com/orktes/homeautomation/config"
	"github.com/orktes/homeautomation/hub"
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
		switch adapterConfig.Type {
		case "deconz":
			ad, err := deconz.Create(adapterConfig.ID, adapterConfig.Config, h)
			if err != nil {
				panic(err)
			}
			h.AddAdapter(adapterConfig.ID, ad)
		case "dra":
			ad, err := dra.Create(adapterConfig.ID, adapterConfig.Config, h)
			if err != nil {
				panic(err)
			}
			h.AddAdapter(adapterConfig.ID, ad)
		}
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	<-c

	h.Close()
}
