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

func ReadPersonWithAge(priority int, db *sql.DB, age int) ([]Person, error) {
	ret := make([]Person, 0)
	pageSize := 20
	if age == 0 || age == 100 {
		pageSize = 40
	}

	rows, err := buildQuery(priority, db, age, pageSize)
	if err != nil {
		return ret, err
	}
	defer rows.Close()

	for rows.Next() {
		var p Person
		err := rows.Scan(&p.Id, &p.Name, &p.BirthDate, &p.DeathDate, &p.Pic, &p.SiteLinksCnt, &p.Age)
		if err != nil {
			return ret, err
		}
		if strings.HasPrefix(p.BirthDate, "http") {
			p.BirthDate = ""
		}
		if strings.HasPrefix(p.DeathDate, "http") {
			p.DeathDate = ""
		}
		ret = append(ret, p)
	}
	if age == 100 {
		return ret, nil
	}
	return ret, nil
}

func buildQuery(priority int, db *sql.DB, age int, pageSize int) (*sql.Rows, error) {
	if age < 100 {
		return db.Query("SELECT id, name, birthDate, deathDate, pic, siteLinksCnt, age "+
			"FROM persons WHERE age = ? ORDER BY siteLinksCnt DESC, id ASC LIMIT ? OFFSET ?", age, pageSize, priority*pageSize)
	} else {
		return db.Query("SELECT id, name, birthDate, deathDate, pic, siteLinksCnt, age "+
			"FROM persons WHERE age >= 100 ORDER BY siteLinksCnt DESC, id ASC LIMIT ? OFFSET ?", pageSize, priority*pageSize)
	}
}
