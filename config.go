package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// GameClient data about a game client being watched for updates
type GameClient struct {
	Name         string `json:"name"`
	BlizztrackID string `json:"blizztrackId"`
	Version      string `json:"version"`
}

// Config data loaded from the configuration file
type Config struct {
	ServiceOn    bool         `json:"serviceOn"`
	FailureCount int          `json:"failureCount"`
	MaxFailures  int          `json:"maxFailures"`
	POAppToken   string       `json:"pushoverAppToken"`
	POUserToken  string       `json:"pushoverUserToken"`
	PODevice     string       `json:"pushoverDevice"`
	GameClients  []GameClient `json:"gameClients"`
}

// WriteToFile writes the input Config struct to the config.json file
func (c *Config) WriteToFile() error {
	config, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("./config.json", config, 0666)
	if err != nil {
		return err
	}
	return nil
}

// GetConfig gets the contents of the config.json file and returns a Config
// struct containing config data
func GetConfig() (*Config, error) {
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Printf("ERROR %v", err)
		return nil, err
	}
	var conf Config
	err = json.Unmarshal(file, &conf)
	if err != nil {
		log.Printf("ERROR %v", err)
		return nil, err
	}
	return &conf, nil
}
