package erdh

import "github.com/iwot/erdh-go/config"

type Construction struct {
	DBName string  `yaml:"db_name"`
	Tables []Table `yaml:"tables"`
}

type Table struct {
	Name        string       `yaml:"table"`
	Group       string       `yaml:"group"`
	Columns     []Column     `yaml:"columns"`
	Indexes     []Index      `yaml:"indexes"`
	ForeginKeys []ForeginKey `yaml:"foreign_keys"`
	ExRelations []ExRelation `yaml:"ex-relations"`
	IsMaster    bool         `yaml:"is-master"`
}

func (c *Construction) GetTableMut(tblName string) *Table {
	for idx, tbl := range c.Tables {
		if tbl.Name == tblName {
			return &c.Tables[idx]
		}
	}
	c.Tables = append(c.Tables, Table{Name: tblName})
	return &c.Tables[len(c.Tables)-1]
}

func (c *Construction) ApplyExInfo(exInfo config.ExtraConfig) {
	for _, ex := range exInfo.Tables {
		table := c.GetTableMut(ex.Name)
		table.IsMaster = ex.IsMaster
		table.Group = ex.Group
		for _, exr := range ex.Relations {
			var columns []ExRelationColumn
			for _, exrc := range exr.Columns {
				columns = append(columns, ExRelationColumn{exrc.From, exrc.To})
			}
			var t = ExRelation{exr.ReferencedTableName, columns, exr.ThisConnection, exr.ThatConnection}
			table.ExRelations = append(table.ExRelations, t)
		}
	}
}

func (t *Table) AddColumn(name, columnType, key, extra, def string, notnull, isPrimary bool) {
	t.Columns = append(t.Columns, Column{name, columnType, key, extra, def, notnull, isPrimary})
}

func (t *Table) AddIndex(indexName, columnName string) {
	t.Indexes = append(t.Indexes, Index{indexName, columnName})
}

func (t *Table) AddForeginKeys(constraintName, columnName, referencedTableName, ReferencedColumnName string) {
	t.ForeginKeys = append(t.ForeginKeys, ForeginKey{constraintName, columnName, referencedTableName, ReferencedColumnName})
}

func (t *Table) AddExRelations(referencedTableName string, columns []ExRelationColumn, thisConn, thatConn string) {
	t.ExRelations = append(t.ExRelations, ExRelation{referencedTableName, columns, thisConn, thatConn})
}

type Column struct {
	Name       string `yaml:"name"`
	ColumnType string `yaml:"type"`
	Key        string `yaml:"key"`
	Extra      string `yaml:"extra"`
	Default    string `yaml:"default"`
	NotNull    bool   `yaml:"not_null"`
	IsPrimary  bool   `yaml:"is_primary"`
}

type Index struct {
	Name       string `yaml:"name"`
	ColumnName string `yaml:"column_name"`
}

type ForeginKey struct {
	ConstraintName       string `yaml:"constraint_name"`
	ColumnName           string `yaml:"column_name"`
	ReferencedTableName  string `yaml:"referenced_table_name"`
	ReferencedColumnName string `yaml:"referenced_column_name"`
}

type ExRelation struct {
	ReferencedTableName string             `yaml:"referenced_table_name"`
	Columns             []ExRelationColumn `yaml:"columns"`
	ThisConn            string             `yaml:"this_conn"`
	ThatConn            string             `yaml:"that_conn"`
}

type ExRelationColumn struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}
