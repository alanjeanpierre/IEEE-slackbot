package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// Watson object for reading json
type Watson struct {
	DocumentTone struct {
		ToneCategories []struct {
			CategoryID   string `json:"category_id"`
			CategoryName string `json:"category_name"`
			Tones        []struct {
				Score    float64 `json:"score"`
				ToneID   string  `json:"tone_id"`
				ToneName string  `json:"tone_name"`
			} `json:"tones"`
		} `json:"tone_categories"`
	} `json:"document_tone"`
}

//https://watson-api-explorer.mybluemix.net/apis/tone-analyzer-v3#!/tone/GetTone
func watsonToneAnalyzerQuery(text, tones string) (d Watson) {

	t := time.Now()
	time := fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
	textQ := url.QueryEscape(text)
	tonesQ := url.QueryEscape(tones)
	api := "https://watson-api-explorer.mybluemix.net/tone-analyzer/api/v3/tone?"

	url := fmt.Sprintf("%stext=%s&tones=%s&sentences=false&version=%s", api, textQ, tonesQ, time)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Println(err)
	}

	var respObj Watson
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		log.Println(err)
	}

	d = respObj

	return
}

func watsonToneAnalyzer(text, tones string) (s string) {

	var buffer bytes.Buffer

	d := watsonToneAnalyzerQuery(text, tones)

	for _, cat := range d.DocumentTone.ToneCategories {
		buffer.WriteString(fmt.Sprintf("%s:\n", cat.CategoryName))
		for _, tone := range cat.Tones {
			buffer.WriteString(fmt.Sprintf("\t%s -> %f\n", tone.ToneName, tone.Score))
		}
		buffer.WriteString("\n")
	}

	s = buffer.String()

	return
}
