# erdh-go

DBからテーブル構造を読み、PlantUML形式のファイルを出力する。  
追加情報ファイルを別途定義して、Foregin Keyを定義していないテーブル間の関連を含めることも可能。

## 設定ファイルなど
source はmysql,sqlite,yamlを指定可能。  
yaml を指定したとき、source_from には別プロセスで出力した中間形式ファイル（intermediate.save_to）を指定。  
例：config_mysql.yaml
```config_mysql.yaml
source: mysql
source_from: C:\path\to\db_con_mysql.yaml
group:
- DATA
- MASTER
intermediate:
  save_to: C:\path\to\db_intermediate_m.yaml
ex_info: C:\path\to\ex_table_info.yaml
```


DB接続情報（MySQLを対象にしてパスワードを省略した場合、入力プロンプトが表示される）  
現在対応しているのはMySQLとSQLite。  
例：db_con_mysql.yaml
```db_con_mysql.yaml
dbtype: mysql
host: localhost
dbname: ELTEST01
user: root
#password: password
```


追加情報（テーブルの属するグループ、リレーション定義）  
例：ex_table_info.yaml
```ex_table_info.yaml
tables:
- table: member_items
  is_master: true
  group: DATA
  relations:
    - referenced_table_name: members
      columns:
        - from: "member_id"
          to: "id"
      this_conn: "one"
      that_conn: "zero-or-one"
    - referenced_table_name: items
      columns:
        - from: "item_id"
          to: "id"
      this_conn: "onlyone"
      that_conn: "many"
- table: items
  is_master: true
  group: DATA
- table: item_types
  is_master: true
  group: MASTER
- table: member_items
  group: DATA
- table: members
  group: DATA
```

this_conn に指定できる文字列とカーディナリティの対応。
```go
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
```

that_conn に指定できる文字列とカーディナリティの対応。
```go
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
```

## 実行

以下のようにして実行するとPlantUML形式のファイルを出力する。
```
erdh-go.exe -config config_mysql.yaml -out result.puml
# go run main.go -config config_mysql.yaml -out result.puml
```


以下のようなファイルが出力される。  
これをplantumlに渡せば画像に(java -jar plantuml.jar result.puml)。  
例：result.puml
```uml
@startuml
package "MASTER" as MASTER {
  entity "item_types" as item_types <<D,TRANSACTION_MARK_COLOR>> {
    + id [PK]
    --
    name
  }
}
package "DATA" as DATA {
  entity "items" as items <<D,TRANSACTION_MARK_COLOR>> {
    + id [PK]
    --
    name
    type
  }
  entity "member_items" as member_items <<D,TRANSACTION_MARK_COLOR>> {
    + id [PK]
    --
    member_id
    enable
    .. 2 more ..
  }
  entity "members" as members <<D,TRANSACTION_MARK_COLOR>> {
    + id [PK]
    --
    name
    gender
  }
}
items  ------  item_types
member_items  ----o|  members
member_items  ||---{  items
@enduml
```
