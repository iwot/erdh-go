package config

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type DBConfig struct {
	DBType   string `yaml:"dbtype"`
	Host     string `yaml:"host,omitempty"`
	Port     string `yaml:"port,omitempty"`
	DBName   string `yaml:"dbname"`
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
}

func (c DBConfig) ToDSN() (string, error) {
	var b strings.Builder

	if c.DBType == "mysql" {
		fmt.Fprint(&b, c.User)
		if len(c.Password) > 0 {
			fmt.Fprint(&b, ":")
			fmt.Fprint(&b, c.Password)
		}
		fmt.Fprint(&b, "@")
		if len(c.Host) > 0 {
			fmt.Fprint(&b, "tcp(")
			fmt.Fprint(&b, c.Host)
			if len(c.Port) > 0 {
				fmt.Fprint(&b, ":")
				fmt.Fprint(&b, c.Port)
			}
			fmt.Fprint(&b, ")")
		}
		fmt.Fprint(&b, "/")
		fmt.Fprint(&b, c.DBName)
	}

	return b.String(), nil
}

func NewDBConfigFromYamlFile(path string) (*DBConfig, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return NewDBConfigFromYaml(buf)
}

func NewDBConfigFromYaml(buf []byte) (*DBConfig, error) {
	var result = new(DBConfig)

	err := yaml.Unmarshal(buf, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
