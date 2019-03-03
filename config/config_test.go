package config

import "testing"

func TestDbDsnMySQL(t *testing.T) {
	var dbconf = DBConfig{"mysql", "localhost", "3306", "testdb", "user", "password"}
	dsn, err := dbconf.ToDSN()

	if err != nil {
		t.Fatalf("failed test dbconf.ToDSN() %#v", err)
	}

	if dsn != "user:password@tcp(localhost:3306)/testdb" {
		t.Fatalf("failed test dsn %#v", dsn)
	}
}
