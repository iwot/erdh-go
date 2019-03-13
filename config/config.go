package config

import (
	"io/ioutil"
	"strings"

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
	SourceFrom string       `yaml:"source_from"`
	Group      []string     `yaml:"group"`
	Im         Intermediate `yaml:"intermediate,omitempty"`
	ExInfo     string       `yaml:"ex_info"`
}

// IsDBSource はソースがDBであればtrueを返す
func (c Config) IsDBSource() bool {
	test := strings.ToLower(c.Source)
	if test == "mysql" || test == "sqlite" {
		return true
	}
	return false
}

// IsYAMLSource はソースがYAML（中間形式ファイル）であればtrueを返す
func (c Config) IsYAMLSource() bool {
	test := strings.ToLower(c.Source)
	if test == "yaml" {
		return true
	}
	return false
}

// Intermediate は出力する中間形式ファイルのパスの定義
type Intermediate struct {
	SaveTo string `yaml:"save_to,omitempty"`
}

// NewConfigFromYamlFile はコンフィグをファイルから読む
func NewConfigFromYamlFile(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return NewConfigFromYaml(buf)
}

// NewConfigFromYaml はコンフィグをバッファから読む
func NewConfigFromYaml(buf []byte) (*Config, error) {
	var result = new(Config)

	err := yaml.Unmarshal(buf, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
