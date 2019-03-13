package erdh

import (
	"fmt"
	"io"

	"github.com/iwot/erdh-go/config"
)

func WritePuml(w io.Writer, cons *Construction, conf *config.Config) error {
	fmt.Fprintln(w, "@startuml")

	// グループ一覧
	groupes := []string{}
	encountered := map[string]bool{}
	table2group := map[string]string{}
	for _, tbl := range cons.Tables {
		encountered[tbl.Group] = true
		table2group[tbl.Name] = tbl.Group
	}
	for key := range encountered {
		groupes = append(groupes, key)
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

	for _, group := range groupes {
		if !isTargetGroup(group) {
			continue
		}

		groupTables := filterByGroup(group)

		// package start
		if len(groupTables) > 0 {
			fmt.Fprintf(w, "package \"%s\" as %s {\n", group, group)
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
