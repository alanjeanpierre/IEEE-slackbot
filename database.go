package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/websocket"
)

type Database struct {
	users     map[string]string
	channels  map[string]string
	banlist   map[string]bool
	reactions map[string]string
	relations map[string]bool
	boss      string
	rootloc   string
	token     string
	ws        *websocket.Conn
	nmsg      int
	uptime    time.Time
	botid     string
	db        *sql.DB
	mutex     sync.Mutex
}

func (db *Database) insertUser(id, usr string) error {

	db.mutex.Lock()
	_, err := db.db.Exec("insert into users values (?, ?)", id, usr)
	db.mutex.Unlock()
	return err
}

func (db *Database) insertChannel(id, channel string) error {

	db.mutex.Lock()
	_, err := db.db.Exec("insert into channels values (?, ?)", id, channel)
	db.mutex.Unlock()
	return err
}

func (db *Database) logMessage(m Message) error {

	db.mutex.Lock()
	_, err := db.db.Exec("insert into logs values (?, ?, ?, ?)", getUnix(m.TS), m.Channel, m.User, m.Text)
	db.mutex.Unlock()
	return err
}

func (db *Database) insertLink(uid, link string) error {

	db.mutex.Lock()
	_, err := db.db.Exec("insert into links values (?, ?)", uid, link)
	db.mutex.Unlock()
	return err
}

func (db *Database) addReaction(uid, trigger, reaction string) error {

	db.reactions[trigger] = reaction
	db.mutex.Lock()
	_, err := db.db.Exec("insert into reactions values (?, ?, ?)", uid, trigger, reaction)
	db.mutex.Unlock()
	return err
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

		// save to db
		db.insertUser(id, usr)
	}

	return usr
}

func (db *Database) getChannel(id string) string {
	channel, ok := db.channels[id]
	if !ok {
		channel = getChannel(id, db.token)
		db.channels[id] = channel
		db.insertChannel(id, channel)
	}

	return channel
}

func (db *Database) getRelation(trigger string) (err error, relation, data string) {
	row := db.db.QueryRow("select * from relations where trigger = ? order by random() limit 1;", trigger)
	if err != nil {
		return err, "", ""
	}
	err = row.Scan(&trigger, &relation, &data)
	return err, relation, data
}

func (db *Database) getAllRelations(trigger string) (err error, data string) {
	rows, err := db.db.Query("select relation, data from relations where trigger = ?", trigger)
	if err != nil {
		return err, ""
	}

	var buffer []string
	for rows.Next() {
		var relation string
		var data string
		err := rows.Scan(&relation, &data)
		if err != nil {
			log.Println("error scanning all relations")
			continue
		}
		buffer = append(buffer, fmt.Sprintf("%s %s", relation, data))
	}
	rows.Close()

	return nil, fmt.Sprintf("%s ", trigger) + strings.Join(buffer, "; ")
}

func (db *Database) addRelation(trigger, relation, data string) error {
	db.relations[trigger] = true
	db.mutex.Lock()
	_, err := db.db.Exec("insert into relations values (?, ?, ?);", trigger, relation, data)
	db.mutex.Unlock()
	return err
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
func (db *Database) load() error {

	d, err := setupDatabase(db.rootloc)
	db.db = d
	if err != nil {
		return err
	}

	db.users = make(map[string]string)
	db.channels = make(map[string]string)
	db.banlist = make(map[string]bool)
	db.reactions = make(map[string]string)
	db.relations = make(map[string]bool)

	rows, err := db.db.Query("select * from users;")
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			log.Println("error scanning users")
			continue
		}
		db.users[id] = name
	}
	rows.Close()

	rows, err = db.db.Query("select * from channels;")
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			log.Println("error scanning channels")
			continue
		}
		db.channels[id] = name
	}
	rows.Close()

	rows, err = db.db.Query("select trigger, reaction from reactions;")
	if err != nil {
		log.Fatal("Unable to access reactions")
	}
	for rows.Next() {
		var trigger string
		var reaction string
		err := rows.Scan(&trigger, &reaction)
		if err != nil {
			log.Println("error scanning reaction rows")
			continue
		}
		db.reactions[trigger] = reaction
	}
	rows.Close()

	rows, err = db.db.Query("select trigger from relations;")
	if err != nil {
		log.Fatal("Unable to access relations")
	}

	for rows.Next() {
		var trigger string
		err := rows.Scan(&trigger)
		if err != nil {
			log.Println("error scanning relation rows")
			continue
		}
		db.relations[trigger] = true
	}
	rows.Close()

	file, err := os.OpenFile(db.rootloc+"banlist", os.O_RDONLY|os.O_CREATE, 0664)
	if err != nil {

		return err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		banned := scanner.Text()
		db.banlist[banned] = true
	}

	return nil
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

// opens (and creates the tables) of the database
func setupDatabase(rootloc string) (*sql.DB, error) {

	db, err := sql.Open("sqlite3", rootloc+"database.db")
	if err != nil {
		return nil, err
	}

	sqlStmt := `
    create table if not exists logs (time datetime, cid text, uid text, message text);
    create table if not exists links (uid text, link text);
    create table if not exists users (uid text primary key, username text);
    create table if not exists channels(cid text primary key, channel text);
	create table if not exists reactions(uid text, trigger text, reaction text);
	create table if not exists relations(trigger text, relation text, data text);
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func getUnix(time string) int {

	ms, err := strconv.Atoi(strings.Split(time, ".")[0])
	if err != nil {
		return 0
	}

	return ms

}
