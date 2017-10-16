package main

import (
    "golang.org/x/net/websocket"
    "time"
    "strings"
    "fmt"
    "os"
    "log"
    "bufio"
    "net/http"
    "encoding/json"
    "io/ioutil"

)

type Database struct {
    users       map[string]string
    channels    map[string]string
    banlist     map[string]bool
    boss        string
    rootloc     string
    token       string
    ws          *websocket.Conn
    nmsg        int
    uptime      time.Time
    botid       string
}

func (db *Database) isElevated(id string) bool {
    return db.boss == db.getUser(id)
}

func (db *Database) isBanned(usr string) bool {
    return usr != db.boss && db.banlist[usr]
}

func (db *Database) getUser(id string) string {
    usr, ok := db.users[id]
    if !ok {
        usr = getUser(id, db.token)
        db.users[id] = usr
    }
    
    return usr
}

func (db *Database) getChannel(id string) string {
    channel, ok := db.channels[id]
    if !ok {
        channel = getChannel(id, db.token)
        db.channels[id] = channel
    }
    
    return channel
}

// save the loaded maps to disk
func (db *Database) save() {
    file, err := os.OpenFile(db.rootloc+"usrs", os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		log.Fatal(err)
	}
	for key, value := range db.users {
		fmt.Fprintf(file, "%s %s\n", key, value)
	}
	file.Close()

	file, err = os.OpenFile(db.rootloc+"channels", os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		log.Fatal(err)
	}
	for key, value := range db.channels {
		fmt.Fprintf(file, "%s %s\n", key, value)
	}
	file.Close()
}

// loads the users and channels and banlist from disk
func (db *Database) load() {

	db.users = make(map[string]string)
	db.channels = make(map[string]string)
	db.banlist = make(map[string]bool)
    
	file, err := os.OpenFile(db.rootloc + "usrs", os.O_RDONLY, 0664)
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			users := strings.Fields(scanner.Text())
			db.users[users[0]] = strings.Join(users[1:], " ")
		}
	}

	file, err = os.OpenFile(db.rootloc + "channels", os.O_RDONLY, 0664)
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			channels := strings.Fields(scanner.Text())
			db.channels[channels[0]] = strings.Join(channels[1:], " ")
		}
	}

	file, err = os.OpenFile(db.rootloc + "banlist", os.O_RDONLY, 0664)
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			banned := scanner.Text()
			db.banlist[banned] = true
		}
	}
}

// queries slack api for user name from ID
func getUser(id string, token string) string {

	url := fmt.Sprintf("https://slack.com/api/users.info?token=%s&user=%s", token, id)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		fmt.Println(err)
		return "idk"
	}

	body, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()
	if err != nil {
		fmt.Println(err)
		return "idk"
	}

	var response SlackUserObject
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("I died - find user")
		fmt.Println(err)
		return "idk"
	}
	return response.User.Name    
}


// queries slack api for channel name
func getChannel(channel string, token string) string {

	url := fmt.Sprintf("https://slack.com/api/channels.info?token=%s&channel=%s", token, channel)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		fmt.Println(err)
		return "idk"
	}

	body, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()
	if err != nil {

		fmt.Println(err)
		return "idk"
	}

	var response SlackChannelObject
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("I died - find channel")
		fmt.Println(err)
		return "idk"
	}

	if response.Ok {
		return response.Channel.Name
	}
	// else its private
	return getGroup(channel, token)
}

func getGroup(channel string, token string) string {
	url := fmt.Sprintf("https://slack.com/api/groups.info?token=%s&channel=%s", token, channel)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		fmt.Println(err)
		return "idk"
	}

	body, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()
	if err != nil {

		fmt.Println(err)
		return "idk"
	}

	var response SlackGroupObject
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("I died - find channel")
		fmt.Println(err)
		return "idk"
	}

	if response.Ok {
		return response.Group.Name
	}
	// else must be a single DM
	return "Private - " + channel
}