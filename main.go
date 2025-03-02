package main

import (
	"encoding/json"
	"fmt"
	"log"
	"stillrunning.org/people-verifier/chatgpt"
	"stillrunning.org/people-verifier/wikidata"
)

func main() {

	entityId := "Q51827254"
	db := OpenDb()
	defer db.Close()

	person, err := ReadPerson(db, entityId)
	if err != nil {
		panic(err)
	}
	personJson, _ := json.MarshalIndent(person, "", "  ")
	print("Person JSON: ", string(personJson))

	entityJson, err := wikidata.DownloadWikidataJSON(entityId)
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

	shortDescr, err := wikidata.DownloadWikipediaBriefSummary(lang, title)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Wikipedia summary:", shortDescr)

	messages := []chatgpt.Message{
		{Role: "system", Content: "You are a helpful assistant."},
	}

	prompts := []string{
		"I have these two JSONs related to some person\n\n" +
			"JSON 1:\n" + string(personJson) + "\n\nJSON 2:\n" + shortDescr +
			"\n\nWhich well-known historical figure is this information about?",

		"Reply with a single word \"True\" or \"False\" if that was a real person (True) or not (False): neither a mythical nor a fictional character, nor a non-human (e.g. a famous animal also should fall into False bucket): someone who actually lived on Earth.",

		"Your answer MUST contain a number following by an explanation, no matter how inaccurate it might be, even if it is just a random number. \n" +
			"Calculate their age at the moment of dying. Double-check the provided earlier information in external sources. " +
			"Find out the missing information in external sources. Do your best to do any estimation, judgement or guesses " +
			"with whatever information you have by hand. ",

		"give me their \"birthDate\", \"deathDate\" and the \"deathAge\" as JSON. Use your best guess and other sources of information if the provided information was not enough. Also, give me confidence score in percentage about the values (\"confidence\" key). Return only JSON; any explanation include as \"explanation\" field in JSON",
		"Give me a short description about this person. Include information about their age at the moment of dying and also any circumstances around his death (but not too much, maybe 1-2 sentences: 70% of response about general info, 30% about his death). Return as a plain text.",
	}

	for _, prompt := range prompts {
		println("\n\n-----------------------------------\n\n")
		messages = append(messages, chatgpt.Message{
			Role:    "user",
			Content: prompt,
		})
		fmt.Println("Asking ChatGPT:", prompt)
		response, err := chatgpt.AskChatGPT(messages)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("....ChatGPT replied:", response)
		messages = append(messages, chatgpt.Message{
			Role:    "assistant",
			Content: response,
		})
	}
}
