package helper

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql" // MySQL 驱动
	_ "github.com/lib/pq"              // PostgreSQL 驱动
)

// 验证 DSN 是否可以连接成功
func DBConnTest(dsn string) (bool, error) {
	var db *sql.DB
	var err error

	if strings.HasPrefix(dsn, "postgres://") {
		// PostgreSQL 连接
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return false, fmt.Errorf("failed to open PostgreSQL connection: %v", err)
		}
	} else if strings.HasPrefix(dsn, "mysql://") {
		// MySQL 连接
		dsn = strings.TrimPrefix(dsn, "mysql://")
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			return false, fmt.Errorf("failed to open MySQL connection: %v", err)
		}
	} else {
		return false, fmt.Errorf("unsupported DSN format: must start with 'postgres://' or 'mysql://'")
	}

	defer db.Close()

	// 测试连接
	err = db.Ping()
	if err != nil {
		return false, fmt.Errorf("failed to ping database: %v", err)
	}

	return true, nil
}

// func main() {
// 	// 示例 PostgreSQL DSN
// 	postgresDSN := "postgres://myuser:mypassword@10.0.2.15:5432/mydb"
// 	success, err := testDSN(postgresDSN)
// 	if err != nil {
// 		fmt.Printf("PostgreSQL DSN test failed: %v\n", err)
// 	} else {
// 		fmt.Printf("PostgreSQL DSN test succeeded: %v\n", success)
// 	}

// 	// 示例 MySQL DSN
// 	mysqlDSN := "mysql://root:123456@172.16.1.13:3306/we7_addons?charset=utf8mb4"
// 	success, err = testDSN(mysqlDSN)
// 	if err != nil {
// 		fmt.Printf("MySQL DSN test failed: %v\n", err)
// 	} else {
// 		fmt.Printf("MySQL DSN test succeeded: %v\n", success)
// 	}
// }
