package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"fmt"
	"os"
	"log"
	"github.com/joho/godotenv"
	"database/sql"
	_ "github.com/lib/pq"
)

var DB *gorm.DB
var testSqlDB *sql.DB

func InitDB() {
	godotenv.Load()
	var err error
	dsn := os.Getenv("dsn")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
}

func InitTestDB() {
	godotenv.Load()
	var err error
    testSqlDB, err = sql.Open("postgres", "postgres://tsdbadmin:a68z7j3oapj49z57@qdr3nzawux.lq1rm1zyzp.tsdb.cloud.timescale.com:36635/tsdb?sslmode=require")
    if err != nil {
        log.Printf("failed to open database connection: %v", err)
    }

    DB, err = gorm.Open(postgres.New(postgres.Config{
        Conn: testSqlDB,
        PreferSimpleProtocol: true, // simplifies testing
    }), &gorm.Config{})
    if err != nil {
        log.Printf("failed to connect to database with gorm: %v", err)
    }

    if err = testSqlDB.Ping(); err != nil {
        log.Printf("failed to ping database: %v", err)
    }

	_ = DB.Exec("DELETE FROM messages;")
	_ = DB.Exec("DELETE FROM message_chats;")
	_ = DB.Exec("DELETE FROM followers;")
	_ = DB.Exec("DELETE FROM users;")
}