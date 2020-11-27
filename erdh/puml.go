package erdh

import (
	"fmt"
	"io"
	"sort"

	"github.com/iwot/erdh-go/config"
)

// WritePuml はPlantUML形式のファイルの@startumlから@endumlをio.Writerに書き込む
func WritePuml(w io.Writer, cons *Construction, conf *config.Config, centerGroup string) error {

	// ファイル名を対象グループ名とする。
	if len(centerGroup) > 0 {
		fmt.Fprintln(w, "@startuml " + centerGroup)
	} else {
		fmt.Fprintln(w, "@startuml")
	}

	// グループ一覧
	groups := []string{}
	encountered := map[string]bool{}
	table2group := map[string]string{}
	for _, tbl := range cons.Tables {
		encountered[tbl.Group] = true
		table2group[tbl.Name] = tbl.Group
	}
	for key := range encountered {
		groups = append(groups, key)
	}

	isTargetGroup := func(group string) bool {
		if len(conf.Group) == 0 {
			return true
		}
		for _, g := range conf.Group {
			if g == group {
				return true
			}
		}
		return false
	}

	filterByGroup := func(group string) []Table {
		result := []Table{}
		for _, tbl := range cons.Tables {
			if tbl.Group == group {
				result = append(result, tbl)
			}
		}
		return result
	}

	for _, group := range groups {
		if !isTargetGroup(group) {
			continue
		}

		groupTables := filterByGroup(group)

		// package start
		if len(groupTables) > 0 {
			if group == centerGroup {
				// 中心となるグループに色を付ける
				fmt.Fprintf(w, "package \"%s\" as %s #DDDDDD {\n", group, group)
			} else {
				fmt.Fprintf(w, "package \"%s\" as %s {\n", group, group)
			}
		}

		for _, table := range groupTables {
			// entity start
			fmt.Fprint(w, "  ")
			fmt.Fprintf(w, "entity \"%s\" as %s <<D,TRANSACTION_MARK_COLOR>> {\n", table.Name, table.Name)

			maxColumnShowCount := 3
			absentColumnCount := 0
			columnCount := 0
			for _, column := range table.Columns {
				columnCount++
				if columnCount > maxColumnShowCount {
					absentColumnCount++
					continue
				}
				fmt.Fprint(w, "    ")
				if column.IsPrimary {
					fmt.Fprint(w, "+ ")
					fmt.Fprint(w, column.Name)
					fmt.Fprintln(w, " [PK]")

					fmt.Fprint(w, "    ")
					fmt.Fprintln(w, "--")
				} else {
					fmt.Fprintln(w, column.Name)
				}
			}

			// 省略したカラムの個数に応じて出力
			if absentColumnCount > 0 {
				fmt.Fprintf(w, "    .. %d more ..\n", absentColumnCount)
			}

			// entity end
			fmt.Fprintln(w, "  }")
		}

		// package end
		if len(groupTables) > 0 {
			fmt.Fprintln(w, "}")
		}
	}

	// カーディナリティ
	for _, tbl := range cons.Tables {
		if !isTargetGroup(tbl.Group) {
			continue
		}

		for _, exr := range tbl.ExRelations {
			if !isTargetGroup(table2group[exr.ReferencedTableName]) {
				continue
			}
			fmt.Fprint(w, tbl.Name)
			fmt.Fprint(w, "  ")
			fmt.Fprint(w, GetThisCardinality(exr.ThisConn))
			fmt.Fprint(w, "--")
			fmt.Fprint(w, GetThatCardinality(exr.ThatConn))
			fmt.Fprint(w, "  ")
			fmt.Fprintln(w, exr.ReferencedTableName)
		}
	}

	fmt.Fprintln(w, "@enduml")

	return nil
}

func contains(s []string, test string) bool {
	for _, v := range s {
		if test == v {
			return true
		}
	}
	return false
}

func getGroupNameByTableName(tbl string, cons *Construction) string {
	for _, c := range cons.Tables {
		if tbl == c.Name {
			return c.Group
		}
	}
	return ""
}

func removeOtherGroupRelation(tbl *Table, cons *Construction, relationGroups []string) {
	newFKeys := []ForeginKey{}
	newEXRels := []ExRelation{}

	for _, f := range tbl.ForeginKeys {
		if contains(relationGroups, getGroupNameByTableName(f.ReferencedTableName, cons)) {
			newFKeys = append(newFKeys, f)
		}
	}
	for _, e := range tbl.ExRelations {
		if contains(relationGroups, getGroupNameByTableName(e.ReferencedTableName, cons)) {
			newEXRels = append(newEXRels, e)
		}
	}
	tbl.ForeginKeys = newFKeys
	tbl.ExRelations = newEXRels
}

// WritePumlByGroup はWritePumlをグループごとに適用する
func WritePumlByGroup(w io.Writer, cons *Construction, conf *config.Config) error {
	// グループ一覧
	groups := []string{}
	// centerGroupTables := []string{}
	for _, t := range cons.Tables {
		if !contains(groups, t.Group) {
			groups = append(groups, t.Group)
			// centerGroupTables = append(centerGroupTables, t.Name)
		}
	}
	sort.SliceStable(groups, func(i, j int) bool { return groups[i] < groups[j] })

	// テーブルから参照しているテーブルを集めるための関数
	addReferenceTableToCons := func (refInfo ReferencedTableInfo, cons *Construction, tables *[]Table) []string {
		relationGroups := []string{}
		for _, tbl1 := range *tables {
			if refInfo.GetReferencedTableName() == tbl1.Name {
				cons.AddTable(tbl1)
				relationGroups = append(relationGroups, tbl1.Group)
				// tbl1が属するグループも集める
				for _, tbl2 := range *tables {
					if tbl1.Group == tbl2.Group {
						cons.AddTable(tbl2)
					}
				}
			}
		}
		return relationGroups
	}

	// グループごとにページ書き出し
	for _, centerGroup := range groups {
		thisCons := &Construction{cons.DBName, []Table{}}
		relationGroups := []string{}
		relationGroups = append(relationGroups, centerGroup)
		for _, t := range cons.Tables {
			if centerGroup == t.Group {
				thisCons.AddTable(t)

				//  このテーブルから参照しているテーブルを取得
				for _, t2 := range t.ForeginKeys {
					t := addReferenceTableToCons(t2, thisCons, &cons.Tables)
					if len(t) > 0 {
						relationGroups = append(relationGroups, t...)
					}
				}
				for _, t2 := range t.ExRelations {
					t := addReferenceTableToCons(t2, thisCons, &cons.Tables)
					if len(t) > 0 {
						relationGroups = append(relationGroups, t...)
					}
				}

				// このテーブルを参照しているテーブルを取得
				for _, t2 := range cons.Tables {
					if centerGroup == t2.Group {
						continue
					}
					found := false
					if !found {
						for _, ex := range t2.ForeginKeys {
							if ex.ReferencedTableName == t.Name {
								found = true
								break
							}
						}
					}
					if !found {
						for _, ex := range t2.ExRelations {
							if ex.ReferencedTableName == t.Name {
								found = true
								break
							}
						}
					}

					if found {
						thisCons.AddTable(t2)
						// t2が属するグループも集める
						for _, t3 := range cons.Tables {
							if t2.Group == t3.Group {
								thisCons.AddTable(t3)
								relationGroups = append(relationGroups, t3.Group)
							}
						}
					}
				}
			}
		}

		for i := range thisCons.Tables {
			removeOtherGroupRelation(&thisCons.Tables[i], cons, relationGroups)
		}

		// Tablesをソート
		sort.SliceStable(thisCons.Tables, func(i, j int) bool { return thisCons.Tables[i].Name < thisCons.Tables[j].Name })
		
		err := WritePuml(w, thisCons, conf, centerGroup)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetThisCardinality は左側のカーディナリティを返す
func GetThisCardinality(this string) string {
	switch this {
	case "one":
		return "--"
	case "only-one":
		fallthrough
	case "onlyone":
		return "||"
	case "zero-or-one":
		fallthrough
	case "zeroorone":
		return "|o"
	case "many":
		return "}-"
	case "onemore":
		fallthrough
	case "one-more":
		return "}|"
	case "zeromany":
		fallthrough
	case "zero-many":
		return "}o"
	default:
		return "--"
	}
}

// GetThatCardinality は右側のカーディナリティを返す
func GetThatCardinality(that string) string {
	switch that {
	case "one":
		return "--"
	case "only-one":
		fallthrough
	case "onlyone":
		return "||"
	case "zero-or-one":
		fallthrough
	case "zeroorone":
		return "o|"
	case "many":
		return "-{"
	case "onemore":
		fallthrough
	case "one-more":
		return "|{"
	case "zeromany":
		fallthrough
	case "zero-many":
		return "o{"
	default:
		return "--"
	}
}
