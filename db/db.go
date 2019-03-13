package db

import (
	"errors"
	"fmt"
	"strings"
	"syscall"

	"github.com/iwot/erdh-go/config"
	"github.com/iwot/erdh-go/erdh"
	"golang.org/x/crypto/ssh/terminal"
)

// ReadDB は対象DBを読み、Constructionを返す
func ReadDB(target string, dbconf config.DBConfig) (*erdh.Construction, error) {
	switch target {
	case "mysql":
		cons, err := ReadMySQL(dbconf)
		return cons, err
	case "sqlite":
		cons, err := ReadSQLite(dbconf)
		return cons, err
	default:
		return &erdh.Construction{}, errors.New("invalid target")
	}
}

// ReadYAML は中間形式ファイルを読み、Constructionを返す
func ReadYAML(path string) (*erdh.Construction, error) {
	return erdh.NewConstructionFromYamlFile(path)
}

func readConsolePassword() (string, error) {
	fmt.Print("Enter DB Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	passwd := string(bytePassword)
	fmt.Println("")
	return strings.TrimSpace(passwd), nil
}
