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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"strconv"
	"golang.org/x/net/websocket"
	"bufio"
	"math/rand"
	"bytes"
	)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "usage: mybot slack-bot-token admin-user /root/working/directory \n")
		os.Exit(1)
	}

	// start a websocket-based Real Time API session
	ws, id := slackConnect(os.Args[1])
	fmt.Println("mybot ready, ^C exits")
	
	boss := os.Args[2]
	rootloc := os.Args[3]

	user_lookup := make(map[string]string)
	channel_lookup := make(map[string]string)
	banlist := make(map[string]bool)
	
	loadmap(user_lookup, channel_lookup, banlist)
	
	elevated := false
	
	for {
		// read each incoming message
		m, err := getMessage(ws)
		if err != nil {
			death(m, ws)
			log.Fatal(err)
		}

		
		if m.Type == "message" && m.Text != "" {
		
			// identify user from ID to readable string
			usr, ok := user_lookup[m.User]
			if !ok {
				usr = findUser(m.User, os.Args[1])
				user_lookup[m.User] = usr
			}
			
			bannedyn, ok := banlist[usr]
			if bannedyn == true && usr != boss {
				continue
			}
				
			// identify channel from ID to readable string
			// Private channels and DMs don't show up though :/
			channel, ok := channel_lookup[m.Channel]
			if !ok {
				channel = findChannel(m.Channel, os.Args[1])
				channel_lookup[m.Channel] = channel
			
			}
			
			// create a log file named after the channel
			// and log the message 
			file, err := os.OpenFile(rootloc+"logs/"+channel+".txt", os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0664)
			if err != nil {
				death(m, ws)
				log.Fatal(err)
			}
			times := strings.Split(m.TS, ".")
			if len(times) != 2 {
				continue
			}
			
			tms, err := strconv.ParseInt(times[0], 10, 64)
			if err != nil {
				death(m, ws)
				log.Fatal(err)
			}
			
			tns, err := strconv.ParseInt(times[1], 10, 64)
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
			
			if (usr == boss) {
				elevated = true
			} else {
				elevated = false
			}
			
			
			// if bot is mentioned
			if strings.HasPrefix(m.Text, "<@"+id+">") {
			
				// if so try to parse if
				parts := strings.Fields(m.Text)
				
				if len(parts) >= 2 {
					switch cmd := parts[1]; cmd {
					// admin commands
					case "ban":
						if elevated && len(parts) == 3 {
							banned_user := parts[2]
							file, err := os.OpenFile(rootloc+"banlist", os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0664)
							if err != nil {
								continue
							}
							fmt.Fprintf(file, "%s\n", banned_user)
							file.Close()
							banlist[banned_user] = true
							go func (m Message) {
								m.Text = fmt.Sprintf("Ok, I banned %s", banned_user)
								postMessage(ws, m)
							}(m)
						}
						
					case "unban":
						if elevated && len(parts) == 3 {
							banned_user := parts[2]
							if banlist[banned_user] != true {
							
								go func (m Message) {
									m.Text = fmt.Sprintf("%s is not banned", banned_user)
									postMessage(ws, m)
								}(m)
								continue
							}
							
							banlist[banned_user] = false
							
							blist, err := ioutil.ReadFile(rootloc+"banlist")
							if err != nil {
								continue
							}
							newblist := bytes.Replace(blist, []byte(banned_user+"\n"), []byte(""), 1)
							err = ioutil.WriteFile(rootloc+"banlist", newblist, 0664)
							go func (m Message) {
								m.Text = fmt.Sprintf("Ok, I unbanned %s", banned_user)
								postMessage(ws, m)
							}(m)
						}
						
					case "cleanup":
						if elevated {
							savemap(user_lookup, channel_lookup, rootloc)
							go func (m Message) {
								m.Text = fmt.Sprintf("Ok, I saved everything")
								postMessage(ws, m)
							}(m)
						}
					
					// normal user commands
					case "stock":
						if len(parts) == 3 {
							// looks good, get the quote and reply with the result
							go func(m Message) {
								m.Text = getQuote(parts[2])
								postMessage(ws, m)
							}(m)
							// NOTE: the Message object is copied, this is intentional
							
						}
						
					case "wiki": 
						if len(parts) == 3 && parts[2] == "challenge" {
							go wikichall(m, ws)
						}
						
					case "help":
						go func(m Message) {
							m.Text = "You can view my readme here: https://github.com/alanjeanpierre/IEEE-slackbot/blob/master/README.md"
							//m.Text = fmt.Sprintf("You can view my readme here: %s\n", "https://github.com/alanjeanpierre/IEEE-slackbot/blob/master/README.md")
							postMessage(ws, m)
						}(m)
						
					case "links":
						if len(parts) == 4 && parts[2] == "add" {
							link := parts[3]
							link = link[1:len(link)-1]
							file, err := os.OpenFile(rootloc+"links", os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0664)
							if err != nil {
								continue
							}
							fmt.Fprintf(file, "%s\n", link)
							file.Close()
											
						} else if len(parts) == 3 && parts[2] == "get" {
							// reads in the links file and sends a randomly selected link
							go func(m Message) {
								links, err := readLines(rootloc+"links")
								if err != nil || len(links) == 0 {
									return
								}
								index := rand.Intn(len(links))
								m.Text = "Enjoy! " + links[index]
								postMessage(ws, m)
							} (m)
						} else {
							// huh?
							m.Text = fmt.Sprintf("sorry, that does not compute. @onee-sama links add|get\n")
							postMessage(ws, m)
						}
					default: 
						m.Text = fmt.Sprintf("sorry, that does not compute\n")
						postMessage(ws, m)
					}
				}
			}
		}
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
	ID int	`json:"id"`
	NS int `json:"ns"`
	Title string `json:"title"`
}

// wikipedia query
type query struct {
	Random[] random `json:"random"`
}

// more wiki structs
type wikiresp struct {
	
	Query query `json:"query"`

}

type SlackUserObject struct {
	Ok	bool	`json:"ok"`
	User User	`json:"user"`
}

type SlackChannelObject struct {
	Ok	bool	`json:"ok"`
	Channel Channel	`json:"channel"`
}

type User struct {
	Id		string `json:"id"`
	Name	string `json:"name"`
}

type Channel struct {
	Id		string `json:"id"`
	Name	string `json:"name"`
}

// load the users and channels from the files
func loadmap(user map[string]string, channel map[string]string, banlist map[string]bool) {
	
	file, err := os.OpenFile("usrs", os.O_RDONLY, 0664)
		if err == nil {
		
			scanner := bufio.NewScanner(file)
			
			for scanner.Scan() {
				
				users := strings.Fields(scanner.Text())
				user[users[0]] = users[1]
			}
		}
	
	file, err = os.OpenFile("channels", os.O_RDONLY, 0664)
		if err == nil {
		
			scanner := bufio.NewScanner(file)
			
			for scanner.Scan() {
				
				channels := strings.Fields(scanner.Text())
				channel[channels[0]] = channels[1]
			}
		}
	
	file, err = os.OpenFile("banlist", os.O_RDONLY, 0664)
		if err == nil {
		
			scanner := bufio.NewScanner(file)
			
			for scanner.Scan() {
				
				banned := scanner.Text()
				banlist[banned] = true
			}
		}
	
	


}

// save the loaded maps to disk
func savemap(user map[string]string, channel map[string]string, location string) {
	file, err := os.OpenFile(location+"usrs", os.O_WRONLY | os.O_CREATE, 0664)
		if err != nil {
			log.Fatal(err)
		}
	for key, value := range user {
		fmt.Fprintf(file, "%s %s\n", key, value)
	}
	file.Close()
	
	file, err = os.OpenFile(location+"channels", os.O_WRONLY | os.O_CREATE, 0664)
		if err != nil {
			log.Fatal(err)
		}
	for key, value := range channel {
		fmt.Fprintf(file, "%s %s\n", key, value)
	}
	file.Close()
}

// queries slack api for channel name
func findChannel(channel string, token string) (string) {

	
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
	} else {
		return "Private - " + channel
	}
	

}

// queries slack api for user name from ID
func findUser(usr string, token string) (string) {
	url := fmt.Sprintf("https://slack.com/api/users.info?token=%s&user=%s", token, usr)
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



// queries wiki api for 2 random pages
func wikichall(m Message, ws *websocket.Conn) (Message) {
	resp, err := http.Get("https://en.wikipedia.org/w/api.php?action=query&format=json&list=random&rnnamespace=0&rnlimit=2")
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		return m
	}
	
	body, err := ioutil.ReadAll(resp.Body)
	
	resp.Body.Close()
	if err != nil {
		return m
	}
	
	var response wikiresp 
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("I died")
		return m
	}
	
	page1:=strings.Replace(response.Query.Random[0].Title, " ", "_", -1)
	page2:=strings.Replace(response.Query.Random[1].Title, " ", "_", -1)
	
	m.Text = fmt.Sprintf("Try to get from https://en.wikipedia.org/wiki/%s to https://en.wikipedia.org/wiki/%s using only the links on the page!", page1, page2)

	postMessage(ws, m)
	return m

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
