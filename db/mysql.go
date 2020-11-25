package db

import (
	"database/sql"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/iwot/erdh-go/config"
	"github.com/iwot/erdh-go/erdh"
)

// ReadMySQL は対象DBを読み、Constructionを返す
func ReadMySQL(dbconf config.DBConfig) (*erdh.Construction, error) {
	var cons erdh.Construction

	if len(dbconf.Password) == 0 {
		passwd, err := readConsolePassword()
		if err != nil {
			return &cons, err
		}
		dbconf.Password = passwd
	}
	dsn, err := dbconf.ToDSN()
	if err != nil {
		return &cons, err
	}

	db, err := sql.Open(dbconf.DBType, dsn)
	if err != nil {
		return &cons, err
	}
	defer db.Close()

	readMySQLDBName(db, &cons)

	readMySQLTables(db, &cons)

	for _, tbl := range cons.Tables {
		readMySQLTableColumns(db, &cons, tbl.Name)
		readMySQLTableIndexes(db, &cons, tbl.Name)
		readMySQLTableForeginKeys(db, &cons, tbl.Name)
	}

	return &cons, nil
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
		cons.Tables = append(cons.Tables, erdh.Table{Name: tblName, Group:cons.DBName})
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
		table.Columns = append(
			table.Columns,
			erdh.Column{
				Name:       columnName,
				ColumnType: columnType,
				Key:        columnkey,
				Extra:      extra,
				Default:    columnDefaultValue,
				NotNull:    isNullableBool,
				IsPrimary:  isPrimary})
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
		table.Indexes = append(table.Indexes, erdh.Index{Name: indexName, ColumnName: columnName})
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
		referencedTableNameTemp := new(sql.NullString)
		referencedColumnNameTemp := new(sql.NullString)
		err = rows.Scan(&constraintName, &columnName, &referencedTableNameTemp, &referencedColumnNameTemp)
		if err != nil {
			panic(err)
		}
		if referencedTableNameTemp != nil && referencedTableNameTemp.Valid {
			referencedTableName = referencedTableNameTemp.String
		}
		if referencedColumnNameTemp != nil && referencedColumnNameTemp.Valid {
			referencedColumnName = referencedColumnNameTemp.String
		}
		table.ForeginKeys = append(
			table.ForeginKeys,
			erdh.ForeginKey{
				ConstraintName:       constraintName,
				ColumnName:           columnName,
				ReferencedTableName:  referencedTableName,
				ReferencedColumnName: referencedColumnName,
			})
	}
}
