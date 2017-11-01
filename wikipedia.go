package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// wikipedia random page struct
type random struct {
	ID    int    `json:"id"`
	NS    int    `json:"ns"`
	Title string `json:"title"`
}

// wikipedia query
type query struct {
	Random []random `json:"random"`
}

// more wiki structs
type wikiresp struct {
	Query query `json:"query"`
}

// queries wiki api for 2 random pages
func wikichall() string {
	resp, err := http.Get("https://en.wikipedia.org/w/api.php?action=query&format=json&list=random&rnnamespace=0&rnlimit=2")
	if err != nil {
		return ""
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()
	if err != nil {
		return ""
	}

	var response wikiresp
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("I died")
		return ""
	}

	page1 := strings.Replace(response.Query.Random[0].Title, " ", "_", -1)
	page2 := strings.Replace(response.Query.Random[1].Title, " ", "_", -1)

	return fmt.Sprintf("try to get from https://en.wikipedia.org/wiki/%s to https://en.wikipedia.org/wiki/%s using only the links on the page!", page1, page2)

}
