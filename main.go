package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/iwot/erdh-go/config"
	"github.com/iwot/erdh-go/db"
	"github.com/iwot/erdh-go/erdh"
	"gopkg.in/yaml.v2"
)

func main() {
	var (
		c = flag.String("config", "", "config yaml file path")
		o = flag.String("out", "", "output puml file path")
	)
	flag.Parse()
	fmt.Println("read from", *c)
	fmt.Println("output to", *o)

	conf, err := config.NewConfigFromYamlFile(*c)
	if err != nil {
		panic(err)
	}

	fmt.Println("conf.SourceFrom", conf.SourceFrom)
	dbConf, err := config.NewDBConfigFromYamlFile(conf.SourceFrom)
	if err != nil {
		panic(err)
	}

	exInfo, err := config.NewExtraConfigFromYamlFile(conf.ExInfo)
	if err != nil {
		panic(err)
	}

	cons, err := db.ReadDB(conf.Source, *dbConf)
	if err != nil {
		panic(err)
	}

	cons.UpdateExRelationsFromForeignKeys()

	cons.ApplyExInfo(*exInfo)

	// 中間形式ファイルを保存
	if len(conf.Im.SaveTo) > 0 {
		fmt.Println("intermediate yaml saving to", conf.Im.SaveTo)
		d, err := yaml.Marshal(&cons)
		if err != nil {
			panic(err)
		}
		file, _ := os.Create(conf.Im.SaveTo)
		fmt.Fprintln(file, string(d))
	}

	if len(*o) > 0 {
		file, _ := os.Create(*o)
		erdh.WritePuml(file, &cons, conf)
	} else {
		erdh.WritePuml(os.Stdout, &cons, conf)
	}
}
