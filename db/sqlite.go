package db

import (
	"database/sql"
	"errors"
	"path/filepath"
	"regexp"

	"github.com/iwot/Sqlite3CreateTableParser/parser"
	"github.com/iwot/erdh-go/config"
	"github.com/iwot/erdh-go/erdh"
	_ "github.com/mattn/go-sqlite3"
)

// ReadSQLite は対象DBを読み、Constructionを返す
func ReadSQLite(dbconf config.DBConfig) (*erdh.Construction, error) {
	var cons erdh.Construction

	db, err := sql.Open("sqlite3", dbconf.DBName)
	if err != nil {
		return &cons, err
	}
	defer db.Close()

	cons.DBName = filepath.Base(dbconf.DBName)

	creates, err := retrieveCreateQueries(db)
	if err != nil {
		return &cons, err
	}

	var tables []erdh.Table
	for tablName, query := range creates {
		table, err := parseCreateQuery(query, cons.DBName, tablName)
		if err != nil {
			return &cons, err
		}
		tables = append(tables, table)
	}

	cons.Tables = tables

	return &cons, nil
}

func retrieveCreateQueries(db *sql.DB) (map[string]string, error) {
	result := map[string]string{}

	sql := `SELECT tbl_name, sql FROM sqlite_master WHERE type = "table"`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var tblName string
		var query string
		err = rows.Scan(&tblName, &query)
		if err != nil {
			return nil, err
		}
		result[tblName] = query
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

var commentReg = regexp.MustCompile(`--\s*.*(\r\n|\n)?`)

func removeQueryComment(query string) string {
	re := commentReg.Copy()
	return re.ReplaceAllString(query, "\n")
}

func parseCreateQuery(query, dbName, tableName string) (erdh.Table, error) {
	var result erdh.Table

	table, errCode := parser.ParseTable(removeQueryComment(query), 0)
	if errCode != parser.ERROR_NONE {
		return result, errors.New("Error during parsing sql")
	}

	result.Name = table.Name
	if table.Schema == "" {
		result.Group = dbName
	} else {
		result.Group = table.Schema
	}
	for _, c := range table.Columns {
		result.AddColumn(
			c.Name,
			c.Type,
			"",
			"",
			c.DefaultExpr,
			c.IsNotnull,
			c.IsPrimaryKey)
	}
	for _, c := range table.Constraints {
		if c.ForeignKeyNum > 0 {
			for i := 0; i < c.ForeignKeyNum; i++ {
				result.AddForeginKey(
					"",
					c.ForeignKeyName[i],
					c.ForeignKeyClause.Table,
					c.ForeignKeyClause.ColumnName[i])
			}
		}
	}

	return result, nil
}
