package erdh

import "github.com/iwot/erdh-go/config"

// Construction 中間形式
type Construction struct {
	DBName string  `yaml:"db_name"`
	Tables []Table `yaml:"tables"`
}

// Table は中間形式中のテーブル型
type Table struct {
	Name        string       `yaml:"table"`
	Group       string       `yaml:"group"`
	Columns     []Column     `yaml:"columns"`
	Indexes     []Index      `yaml:"indexes"`
	ForeginKeys []ForeginKey `yaml:"foreign_keys"`
	ExRelations []ExRelation `yaml:"ex-relations"`
	IsMaster    bool         `yaml:"is-master"`
}

// GetTableMut は指定したテーブル名のTableへのポインタを返す
func (c *Construction) GetTableMut(tblName string) *Table {
	for idx, tbl := range c.Tables {
		if tbl.Name == tblName {
			return &c.Tables[idx]
		}
	}
	c.Tables = append(c.Tables, Table{Name: tblName})
	return &c.Tables[len(c.Tables)-1]
}

// GetGroupToTablesMap はグループからテーブルリストを得るためのマップを返す
func (c Construction) GetGroupToTablesMap() map[string][]string {
	result := map[string][]string{}
	for _, t := range c.Tables {
		if val, ok := result[t.Group]; ok {
			result[t.Group] = append(val, t.Name)
		} else {
			result[t.Group] = []string{t.Name}
		}
	}

	return result
}

// GetTableToGroupMap はテーブルからグループを得るためのマップを返す
func (c Construction) GetTableToGroupMap() map[string]string {
	result := map[string]string{}
	for _, t := range c.Tables {
		result[t.Name] = t.Group
	}

	return result
}

// UpdateExRelationsFromForeignKeys は ForeginKeys を元にして ExRelations を更新する
func (c *Construction) UpdateExRelationsFromForeignKeys() {
	for ti, t := range c.Tables {
		var exRelationsMap = map[string]ExRelation{}
		for _, f := range t.ForeginKeys {
			if val, ok := exRelationsMap[f.ReferencedTableName]; ok {
				val.Columns = append(val.Columns, ExRelationColumn{From: f.ColumnName, To: f.ReferencedColumnName})
			} else {
				exRelationsMap[f.ReferencedTableName] = ExRelation{
					ReferencedTableName: f.ReferencedTableName,
					Columns:             []ExRelationColumn{},
					ThisConn:            "one",
					ThatConn:            "one",
				}
			}
		}

		c.Tables[ti].ExRelations = []ExRelation{}
		for _, e := range exRelationsMap {
			c.Tables[ti].ExRelations = append(c.Tables[ti].ExRelations, e)
		}
	}
}

// ApplyExInfo は config.ExtraConfig を ExRelations に適用する
func (c *Construction) ApplyExInfo(exInfo config.ExtraConfig) {
	for _, ex := range exInfo.Tables {
		table := c.GetTableMut(ex.Name)
		table.IsMaster = ex.IsMaster
		table.Group = ex.Group
		for _, exr := range ex.Relations {
			e := table.GetExRelationOfReferencedTableMut(exr.ReferencedTableName)
			e.ThisConn = exr.ThisConnection
			e.ThatConn = exr.ThatConnection
			// var columns []ExRelationColumn
			for _, exrc := range exr.Columns {
				// columns = append(columns, ExRelationColumn{exrc.From, exrc.To})
				e.Columns = append(e.Columns, ExRelationColumn{exrc.From, exrc.To})
			}
			// var t = ExRelation{exr.ReferencedTableName, columns, exr.ThisConnection, exr.ThatConnection}
			// table.ExRelations = append(table.ExRelations, t)
		}
	}
}

func (t *Table) AddColumn(name, columnType, key, extra, def string, notnull, isPrimary bool) {
	t.Columns = append(t.Columns, Column{name, columnType, key, extra, def, notnull, isPrimary})
}

func (t *Table) AddIndex(indexName, columnName string) {
	t.Indexes = append(t.Indexes, Index{indexName, columnName})
}

func (t *Table) AddForeginKey(constraintName, columnName, referencedTableName, ReferencedColumnName string) {
	t.ForeginKeys = append(t.ForeginKeys, ForeginKey{constraintName, columnName, referencedTableName, ReferencedColumnName})
}

func (t *Table) AddExRelations(referencedTableName string, columns []ExRelationColumn, thisConn, thatConn string) {
	t.ExRelations = append(t.ExRelations, ExRelation{referencedTableName, columns, thisConn, thatConn})
}

func (t *Table) GetExRelationOfReferencedTableMut(tableName string) *ExRelation {
	for i, e := range t.ExRelations {
		if e.ReferencedTableName == tableName {
			return &t.ExRelations[i]
		}
	}

	t.ExRelations = append(
		t.ExRelations,
		ExRelation{
			ReferencedTableName: tableName,
			Columns:             []ExRelationColumn{},
			ThisConn:            "one",
			ThatConn:            "one"})
	return &t.ExRelations[len(t.ExRelations)-1]
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
