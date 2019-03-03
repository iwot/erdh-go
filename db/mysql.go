package db

import (
	"database/sql"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/iwot/erdh-go/config"
	"github.com/iwot/erdh-go/erdh"
)

// ReadMySQL は対象DBを読み、Constructionを返す
func ReadMySQL(dbconf config.DBConfig) erdh.Construction {
	dsn, err := dbconf.ToDSN()
	if err != nil {
		panic(err.Error())
	}

	db, err := sql.Open(dbconf.DBType, dsn)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	var cons erdh.Construction
	readMySQLDBName(db, &cons)

	readMySQLTables(db, &cons)

	for _, tbl := range cons.Tables {
		readMySQLTableColumns(db, &cons, tbl.Name)
		readMySQLTableIndexes(db, &cons, tbl.Name)
		readMySQLTableForeginKeys(db, &cons, tbl.Name)
	}

	return cons
}

func readMySQLDBName(db *sql.DB, cons *erdh.Construction) {
	rows, err := db.Query("SELECT database() AS db_name")
	if err != nil {
		panic(err.Error())
	}
	for rows.Next() {
		var dbName string
		err := rows.Scan(&dbName)
		if err != nil {
			panic(err.Error())
		}
		cons.DBName = dbName
		break
	}
}

func readMySQLTables(db *sql.DB, cons *erdh.Construction) {
	rows, err := db.Query("show tables")
	if err != nil {
		panic(err.Error())
	}
	for rows.Next() {
		var tblName string
		err := rows.Scan(&tblName)
		if err != nil {
			panic(err.Error())
		}
		cons.Tables = append(cons.Tables, erdh.Table{Name: tblName})
	}
}

func readMySQLTableColumns(db *sql.DB, cons *erdh.Construction, tableName string) {
	table := cons.GetTableMut(tableName)

	query := `
	SELECT column_name
        , column_type
        , column_key
        , extra
        , column_default
        , is_nullable
    FROM information_schema.columns c
    WHERE c.table_schema = ?
    AND c.table_name = ?
	ORDER BY ordinal_position`

	stmt, err := db.Prepare(query)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(cons.DBName, tableName)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var (
			columnName    string
			columnType    string
			columnkey     string
			extra         string
			columnDefault sql.NullString
			isNullable    string
		)
		err = rows.Scan(&columnName, &columnType, &columnkey, &extra, &columnDefault, &isNullable)
		if err != nil {
			panic(err)
		}
		var columnDefaultValue string
		if columnDefault.Valid {
			columnDefaultValue = columnDefault.String
		}
		isNullableBool := true
		if strings.ToUpper(isNullable) == "NO" {
			isNullableBool = false
		}
		isPrimary := false
		if strings.ToUpper(columnkey) == "PRI" {
			isPrimary = true
		}
		table.Columns = append(table.Columns, erdh.Column{columnName, columnType, columnkey, extra, columnDefaultValue, isNullableBool, isPrimary})
	}
}

func readMySQLTableIndexes(db *sql.DB, cons *erdh.Construction, tableName string) {
	table := cons.GetTableMut(tableName)

	query := `
	SELECT index_name
         , column_name
      FROM information_schema.statistics
     WHERE table_schema = ?
       AND table_name = ?`

	stmt, err := db.Prepare(query)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(cons.DBName, tableName)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var (
			indexName  string
			columnName string
		)
		err = rows.Scan(&indexName, &columnName)
		if err != nil {
			panic(err)
		}
		table.Indexes = append(table.Indexes, erdh.Index{indexName, columnName})
	}
}

func readMySQLTableForeginKeys(db *sql.DB, cons *erdh.Construction, tableName string) {
	table := cons.GetTableMut(tableName)

	query := `
	SELECT constraint_name
         , column_name
         , referenced_table_name
         , referenced_column_name
      FROM information_schema.key_column_usage
     WHERE constraint_schema = ?
       AND table_name = ?
       AND constraint_name <> 'PRIMARY'
     ORDER BY CONSTRAINT_NAME`

	stmt, err := db.Prepare(query)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(cons.DBName, tableName)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var (
			constraintName       string
			columnName           string
			referencedTableName  string
			referencedColumnName string
		)
		err = rows.Scan(&constraintName, &columnName, &referencedTableName, &referencedColumnName)
		if err != nil {
			panic(err)
		}
		table.ForeginKeys = append(table.ForeginKeys, erdh.ForeginKey{constraintName, columnName, referencedTableName, referencedColumnName})
	}
}
