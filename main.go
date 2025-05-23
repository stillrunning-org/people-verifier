package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"stillrunning.org/people-verifier/wikidata"
	"strconv"
	"strings"
)

func processPerson(priority int, db *sql.DB, person *Person) {
	personJson, _ := json.MarshalIndent(person, "", "  ")
	personJsonStr := strings.Replace(string(personJson), "\n", "", -1)
	print("Person JSON: ", personJsonStr)

	entityJson, err := wikidata.DownloadWikidataJSON(person.Id)
	if err != nil {
		log.Fatal(err)
	}
	var wikidataEntity wikidata.WikidataEntity
	err = json.Unmarshal([]byte(entityJson), &wikidataEntity)
	if err != nil {
		log.Fatal(err)
	}
	sitelinksJson, err := json.MarshalIndent(wikidataEntity.Entities[person.Id[len("http://www.wikidata.org/entity/"):]].Sitelinks, "", "  ")
	sitelinksJsonStr := strings.Replace(string(sitelinksJson), "\n", "", -1)
	if err != nil {
		log.Fatal(err)
	}

	lang, title := wikidataEntity.FindWikipediaArticle()
	fmt.Println("Wikipedia article:", lang, title)
	if lang == "" || title == "" {
		print("No Wikipedia article found for " + person.Name)
		return
	}

	shortDescr, err := wikidata.DownloadWikipediaBriefSummary(lang, title)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Wikipedia summary:", shortDescr)

	dir := filepath.Join(".", "output", "priority-"+strconv.Itoa(priority))
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	fPrompts, err := os.OpenFile(dir+"/age-"+fmt.Sprintf("%03d", person.Age)+".txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer fPrompts.Close()

	_, err = fPrompts.WriteString("\n\n\n-----\n\n\n" + "I have only this information, which might be misleading.\n\n" +
		personJsonStr + "\n" + shortDescr + "\n\n" +
		"Figure out who is the real historical figure behind it. Particularly, " +
		"I need the age they had at the moment of dying. THIS IS THE MUST, THE NUMBER MUST BE IN THE ANSWER. Also include some pictures of the person." +
		"If the exact number is not possible to determine (e.g. birth- or deathDate is not known), " +
		"you must use your best judgement, knowledge and any resources (including external) available to figure it out " +
		"or calculate it, because the number MUST be calculated somehow.\n\n" +
		"Return back your thinking followed by a JSON with the following fields:\n" +
		"- \"id\" (string, use constant \"" + person.Id + "\")\n" +
		"- \"isRealHuman\" (boolean, True/False, if a real human being that ever lived on Earth, not a fictional character, not an animal etc)\n" +
		"- \"birthDate\" (string in format YYYY-MM-DD, might be your estimate)\n" +
		"- \"deathDate\" (string in format YYYY-MM-DD,  might be your estimate),\n" +
		"- \"ageAtDeath\" (integer, might be your estimate),\n" +
		"- \"confidence\" (integer from 0 to 100, your confidence in the correctness of calculated age, in percents)\n" +
		"- \"confidenceExplained\" (string, your explanation of the confidence level)\n" +
		"- \"shortDescriptionEn\" (string, short description in English, must include 70% of general information and 30% about the circumstances of their death)\n" +
		"- \"shortDescriptionFr\" (string, short description in French, same as above)\n" +
		"- \"shortDescriptionDe\" (string, short description in German, same as above)\n" +
		"- \"shortDescriptionEs\" (string, short description in Spanish, same as above)\n" +
		"- \"shortDescriptionRu\" (string, short description in Russian, same as above)\n" +
		"- \"sources\" - (array of strings, the list of sources that you were using to prepare the answer)\n\n\n")
	if err != nil {
		panic(err)
	}

	// store sitelinks in a database

	//fSitelinks, err := os.OpenFile(dir+"/age-"+fmt.Sprintf("%03d", person.Age)+".sitelinks.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	//if err != nil {
	//	panic(err)
	//}
	//defer fSitelinks.Close()
	//
	//_, err = fSitelinks.WriteString("\n\n\n-----\n\n\n" + person.Id + "\n" + person.Name + "\n" + sitelinksJsonStr + "\n")
	//if err != nil {
	//	panic(err)
	//}
}

func main() {

	db := OpenDb()
	defer db.Close()

	starting_age := 29
	priority := 0

	for curr_age := starting_age; curr_age < 100000; curr_age++ {
		persons, err := ReadPersonWithAge(priority, db, curr_age)
		if err != nil {
			log.Fatal(err)
		}
		for _, person := range persons {
			fmt.Println(person)
			processPerson(priority, db, &person)
			println("Successfully processed", curr_age)
		}

	}

}
