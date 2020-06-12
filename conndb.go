package main

import (
	"database/sql"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Conn ...
func Conn() (*sql.DB, error) {

	// Connect to the "domaint" database.
	return sql.Open("postgres",
		"postgresql://root@localhost:26257/domaint?sslmode=disable")
}

// CreateTable ...
func CreateTable() error {

	_, err := globalDB.Exec(
		"CREATE TABLE IF NOT EXISTS dom (id SERIAL PRIMARY KEY, domain STRING, previous_ssl_grade STRING, checked_at TIMESTAMP)")

	if err != nil {
		log.Fatal(err)
		return err
	}
	_, err2 := globalDB.Exec(
		"CREATE TABLE IF NOT EXISTS servers (id SERIAL PRIMARY KEY, address STRING, ssl_grade STRING)")

	if err2 != nil {
		log.Fatal(err2)
		return err2
	}

	return nil
}

// RegisterDom is for register a domain in database.
func RegisterDom(db *sql.DB, dataDomain *DataDomain) error {

	t := time.Now().UTC().String()
	index := strings.Index(t, "+")
	timeStamp := t[:index]

	_, err := db.Exec(
		"INSERT INTO dom (domain, previous_ssl_grade, checked_at) VALUES ('" + dataDomain.domain + "','" + dataDomain.PreviousSSLGrade + "','" + timeStamp + "')")
	if err != nil {
		return err
	}

	return nil
}

// ConsultDomain ...
func ConsultDomains(db *sql.DB) ([]StructSend, error) {

	rows, err := db.Query("SELECT id,domain,previous_ssl_grade,checked_at FROM dom")
	if err != nil {
		return nil, err
	}

	var domainsLog []StructSend

	for rows.Next() {
		var domain StructSend
		if err := rows.Scan(&domain.ID, &domain.Domain, &domain.PreviousSslGrade, &domain.CheckedAt); err != nil {
			log.Println(err)
			continue
		}
		domain.CheckedAt = strings.Replace(domain.CheckedAt, "T", " ", -1)
		domainsLog = append(domainsLog, domain)
	}

	return domainsLog, nil

}

//cockroach start-single-node --insecure --listen-addr=localhost:26257  --http-addr=localhost:8080
//cockroach sql --insecure
