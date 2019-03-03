package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// const (
// 	MySQL = iota
// 	PostgreSQL
// 	SQLite
// 	YAML
// )

type Config struct {
	Source     string       `yaml:"source"`
	SourceFrom string       `yaml:"source-from"`
	Group      []string     `yaml:"group"`
	Im         Intermediate `yaml:"intermediate"`
	ExInfo     string       `yaml:"ex-info"`
}

type Intermediate struct {
	SaveTo string `yaml:"save-to"`
}

func NewConfigFromYamlFile(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return NewConfigFromYaml(buf)
}

func NewConfigFromYaml(buf []byte) (*Config, error) {
	var result = new(Config)

	err := yaml.Unmarshal(buf, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
