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
	"errors"
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: mybot parameter-file \n")
		os.Exit(1)
	}

	var db Database
	db.uptime = time.Now()
	err := db.load(os.Args[1])

	// start a websocket-based Real Time API session
	ws, botid, err := slackConnect(db.token)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("mybot ready, ^C exits")

	db.ws = ws
	db.botid = botid
    if err != nil {
        log.Fatal(err)
    }
	defer db.db.Close()
    
    
	// setup files and folders
	_ = os.Mkdir(db.rootloc + "logs", 0777)
	_ = os.Mkdir(db.rootloc + "files", 0777)

	// heartbeat 10s interval
	go func(db *Database) {
		for {
			m := Message{ID: 1234, Type: "ping"}
			//log.Println("Ping")
			postMessage(db.ws, m)
			time.Sleep(10 * time.Second)
		}
	}(&db)

	readLoop(&db)
	log.Println("Finished")
	
}

func readLoop(db *Database) {
    
    for {

		err := db.ws.SetReadDeadline(time.Now().Add(30 * time.Second))
		if err != nil {
			log.Fatal(err)
		}
		b, err := getRTM(db.ws)
		if err != nil {
			log.Println(err)
			db.ws, db.botid, err = recoverConn()
			if err != nil {
				log.Fatal("Unable to reconnect. Exiting")
			}
			continue
		}

		r, err := getJSON(b)
		if err != nil {
			// bad json?
			continue
		}

		switch r.Type {
            case "pong":
                continue
            case "message":
                db.nmsg += 1
                var m Message
                err := json.Unmarshal(b, &m) // get the full message
                if err != nil {
                    log.Println(err)
                    continue
                }
                
                //don't read blank messages
                if m.Text == "" {
                    continue
                }
                usr := db.getUser(m.User)
                channel := db.getChannel(m.Channel)
                
                logmsg(db, m, usr, channel) // log message to (deprecated) files
                err = db.logMessage(m) // log to sql db
                if err != nil {
                    log.Println(err)
                }
                go parseMessageContent(db, m)
                
                if strings.HasPrefix(m.Text, "<@"+db.botid+">") {

                    go func(db *Database, m Message) {
                        reply := parsecmd(db, m)
                        if reply == "" { // don't reply for silent commands
                            return
                        }
                        respond(fmt.Sprintf("<@%s> %s", m.User, reply), m, db.ws)
                    }(db, m)
                }
                
            case "file_shared":
                go func(db *Database, b []byte) {
                    err := downloadFile(db, b)        
                    if err != nil {
                        log.Println(err)
                        return
                    }
                }(db, b)
        }
	}
}

func parseMessageContent(db *Database, m Message) {    
    text := strings.ToLower(m.Text)
    for trigger, reaction := range db.reactions {
        if strings.Contains(text, trigger) {
            postReaction(db.token, m.Channel, m.TS, reaction) 
            time.Sleep(time.Second)
        }
	}    
	
	ok, _ := db.relations[m.Text]
	if ok {
		err, relation, data := db.getRelation(m.Text)
		if err != nil {
			log.Println(err)
		}
		respond(fmt.Sprintf("%s %s %s", m.Text, relation, data), m, db.ws)
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

func recoverConn() (*websocket.Conn, string, error) {
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
			sleeptime := time.Duration(math.Pow(5, float64(i)))
			log.Printf("Attempt %d failure. Sleeping %d seconds\n", i, sleeptime)
			time.Sleep(sleeptime * time.Second)
		} else {
			break
		}
	}

	if i > 5 {
		return nil, "", errors.New("Failed to reconnect")
	} else {
		log.Println("Successfully reconnected")
		return ws, id, nil
	}
}

func logmsg(db *Database, m Message, usr, channel string) {
    file, err := os.OpenFile(db.rootloc+"logs/"+channel+".txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
    if err != nil {
        death(m, db.ws)
        log.Fatal(err)
    }
    times := strings.Split(m.TS, ".")
    if len(times) != 2 {
        return
    }

    tms, tns, err := getTime(times[0], times[1])
    if err != nil {
        death(m, db.ws)
        log.Fatal(err)
    }

    fmt.Fprintf(file, "%s, %v, %q\n", time.Unix(tms, tns), usr, m.Text)

    err = file.Close()
    if err != nil {
        death(m, db.ws)
        log.Fatal(err)
    }    
}

func downloadFile(db *Database, b []byte) error {
    var file_shared File_Shared
    err := json.Unmarshal(b, &file_shared)
    if err != nil {
        return err
    }
    file_info, err := getFileInformation(db, file_shared)  
    if err != nil {
        return err
    }
    
    if file_info.File.Title == "@"+db.users[db.botid] {
        err = downloadFileFromSlack(db, file_info)
        if err != nil {
            return err
        }
        
    }
    
    return nil
}

func getFileInformation(db *Database, file_shared File_Shared) (file Files_Info, err error) {
    url := fmt.Sprintf("https://slack.com/api/files.info?token=%s&file=%s", db.token, file_shared.File_ID)
    resp, err := http.Get(url)
    if err != nil || resp.StatusCode != 200 {
        return
    }
      
    body, err := ioutil.ReadAll(resp.Body)
    resp.Body.Close()
    if err != nil {
        return
    }  
    
    err = json.Unmarshal(body, &file)
    if err != nil || !file.OK {
        return
    }
    
    return
}

func downloadFileFromSlack(db *Database, file Files_Info) error {
    out, err := os.Create(db.rootloc + "files/" + file.File.Name)
    if err != nil {
        return err
    }

    client := &http.Client{}
    req, err := http.NewRequest("GET", file.File.URL, nil)
    if err != nil {
        return err
    }

    req.Header.Set("Authorization", "Bearer "+db.token)

    dl, err := client.Do(req)
    if err != nil {
        return err
    }

    _, err = io.Copy(out, dl.Body)
    if err != nil {
        return err
    }

    out.Close()
    dl.Body.Close()
    return nil
}
