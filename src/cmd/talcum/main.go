package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	crand "crypto/rand"

	"math/rand"

	"github.com/dollarshaveclub/talcum/src/talcum"
	"github.com/hashicorp/consul/api"
)

func selectRandom(selectorConfig talcum.SelectorConfig, config *talcum.Config) *talcum.SelectorEntry {
	selector := talcum.NewSelector(config, selectorConfig, nil)
	return selector.SelectRandom()
}

func main() {
	logger := log.New(os.Stderr, "", log.LstdFlags)

	seed, err := crand.Int(crand.Reader, big.NewInt(100000))
	if err != nil {
		logger.Printf("Error generating random seed: %s", err)
	}
	rand.Seed(seed.Int64())

	var selectorConfigConsulPath string
	var selectorConfigPath string
	var config talcum.Config
	var mconfig talcum.MetricsConfig
	var consulHost string

	flag.StringVar(&selectorConfigConsulPath, "consul-path", "", "the path to the role configuration in Consul")
	flag.StringVar(&selectorConfigPath, "config-path", "", "the path to the role configuration file")
	flag.StringVar(&consulHost, "consul-host", "localhost:8500", "the location of Consul")
	flag.StringVar(&config.ApplicationName, "app-name", "app", "the name of the current application")
	flag.StringVar(&config.SelectionID, "selection-id", "1", "the ID of the current selection")
	flag.BoolVar(&config.DebugMode, "debug", false, "run in debug mode")
	flag.DurationVar(&config.LockDelay, "lock-delay", 0, "the delay in between lock attempts")
	flag.StringVar(&mconfig.StatsdAddr, "statsd-addr", "0.0.0.0:8125", "statsd (dogstatsd) address")
	flag.BoolVar(&mconfig.Datadog, "datadog", true, "statsd is Datadog (dogstatsd)")
	flag.StringVar(&mconfig.Namespace, "metrics-namespace", "talcum", "Datadog metrics namespace (ignored if not using Datadog)")
	flag.StringVar(&mconfig.TagStr, "metrics-tags", "production", "Metrics tags (comma-delimited, either datadog <key>:<value> or influxdb <key>=<value>")
	flag.Parse()

	if mconfig.TagStr != "" {
		mconfig.Tags = strings.Split(mconfig.TagStr, ",")
	}

	mc, err := talcum.NewStatsdCollector(&mconfig, logger)
	if err != nil {
		logger.Printf("error initializing datadog collector: %v", err)
	}

	clierr := func(msg string, params ...interface{}) {
		mc.RoleError()
		mc.Flush()
		fmt.Fprintf(os.Stderr, msg+"\n", params...)
		os.Exit(1)
	}

	defer mc.Flush()
	defer mc.TimeToPickRole(time.Now().UTC())

	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulHost
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		clierr("consul error: %v", err)
	}
	kvClient := consulClient.KV()
	locker := talcum.NewConsulLocker(kvClient)

	var selectorConfig talcum.SelectorConfig
	if selectorConfigPath != "" {
		f, err := os.Open(selectorConfigPath)
		if err != nil {
			clierr("error opening config: %v", err)
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&selectorConfig); err != nil {
			clierr("error unmarshaling config: %v", err)
		}
	} else if selectorConfigConsulPath != "" {
		kvPair, _, err := kvClient.Get(selectorConfigConsulPath, nil)
		if err != nil {
			clierr("error reading consul KV: %v", err)
		}

		if kvPair != nil {
			if err := json.Unmarshal(kvPair.Value, &selectorConfig); err != nil {
				clierr("error unmarshaling config: %v", err)
			}
		} else {
			clierr("kv pair equal to nil")
		}

	} else {
		clierr("Selector config not provided")
	}

	var entry *talcum.SelectorEntry
	selector := talcum.NewSelector(&config, selectorConfig, locker)
	entry, err = selector.Select()
	if err != nil {
		logger.Printf("Error selecting an entry: %s", err)
		logger.Printf("Selecting random entry")
		entry = selectRandom(selectorConfig, &config)
		mc.RandomRoleChosen()
	}

	mc.RoleChosen(entry.RoleName)
	logger.Printf("role: %v", entry.RoleName)
	fmt.Println(entry.RoleDefinition)
}
