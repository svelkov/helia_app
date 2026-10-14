package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"helia/config"

	_ "github.com/lib/pq"
)

func main() {
	data, _ := os.ReadFile("config/config.json")
	var cfg config.Config
	_ = json.Unmarshal(data, &cfg)
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s search_path=%s sslmode=disable",
		cfg.DBConfig.DBHost, cfg.DBConfig.DBPort, cfg.DBConfig.DBUser, cfg.DBConfig.DBPassword, cfg.DBConfig.DBName, cfg.DBConfig.DBSearchPath)
	d, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	tx, err := d.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()
	_, _ = tx.Exec("set transaction read only")
	for _, q := range os.Args[1:] {
		var out sql.NullString
		if err := tx.QueryRow(`select string_agg(row_to_json(x)::text, E'\n') from (` + q + `) x`).Scan(&out); err != nil {
			fmt.Println("ERR", err)
			_ = tx.Rollback()
			tx, _ = d.Begin()
			_, _ = tx.Exec("set transaction read only")
			continue
		}
		fmt.Println(out.String)
		fmt.Println("----")
	}
}
