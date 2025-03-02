package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"stillrunning.org/people-verifier/chatgpt"
	"stillrunning.org/people-verifier/wikidata"
	"time"
)

func processPerson(db *sql.DB, person *Person) {
	personJson, _ := json.MarshalIndent(person, "", "  ")
	print("Person JSON: ", string(personJson))

	entityJson, err := wikidata.DownloadWikidataJSON(person.Id)
	if err != nil {
		log.Fatal(err)
	}
	var wikidataEntity wikidata.WikidataEntity
	err = json.Unmarshal([]byte(entityJson), &wikidataEntity)
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

	messages := []chatgpt.Message{
		//{Role: "system", Content: "You are a helpful assistant."},
	}

	prompts := []struct {
		code   string
		prompt string
	}{

		{
			"load_info",
			"I have these two JSONs related to some person\n\n" +
				"JSON 1:\n" + string(personJson) + "\n\nJSON 2:\n" + shortDescr +
				"\n\nWhich well-known historical figure is this information about?",
		},

		{
			"is_real",
			"Reply with a single word \"True\" or \"False\" if that was a real person (True) or not (False): neither a mythical nor a fictional character, nor a non-human (e.g. a famous animal also should fall into False bucket): someone who actually lived on Earth.",
		},

		{
			"calculus",
			"Your answer MUST contain a number following by an explanation, no matter how inaccurate it might be, even if it is just a random number. \n" +
				"Calculate their age at the moment of dying. Double-check the provided earlier information in external sources. " +
				"Find out the missing information in external sources. Do your best to do any estimation, judgement or guesses " +
				"with whatever information you have by hand. ",
		},

		{
			"corrected_json",
			"give me their \"birthDate\", \"deathDate\" and the \"deathAge\" as JSON. Use your best guess and other sources of information if the provided information was not enough. Also, give me confidence score in percentage about the values (\"confidence\" key). Return only JSON; any explanation include as \"explanation\" field in JSON",
		},

		{
			"short_descr",
			"Give me a short description about this person, in English. Include information about their age at the moment of dying and also any circumstances around his death (but not too much, maybe 1-2 sentences: 70% of response about general info, 30% about his death). Return as a plain text.",
		},
	}

	for _, prompt := range prompts {
		println("\n\n-----------------------------------\n\n")
		messages = append(messages, chatgpt.Message{
			Role:    "user",
			Content: prompt.prompt,
		})
		fmt.Println("Asking ChatGPT:", prompt)
		response, err := chatgpt.AskChatGPT(messages)
		if err != nil {
			// sleep 1 second and retry
			println("chatgpt error: ", err)
			err = nil
			time.Sleep(1 * time.Second)
			response, err = chatgpt.AskChatGPT(messages)
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Println("....ChatGPT replied:", response)
		if prompt.code == "is_real" && response == "False" {
			break
		}
		messages = append(messages, chatgpt.Message{
			Role:    "assistant",
			Content: response,
		})
		if prompt.code == "short_descr" {
			response, err = chatgpt.AskChatGPT([]chatgpt.Message{
				//{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "Translate to Russian: " + response},
			})
			fmt.Println(".... Russian description:", response)
		}
	}
}

func main() {

	cursor := "http://www.wikidata.org/entity/Q1000203"
	db := OpenDb()
	defer db.Close()

	for {
		persons, nextCursor, err := ReadNextPersons(db, cursor)
		if err != nil {
			log.Fatal(err)
		}
		for _, person := range persons {
			fmt.Println(person)
			processPerson(db, &person)
			cursor = person.Id
			println("Successfully processed", cursor)
		}
		if nextCursor == "" {
			break
		}
	}

}
