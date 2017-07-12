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
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"strconv"
	"golang.org/x/net/websocket"
	)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "usage: mybot slack-bot-token admin-user log/file/location/\n")
		os.Exit(1)
	}

	// start a websocket-based Real Time API session
	ws, id := slackConnect(os.Args[1])
	fmt.Println("mybot ready, ^C exits")
	
	boss := os.Args[2]
	logloc := os.Args[3]

	for {
		// read each incoming message
		m, err := getMessage(ws)
		if err != nil {
			death(m, ws)
			log.Fatal(err)
		}

		
		if m.Type == "message" {
			file, err := os.OpenFile(logloc+m.Channel+".txt", os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0664)
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
			//fmt.Println(m.Channel)
			fmt.Fprintf(file, "%s, %v, \"%q\"\n", time.Unix(tms, tns), m.User, m.Text)
			//fmt.Fprintf(file, 
			
			err = file.Close()
			if err != nil {
				death(m, ws)
				log.Fatal(err)
			}
			
			
			// if bot is mentioned
			if strings.HasPrefix(m.Text, "<@"+id+">") {
			
				// elevated priviledges?
				if m.User == boss {
				
				}
				// if so try to parse if
				parts := strings.Fields(m.Text)
				if len(parts) == 3 && parts[1] == "stock" {
					// looks good, get the quote and reply with the result
					go func(m Message) {
						m.Text = getQuote(parts[2])
						postMessage(ws, m)
					}(m)
					// NOTE: the Message object is copied, this is intentional
				} else {
					// huh?
					m.Text = fmt.Sprintf("sorry, that does not compute\n")
					postMessage(ws, m)
				}
			}
		}
		
	}
}

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
