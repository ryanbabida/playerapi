package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type config struct {
	InMemory        bool     `json:"inMemory"`
	DbFilePath      string   `json:"dbFilePath"`
	Port            string   `json:"port"`
	CbsApiUrl       string   `json:"cbsApiUrl"`
	Sports          []string `json:"sports"`
	ImportBatchSize int      `json:"importBatchSize"`
	dbSourceFile    string
}

func NewConfig(filePath string) (*config, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %v", err)
	}

	var config config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %v", err)
	}

	// TODO: overrides here

	config.dbSourceFile = config.DbFilePath
	if config.InMemory {
		config.dbSourceFile = ":memory:"
	}

	return &config, nil
}
