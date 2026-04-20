package wikidata

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const userAgent = "people-verifier/1.0 (+https://stillrunning.org/people-verifier)"

type WikidataEntity struct {
	Entities map[string]struct {
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
	}
}

func (entity *WikidataEntity) FindWikipediaArticle() (string, string) {
	if entity.Entities == nil {
		return "", ""
	}
	for _, entity := range entity.Entities {
		if entity.Sitelinks != nil {
			for _, sitelink := range entity.Sitelinks {
				if sitelink.Site == "enwiki" {
					return "en", sitelink.Title
				}
			}
		}
	}
	for _, entity := range entity.Entities {
		if entity.Sitelinks != nil {
			for _, sitelink := range entity.Sitelinks {
				if len(sitelink.Site) == len("01wiki") {
					return sitelink.Site[:2], sitelink.Title
				}
			}
		}
	}
	return "", ""
}

func DownloadWikidataJSON(entityUrl string) (string, error) {
	// split entityUrl by Q delimeter
	entityID := entityUrl[len("http://www.wikidata.org/entity/"):]
	url := "https://www.wikidata.org/w/api.php?action=wbgetentities&ids=" + entityID + "&format=json"
	println("Downloading JSON file: ", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build Wikidata request: %v", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Wikidata entity: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}
	println("Downloaded")
	return string(body), nil
}

func DownloadWikipediaBriefSummary(wikiLang string, title string) (string, error) {
	downloadUrl := "https://" + wikiLang + ".wikipedia.org/w/api.php?format=json&action=query&prop=extracts&exlimit=max&explaintext&exintro&titles=" +
		url.QueryEscape(title) + "&redirects="
	println("Downloading Wikipedia summary: ", downloadUrl)
	req, err := http.NewRequest(http.MethodGet, downloadUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build Wikipedia request: %v", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	println("Downloaded")
	return string(body), nil
}
