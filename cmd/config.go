package cmd

import (
	"time"

	"github.com/andig/evcc/api"
	"github.com/andig/evcc/charger"
	"github.com/andig/evcc/meter"
	"github.com/andig/evcc/provider"
	"github.com/andig/evcc/push"
	"github.com/andig/evcc/server"
	"github.com/andig/evcc/vehicle"
)

type config struct {
	URI        string
	Log        string
	Levels     map[string]string
	Interval   time.Duration
	Mqtt       provider.MqttConfig
	Influx     server.InfluxConfig
	Menu       []server.MenuConfig
	Messaging  messagingConfig
	Meters     []qualifiedConfig
	Chargers   []qualifiedConfig
	Vehicles   []qualifiedConfig
	Site       map[string]interface{}
	LoadPoints []map[string]interface{}
}

type qualifiedConfig struct {
	Name, Type string
	Other      map[string]interface{} `mapstructure:",remain"`
}

type typedConfig struct {
	Type  string
	Other map[string]interface{} `mapstructure:",remain"`
}

type messagingConfig struct {
	Events   map[string]push.EventTemplate
	Services []typedConfig
}

// ConfigProvider provides configuration items
type ConfigProvider struct {
	meters   map[string]api.Meter
	chargers map[string]api.Charger
	vehicles map[string]api.Vehicle
}

// Meter provides meters by name
func (cp *ConfigProvider) Meter(name string) api.Meter {
	if meter, ok := cp.meters[name]; ok {
		return meter
	}
	log.FATAL.Fatalf("invalid meter: %s", name)
	return nil
}

// Charger provides chargers by name
func (cp *ConfigProvider) Charger(name string) api.Charger {
	if charger, ok := cp.chargers[name]; ok {
		return charger
	}
	log.FATAL.Fatalf("invalid charger: %s", name)
	return nil
}

// Vehicle provides vehicles by name
func (cp *ConfigProvider) Vehicle(name string) api.Vehicle {
	if vehicle, ok := cp.vehicles[name]; ok {
		return vehicle
	}
	log.FATAL.Fatalf("invalid vehicle: %s", name)
	return nil
}

func (cp *ConfigProvider) configure(conf config) {
	cp.configureMeters(conf)
	cp.configureChargers(conf)
	cp.configureVehicles(conf)
}

func (cp *ConfigProvider) configureMeters(conf config) {
	cp.meters = make(map[string]api.Meter)
	for _, cc := range conf.Meters {
		m, err := meter.NewFromConfig(cc.Type, cc.Other)
		if err != nil {
			log.FATAL.Fatal(err)
		}

		if _, exists := cp.meters[cc.Name]; exists {
			log.FATAL.Fatalf("duplicate meter name: %s already defined and must be unique", cc.Name)
		}

		cp.meters[cc.Name] = m
	}
}

func (cp *ConfigProvider) configureChargers(conf config) {
	cp.chargers = make(map[string]api.Charger)
	for _, cc := range conf.Chargers {
		c, err := charger.NewFromConfig(cc.Type, cc.Other)
		if err != nil {
			log.FATAL.Fatal(err)
		}

		if _, exists := cp.chargers[cc.Name]; exists {
			log.FATAL.Fatalf("duplicate charger name: %s already defined and must be unique", cc.Name)
		}

		cp.chargers[cc.Name] = c
	}
}

func (cp *ConfigProvider) configureVehicles(conf config) {
	cp.vehicles = make(map[string]api.Vehicle)
	for _, cc := range conf.Vehicles {
		v, err := vehicle.NewFromConfig(cc.Type, cc.Other)
		if err != nil {
			log.FATAL.Fatal(err)
		}

		if _, exists := cp.vehicles[cc.Name]; exists {
			log.FATAL.Fatalf("duplicate vehicle name: %s already defined and must be unique", cc.Name)
		}

		cp.vehicles[cc.Name] = v
	}
}
