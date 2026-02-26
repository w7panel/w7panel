package helper

import "testing"

func TestDSN_Ping2(t *testing.T) {
	dsn, err := DBConnTest("mysql://root:123456@tcp(172.16.1.13:3306)/we7_addons?charset=utf8mb4")
	if err != nil {
		t.Errorf("Failed to parse DSN: %v", err)
	}
	if !dsn {
		t.Error("Failed to ping MySQL database")
	}
}

func TestDSN_PingPostgres2(t *testing.T) {
	dsn, err := DBConnTest("postgres://myuser:mypassword@10.0.2.15:5432/mydb?sslmode=disable")
	if err != nil {
		t.Errorf("Failed to parse DSN: %v", err)
	}
	if !dsn {
		t.Error("Failed to ping Postgres database")
	}
}
