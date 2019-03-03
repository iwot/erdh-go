package db

import (
	"errors"

	"github.com/iwot/erdh-go/config"
	"github.com/iwot/erdh-go/erdh"
)

// ReadDB は対象DBを読み、Constructionを返す
func ReadDB(target string, dbconf config.DBConfig) (erdh.Construction, error) {
	switch target {
	case "mysql":
		cons := ReadMySQL(dbconf)
		return cons, nil
	default:
		return erdh.Construction{}, errors.New("invalid target")
	}
}
