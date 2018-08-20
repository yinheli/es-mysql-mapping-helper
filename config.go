package main

import (
	"github.com/gobwas/glob"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Database   *Database
	Index      *IndexSetting
	Tables     []string
	TablesGlob []glob.Glob
	Rules      []*Rule
}

type Database struct {
	Uri string `yaml:"uri"`
}

type IndexSetting struct {
	Prefix   string `yaml:"prefix"`
	Shards   int    `yaml:"shards"`
	Replicas int    `yaml:"replicas"`
}

type Rule struct {
	Table             string           `yaml:"table"`
	Index             string           `yaml:"index"`
	Shards            int              `yaml:"shards"`
	Replicas          int              `yaml:"replicas"`
	SearchableColumns []string         `yaml:"searchableColumns"`
	Columns           []*ColumnSetting `yaml:"columns"`
}

type ColumnSetting struct {
	Name           string `yaml:"name"`
	Type           string `yaml:"type"`
	Analyzer       string `yaml:"analyzer"`
	SearchAnalyzer string `yaml:"search_analyzer"`
}

func loadConfig(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		return nil, err
	}

	var tableGlobs []glob.Glob
	for _, v := range config.Tables {
		g, err := glob.Compile(v, '.')
		if err != nil {
			return nil, err
		}
		tableGlobs = append(tableGlobs, g)
	}
	config.TablesGlob = tableGlobs

	return &config, nil
}
