package wikidata

import (
	"encoding/json"
	"testing"
)

func TestWikidataEntity_FindWikipediaArticle_NoSiteLInks(t *testing.T) {
	entity := WikidataEntity{
		Entities: map[string]struct {
			Labels map[string]struct {
				Language string `json:"language"`
				Value    string `json:"value"`
			} `json:"labels"`
			Descriptions map[string]struct {
				Language string `json:"language"`
				Value    string `json:"value"`
			} `json:"descriptions"`
			Sitelinks map[string]struct {
				Site  string `json:"site"`
				Title string `json:"title"`
			} `json:"sitelinks"`
		}{
			"Q46561676": {
				Labels: map[string]struct {
					Language string `json:"language"`
					Value    string `json:"value"`
				}{
					"en": {
						Language: "en",
						Value:    "Mordechai Kikayon",
					},
				},
				Descriptions: map[string]struct {
					Language string `json:"language"`
					Value    string `json:"value"`
				}{
					"en": {
						Language: "en",
						Value:    "Israeli computer scientist",
					},
				},
				Sitelinks: map[string]struct {
					Site  string `json:"site"`
					Title string `json:"title"`
				}{},
			},
		},
	}

	lang, title := entity.FindWikipediaArticle()
	if lang != "" || title != "" {
		t.Errorf("Expected empty strings, got %s and %s", lang, title)
	}
}

func TestWikidataEntity_FindWikipediaArticle_EnSitelinks(t *testing.T) {
	entity := WikidataEntity{
		Entities: map[string]struct {
			Labels map[string]struct {
				Language string `json:"language"`
				Value    string `json:"value"`
			} `json:"labels"`
			Descriptions map[string]struct {
				Language string `json:"language"`
				Value    string `json:"value"`
			} `json:"descriptions"`
			Sitelinks map[string]struct {
				Site  string `json:"site"`
				Title string `json:"title"`
			} `json:"sitelinks"`
		}{
			"Q46561676": {
				Labels: map[string]struct {
					Language string `json:"language"`
					Value    string `json:"value"`
				}{
					"en": {
						Language: "en",
						Value:    "Mordechai Kikayon",
					},
				},
				Descriptions: map[string]struct {
					Language string `json:"language"`
					Value    string `json:"value"`
				}{
					"en": {
						Language: "en",
						Value:    "Israeli computer scientist",
					},
				},
				Sitelinks: map[string]struct {
					Site  string `json:"site"`
					Title string `json:"title"`
				}{
					"enwiki": {
						Site:  "enwiki",
						Title: "Mordechai Kikayon",
					},
					"ruwiki": {
						Site:  "ruwiki",
						Title: "\\u0418\\u0441\\u0442\\u043e\\u0440\\u0438\\u044f",
					},
				},
			},
		},
	}

	lang, title := entity.FindWikipediaArticle()
	if lang != "en" || title != "Mordechai Kikayon" {
		t.Errorf("Expected empty strings, got %s and %s", lang, title)
	}
}

func TestWikidataEntity_FindWikipediaArticle_RuSitelinks(t *testing.T) {
	entity := WikidataEntity{
		Entities: map[string]struct {
			Labels map[string]struct {
				Language string `json:"language"`
				Value    string `json:"value"`
			} `json:"labels"`
			Descriptions map[string]struct {
				Language string `json:"language"`
				Value    string `json:"value"`
			} `json:"descriptions"`
			Sitelinks map[string]struct {
				Site  string `json:"site"`
				Title string `json:"title"`
			} `json:"sitelinks"`
		}{
			"Q46561676": {
				Labels: map[string]struct {
					Language string `json:"language"`
					Value    string `json:"value"`
				}{
					"en": {
						Language: "en",
						Value:    "Mordechai Kikayon",
					},
				},
				Descriptions: map[string]struct {
					Language string `json:"language"`
					Value    string `json:"value"`
				}{
					"en": {
						Language: "en",
						Value:    "Israeli computer scientist",
					},
				},
				Sitelinks: map[string]struct {
					Site  string `json:"site"`
					Title string `json:"title"`
				}{
					"ruwiki": {
						Site:  "ruwiki",
						Title: "\\u0418\\u0441\\u0442\\u043e\\u0440\\u0438\\u044f",
					},
				},
			},
		},
	}

	jsonStr := "{\n                    \"site\": \"ruwiki\",\n                    \"title\": \"\\u0418\\u0441\\u0442\\u043e\\u0440\\u0438\\u044f\",\n                    \"badges\": []\n                }"
	type SiteLinks struct {
		Site  string `json:"site"`
		Title string `json:"title"`
	}
	var ruSiteLinks SiteLinks
	if err := json.Unmarshal([]byte(jsonStr), &ruSiteLinks); err != nil {
		panic(err)
	}
	entity.Entities["Q46561676"].Sitelinks["ruwiki"] = ruSiteLinks

	lang, title := entity.FindWikipediaArticle()
	if lang != "ru" || title != "История" {
		t.Errorf("Expected empty strings, got %s and %s", lang, title)
	}
}
