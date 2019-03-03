package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type ExtraConfig struct {
	Tables []Table `yaml:"tables"`
}

type Table struct {
	Name      string       `yaml:"table"`
	IsMaster  bool         `yaml:"is_master"`
	Group     string       `yaml:"group"`
	Relations []ExRelation `yaml:"relations"`
}

type ExRelation struct {
	ReferencedTableName string           `yaml:"referenced_table_name"`
	Columns             []ColumnRelation `yaml:"columns"`
	ThisConnection      string           `yaml:"this_conn"`
	ThatConnection      string           `yaml:"that_conn"`
}

type ColumnRelation struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

func NewExtraConfigFromYamlFile(path string) (*ExtraConfig, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return NewExtraConfigFromYaml(buf)
}

func NewExtraConfigFromYaml(buf []byte) (*ExtraConfig, error) {
	var result = new(ExtraConfig)

	err := yaml.Unmarshal(buf, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
