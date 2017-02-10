package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"

	crand "crypto/rand"

	"math/rand"

	"github.com/dollarshaveclub/talcum/src/talcum"
	"github.com/hashicorp/consul/api"
)

func selectRandom(selectorConfig talcum.SelectorConfig, config *talcum.Config) {
	selector := talcum.NewSelector(config, selectorConfig, nil)
	entry := selector.SelectRandom()
	fmt.Println(entry.RoleName)
}

func main() {
	log.SetOutput(os.Stderr)

	seed, err := crand.Int(crand.Reader, big.NewInt(100000))
	if err != nil {
		log.Printf("Error generating random seed: %s", err)
	}
	rand.Seed(seed.Int64())

	var selectorConfigPath string
	var config talcum.Config
	var consulHost string

	flag.StringVar(&selectorConfigPath, "config-path", "", "the path to the role configuration file")
	flag.StringVar(&consulHost, "consul-host", "localhost:8500", "the location of Consul")
	flag.StringVar(&config.ApplicationName, "app-name", "app", "the name of the current application")
	flag.StringVar(&config.SelectionID, "selection-id", "1", "the ID of the current selection")
	flag.BoolVar(&config.DebugMode, "debug", false, "run in debug mode")
	flag.DurationVar(&config.LockDelay, "lock-delay", 0, "the delay in between lock attempts")
	flag.Parse()

	var selectorConfig talcum.SelectorConfig
	f, err := os.Open(selectorConfigPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&selectorConfig); err != nil {
		panic(err)
	}

	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulHost
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Printf("Error creating Consul client: %s", err)
		log.Printf("Selecting random entry")
		selectRandom(selectorConfig, &config)
		return
	}
	kvClient := consulClient.KV()
	locker := talcum.NewConsulLocker(kvClient)

	selector := talcum.NewSelector(&config, selectorConfig, locker)
	entry, err := selector.Select()
	if err != nil {
		log.Printf("Error selecting an entry: %s", err)
		log.Printf("Selecting random entry")
		selectRandom(selectorConfig, &config)
		return
	}

	fmt.Println(entry.RoleName)
}
