package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strings"
)

func OpenDb() *sql.DB {
	db, err := sql.Open("sqlite3", "../people-finder/people.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	return db
}

type Person struct {
	Id           string
	Name         string
	BirthDate    string
	DeathDate    string
	Pic          string
	SiteLinksCnt int
	Age          int
}

func ReadPerson(db *sql.DB, id string) (Person, error) {
	var p Person
	err := db.QueryRow("SELECT id, name, birthDate, deathDate, pic, siteLinksCnt, age "+
		"FROM persons WHERE id = ?", "http://www.wikidata.org/entity/"+id).Scan(&p.Id, &p.Name, &p.BirthDate, &p.DeathDate, &p.Pic, &p.SiteLinksCnt, &p.Age)
	if err != nil {
		return p, err
	}
	if strings.HasPrefix(p.BirthDate, "http") {
		p.BirthDate = ""
	}
	if strings.HasPrefix(p.DeathDate, "http") {
		p.DeathDate = ""
	}
	return p, err
}

func ReadNextPersons(db *sql.DB, cursor string) ([]Person, string, error) {
	var p Person
	ret := make([]Person, 0)
	rows, err := db.Query("SELECT id, name, birthDate, deathDate, pic, siteLinksCnt, age "+
		"FROM persons WHERE id > ? ORDER BY id ASC LIMIT 10", cursor)
	if err != nil {
		return ret, "", err
	}
	defer rows.Close()

	nextCursor := ""
	for rows.Next() {
		err := rows.Scan(&p.Id, &p.Name, &p.BirthDate, &p.DeathDate, &p.Pic, &p.SiteLinksCnt, &p.Age)
		if err != nil {
			return ret, "", err
		}
		if strings.HasPrefix(p.BirthDate, "http") {
			p.BirthDate = ""
		}
		if strings.HasPrefix(p.DeathDate, "http") {
			p.DeathDate = ""
		}
		ret = append(ret, p)
		nextCursor = p.Id
	}
	return ret, nextCursor, nil
}
