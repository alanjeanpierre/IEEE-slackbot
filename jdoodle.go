package main

import (
	"net/http"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"bytes"
	"errors"
	"strings"
	"strconv"
)

type JDoodleExecuteInput struct {
	ClientId	string	`json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Script string	`json:"script"`
	Stdin string 	`json:"stdin"`
	Language	string `json:"language"`
	VersionIndex	int	`json:"versionIndex"`
}

type JDoodleExecuteOutputSuccess struct {
	Output string `json:"output"`
	StatusCode int `json:"statusCode"`
	Memory	string `json:"memory"`
	CpuTime string `json:"cpuTime"`
}

func (s JDoodleExecuteOutputSuccess) format() string {
	return fmt.Sprintf("StatusCode: %d\nMemory Usage: %s\nCPU Time: %s\nOutput:```%s```", s.StatusCode, s.Memory, s.CpuTime, s.Output)
}

type JDoodleExecuteOutputError struct {
	Error string `json:"error"`
	StatusCode int `json:"statusCode"`
}

func (e JDoodleExecuteOutputError) format() string {
	return fmt.Sprintf("StatusCode: %d\nError message:```%s```", e.StatusCode, e.Error)
}


func jdoodleRun(m Message, db *Database) string {
	
	// remove the trailing code snippet markers ```
	s := strings.Trim(m.Text, "` \n")
	// find the beginning of the final block
	end := strings.LastIndex(s, "```")
	if end == -1 {
		return "I need a code snippet man..."
	}
	snippet := s[end + 3:]

	s = strings.TrimRight(s[:end], "` \n")
	end = strings.LastIndex(s, "```")

	in := ""
	if end != -1 {
		in = s[end+3:]
		s = s[:end]
	} 

	parts := strings.Fields(s)
	if len(parts) < 3 {
		return "I need at least a language..."
	}

	ver := 0
	if len(parts) == 4 {
		i, err := strconv.Atoi(parts[3])
		if err == nil {
			ver = i
		}
	}

	lang := parts[2]
	

	err, success, fail := runSnippet(db.parameters["jdoodleId"], db.parameters["jdoodleSecret"], snippet, in, lang, ver)

	if err == nil {
		return success.format()
	} else {
		return fail.format()
	}
}

func runSnippet(id, secret, script, stdin, language string, version int) (error, JDoodleExecuteOutputSuccess, JDoodleExecuteOutputError) {
	
	
	var success JDoodleExecuteOutputSuccess
	var failure JDoodleExecuteOutputError

	input := JDoodleExecuteInput{
		ClientId: id,
		ClientSecret: secret,
		Script: script,
		Stdin: stdin,
		Language: language,
		VersionIndex: version}

	url := "https://api.jdoodle.com/v1/execute"
	j, err := json.Marshal(input)
	if err != nil {
		return err, success, failure
	}


	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err, success, failure
	}
	defer resp.Body.Close()
	
    fmt.Println("response Status:", resp.Status)
    fmt.Println("response Headers:", resp.Header)
    body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	var e error
	if resp.Status == "200 OK" {
		e = nil
		err = json.Unmarshal(body, &success)
	} else {
		e = errors.New("Fail!")
		err = json.Unmarshal(body, &failure)
	}

	if err != nil {
		return err, success, failure
	}
	return e, success, failure

}
