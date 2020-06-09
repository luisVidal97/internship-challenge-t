package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

// CREATE TABLE domains (
// 	-> id STRING(32),
// 	-> created_at TIMESTAMPTZ NOT NULL,
// 	->    updated_at TIMESTAMPTZ,
// 	->    PRIMARY KEY ("id")
// 	-> );

// Conn ...
func Conn() (*sql.DB, error) {

	// Connect to the "domaint" database.
	return sql.Open("postgres",
		"postgresql://root@localhost:26257/domaint?sslmode=disable")
}

// registerDmain ...
func registerDomain(db *sql.DB, domain string) error {

	_, err := db.Exec(
		"INSERT INTO domains (id) VALUES ('" + domain + "')")
	if err != nil {
		return err
	}
	return nil
}

//CompileDaemon -command="yeah.exe"
// ConsultDomain ...
func ConsultDomains(db *sql.DB) ([]string, error) {

	rows, err := db.Query("SELECT id FROM domains")
	if err != nil {
		return nil, err
	}

	var domainsLog []string

	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			log.Println(err)
			continue
		}
		domainsLog = append(domainsLog, domain)
	}

	return domainsLog, nil

}
