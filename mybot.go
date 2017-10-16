/*

mybot - Illustrative Slack bot in Go

Copyright (c) 2015 RapidLoop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
    "io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
    //"errors"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "usage: mybot slack-bot-token admin-user /root/working/directory \n")
		os.Exit(1)
	}

    var db Database
    db.uptime = time.Now()
    
	// start a websocket-based Real Time API session
	ws, botid, err := slackConnect(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	log.Println("mybot ready, ^C exits")

    db.ws = ws
    db.botid = botid
	db.boss = os.Args[2]
	db.rootloc = os.Args[3]
	db.token = os.Args[1]


	db.load()

	// heartbeat 10s interval
	go func(db *Database) {
		for {
			m := Message{ID : 1234, Type : "ping"}
			//log.Println("Ping")
			postMessage(db.ws, m)
			time.Sleep(10 * time.Second)
		}
	}(&db)

	for {

		err = db.ws.SetReadDeadline(time.Now().Add(30*time.Second))
		if err != nil {
			log.Fatal(err)
		}
		b, err := getRTM(db.ws)
		if err != nil {
			log.Println(err)
			// connection failure
			// try reconnecting..?
			log.Println("Connection failure, attempting to reconnect on 5s intervals...")
			i := 0
            var ws *websocket.Conn
            var id string
			for i = 1; i <= 5; i++ {
				log.Printf("Attempt %d... \n", i)
				var err error
				ws, id, err = slackConnect(os.Args[1])
				if err != nil {
					log.Printf("Attempt %d failure. Sleeping 5 seconds\n", i)
					time.Sleep(5 * time.Second)
				} else {
					break
				}
			}
			
			if i > 5 {
				log.Fatal("Couldn't reconnect. Exiting")
			} else {
				log.Println("Successfully reconnected")
                db.ws = ws
                db.botid = id
				continue // from message read loop
			}
			
		}
		
		r, err := getJSON(b) 
		if err != nil {
			// bad json?
			continue
		}

		if r.Type == "message" {
            db.nmsg += 1
			var m Message
            
            // get full message
			err := json.Unmarshal(b, &m)
			//err := json.Unmarshal(r.X, &m)
			if err != nil {
				panic(err)
			}

			// identify user from ID to readable string
			usr := db.getUser(m.User)

			// identify channel from ID to readable string
			// Private channels and DMs don't show up though :/
			channel := db.getChannel(m.Channel)

			// create a log file named after the channel
			// and log the message
			file, err := os.OpenFile(db.rootloc+"logs/"+channel+".txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
			if err != nil {
				death(m, db.ws)
				log.Fatal(err)
			}
			times := strings.Split(m.TS, ".")
			if len(times) != 2 {
				continue
			}

			tms, tns, err := getTime(times[0], times[1])
			if err != nil {
				death(m, ws)
				log.Fatal(err)
			}

			fmt.Fprintf(file, "%s, %v, %q\n", time.Unix(tms, tns), usr, m.Text)

			err = file.Close()
			if err != nil {
				death(m, ws)
				log.Fatal(err)
			}

            // don't respond to banned people
			if db.isBanned(usr) {
				continue
			}
			// shit posting
			go func(m Message) {

				text := strings.ToLower(m.Text)

				if strings.Contains(text, "doing it") {
					go postReaction(db.token, m.Channel, m.TS, "doing_it")
				}

				if strings.Contains(text, "fucked up") {
					go postReaction(db.token, m.Channel, m.TS, "shinji-sauce")
				}

				// should fix a lot of false positives, but keep onee-sama triggering
				if strings.Contains(text, " one") {
					go postReaction(db.token, m.Channel, m.TS, "wanwanwan")
				}
				
				if strings.Contains(text, "eh") {
					go postReaction(db.token, m.Channel, m.TS, "flag-ca")
				}
			}(m)

			// if bot is mentioned
            if !strings.HasPrefix(m.Text, "<@"+db.botid+">") {
                continue
            }
            
            go func(db *Database, m Message) {
                reply := parsecmd(db, m)
                if reply == "" { // don't reply for silent commands
                    return
                }
                respond(fmt.Sprintf("<@%s> %s", m.User, reply), m, db.ws)
            }(&db, m)
            
		} else if r.Type == "pong" {
			//log.Println("Pong")
			// maybe add some latency checking? idk
			continue
		} else if r.Type == "file_shared" {
        
			var file_shared File_Shared
			err := json.Unmarshal(b, &file_shared)
			if err != nil {
				panic(err)
			}
            
            go func(db *Database, file_shared File_Shared) {
                url := fmt.Sprintf("https://slack.com/api/files.info?token=%s&file=%s", db.token, file_shared.File_ID)
                resp, err := http.Get(url)
                if err != nil || resp.StatusCode != 200{
                    log.Println("Error getting file info from slack")
                    log.Println(err)
                    return
                }
                
                body, err := ioutil.ReadAll(resp.Body)
                resp.Body.Close()
                if err != nil {
                    log.Println(err)
                    return
                }

                var file Files_Info
                err = json.Unmarshal(body, &file)
                if err != nil || !file.OK {
                    log.Println(err)
                    return
                }
            
                if file.File.Title == "@" + db.users[db.botid] {
                
                    out, err := os.Create(db.rootloc + "files/" + file.File.Name)
                    if err != nil {
                        log.Println("Error creating file")
                        log.Println(err)
                        return
                    }
                    
                    client := &http.Client{}
                    req, err:= http.NewRequest("GET", file.File.URL, nil)
                    if err != nil {
                        log.Println("Error downloading file")
                        log.Println(err)
                        return
                    }
                    
                    req.Header.Set("Authorization", "Bearer " + db.token)
                    
                    dl , err := client.Do(req)
                    if err != nil {
                        log.Println(err)
                        out.Close()
                        return
                    }
                    
                    _, err = io.Copy(out, dl.Body)
                    if err != nil {
                        log.Println("Error writing file")
                        log.Println(err)
                    }
                    
                    out.Close()
                    dl.Body.Close()
                }
            }(&db, file_shared)
        }
	}
}

func getTime(tms, tns string) (int64, int64, error) {

	millis, err := strconv.ParseInt(tms, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	nanos, err := strconv.ParseInt(tns, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return millis, nanos, nil

}

func postReaction(token string, channel string, timestamp string, reaction string) {

	url := fmt.Sprintf("https://slack.com/api/reactions.add?token=%s&name=%s&channel=%s&timestamp=%s", token, reaction, channel, timestamp)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		fmt.Println(err)
	}
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// Send a death message to me if my error checking fails
func death(m Message, ws *websocket.Conn) {
	m.Channel = "D67GB3LJ0" // dm to me
	m.Text = "Rip, I'm dead"
	postMessage(ws, m)
}

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

// SlackUserObject RTM response user object
type SlackUserObject struct {
	Ok   bool `json:"ok"`
	User User `json:"user"`
}

// SlackChannelObject RTM response channel object
type SlackChannelObject struct {
	Ok      bool    `json:"ok"`
	Channel Channel `json:"channel"`
}

// User object
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Channel object
type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SlackGroupObject web api response for groups (private channels)
type SlackGroupObject struct {
	Ok    bool `json:"ok"`
	Group struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"group"`
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

// Get the quote via Yahoo. You should replace this method to something
// relevant to your team!
func getQuote(sym string) string {
	sym = strings.ToUpper(sym)
	url := fmt.Sprintf("http://download.finance.yahoo.com/d/quotes.csv?s=%s&f=nsl1op&e=.csv", sym)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	rows, err := csv.NewReader(resp.Body).ReadAll()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	if len(rows) >= 1 && len(rows[0]) == 5 {
		return fmt.Sprintf("%s (%s) is trading at $%s", rows[0][0], rows[0][1], rows[0][2])
	}
	return fmt.Sprintf("unknown response format (symbol was \"%s\")", sym)
}

func respond(response string, m Message, ws *websocket.Conn) {
    m.Text = response
    postMessage(ws, m)
}