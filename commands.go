// parse cmd should execute the command
// execution: pass arg string to work with
// commands should do what they say, and return errors
// parse cmd or caller should act on those errors, NOT the functions

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func parsecmd(db *Database, m Message) string {

	parts := strings.Fields(m.Text)

	if len(parts) < 2 { // no commands
		return "... Yes?"
	}
	switch cmd := strings.ToLower(parts[1]); cmd {
	// admin commands
	case "ban":
		return ban(db, m)
	case "unban":
		return unban(db, m)
	case "cleanup":
		return cleanup(db, m)

	// normal user commands
	case "stock":
		if len(parts) == 3 {
			return (getQuote(parts[2]))
		} else {
			return "Invalid arguments for stock"
		}

	case "wiki":
		if len(parts) == 3 && parts[2] == "challenge" {
			wikileanks := wikichall()
			var r string
			if wikileanks == "" { // err from wiki api
				r = fmt.Sprintf("wiki failed us. zip. nada. nil. Try again?")
			} else {
				r = fmt.Sprintf("%s", wikileanks)
			}
			return r
		} else {
			return "Invalid wiki syntax"
		}

	case "help":
		rdme := "You can view my readme here: https://github.com/alanjeanpierre/IEEE-slackbot/blob/master/README.md"
		return rdme

	case "links":
		if len(parts) == 4 && parts[2] == "add" {
			return addlink(db, m.User, strings.TrimSpace(parts[3]))
		} else if len(parts) == 3 && parts[2] == "get" {
			// reads in the links file and sends a randomly selected link
			return getlink(db)
		} else {
			return fmt.Sprintf("sorry, that does not compute. @onee-sama links add|get\n")
		}

	case "status":
		return fmt.Sprintf("I have been running for %v and have read %d messages", time.Since(db.uptime), db.nmsg)

	case "watson":
		if len(parts) > 4 {
			text := m.Text[3:]
			return fmt.Sprintf("\n%s", watsonToneAnalyzer(text, parts[2]))
		} else if len(parts) == 3 && parts[2] == "help" {
			return "call @onee-sama watson tones sentence\nAvailable tones are emotion, language, social and those wanted should be written like emotion,language,social"
		} else {
			return "sorry, that does not compute. Try @onee-sama watson help"
		}

	case "poll":
		if len(parts) >= 3 {
			num := []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "keycap_ten"}
			n, err := strconv.Atoi(parts[2])
			if err != nil || n > 10 {
				return "I have received an invalid poll"
			}
			for i := 1; i <= n; i++ {
				postReaction(db.token, m.Channel, m.TS, num[i])
                time.Sleep(time.Second)
			}
		}
		return ""
	case "schedule":
		if len(parts) > 3 {
			free := parts[2] == "free"
			text := strings.ToLower(strings.TrimPrefix(m.Text, strings.Join(parts[:3], " ")+" "))
			err := readSchedule(text, free, db.rootloc, m.User)
			if err != nil {
				return "error, bad scheduling"
			}
			return "thanks, got your schedule"
		}
	case "meeting":
		day, time, n := bestTime(db.rootloc)
		return fmt.Sprintf("the best time to meet is %s at %d:00, %d people should be in attendance", day, time, n)
	case "meetings":
		return allTimes(db.rootloc)
	case "remindme":
		return remindmeSetup(m, db)
	case "remindall":
		return remindmeSetup(m, db)
    case "addreaction":
        return addreaction(m, db)
	default:
		return "sorry that does not compute"
	}

	return "sorry that does not compute"
}
func ban(db *Database, m Message) string {

	if !db.isElevated(m.User) {
		r := "You aint the boss of me http://alanjeanpierre.hopto.org/bs/vanned.png"
		return r
	}

	parts := strings.Fields(m.Text)
	if len(parts) == 3 {
		bannedUser := parts[2]
		file, err := os.OpenFile(db.rootloc+"banlist", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
		if err != nil {
			log.Println("Can't open banlist")
			return "whoops, can't get the banlist"
		}
		fmt.Fprintf(file, "%s\n", bannedUser)
		file.Close()
		db.banlist[bannedUser] = true
		return fmt.Sprintf("Ok, I banned %s", bannedUser)
	}

	return "ban who?"

}

func unban(db *Database, m Message) string {

	if !db.isElevated(m.User) {
		r := "You aint the boss of me http://alanjeanpierre.hopto.org/bs/vanned.png"
		return r
	}

	parts := strings.Fields(m.Text)
	if len(parts) == 3 {
		bannedUser := parts[2]
		if db.banlist[bannedUser] != true {
			return fmt.Sprintf("%s is not banned", bannedUser)
		}

		db.banlist[bannedUser] = false

		blist, err := ioutil.ReadFile(db.rootloc + "banlist")
		if err != nil {
			log.Println(err)
			return "whoops, can't get the banlist"
		}
		newblist := bytes.Replace(blist, []byte(bannedUser+"\n"), []byte(""), 1)
		err = ioutil.WriteFile(db.rootloc+"banlist", newblist, 0664)
		if err != nil {
			log.Println(err)
			return "whoops, can't save the banlist"
		}
		return fmt.Sprintf("Ok, I unbanned %s", bannedUser)
	} else {
		return "Unban who?"
	}
}

func cleanup(db *Database, m Message) string {
	if !db.isElevated(m.User) {
		return "You aint the boss of me http://alanjeanpierre.hopto.org/bs/vanned.png"
	}

	db.save()
	return "Ok, I saved everything"
}

func getlink(db *Database) string {

	//links, err := readLines(db.rootloc + "links")
	var link string
	err := db.db.QueryRow("select link from links order by random() limit 1").Scan(&link)
	if err != nil {
		log.Println(err)
		return "uh oh, no links"
	}
	return fmt.Sprintf("enjoy! %s", link)
}

func addlink(db *Database, uid, link string) string {
	link = strings.Trim(link, " <>")
	file, err := os.OpenFile(db.rootloc+"links", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	err2 := db.insertLink(uid, link)
	if err != nil || err2 != nil {
		log.Println("Unable to open links file")
		return "sorry, can't seem to access the links right now"
	}
	fmt.Fprintf(file, "%s\n", link)
	file.Close()

	return "thanks for the link!"
}

func remindmeSetup(m Message, db *Database) string {
	parts := strings.Fields(m.Text)

	if len(parts) < 5 {
		return "remind you of what?"
	}

	text := strings.Join(parts[4:], " ")
	iN, err := strconv.Atoi(parts[2])
	if err != nil {
		return "invalid multiplier"
	}
	n := time.Duration(iN)
	sTime := strings.ToLower(parts[3])
	var t time.Duration

	switch sTime {
	case "s", "second", "seconds":
		t = n * time.Second
	case "m", "minute", "minutes":
		t = n * time.Minute
	case "h", "hour", "hours":
		t = n * time.Hour
	case "d", "day", "days":
		t = n * time.Hour * 24
	default:
		return "invalid time"
	}
	message := fmt.Sprintf("<@%s> %s", m.User, text)
	if parts[1][len(parts[1])-3:] == "all" {
		message = fmt.Sprintf("<!channel> %s", text)
	}
	go remindme(message, t, m, db)
	return fmt.Sprintf("ok, I'll remind you at %v", time.Now().Add(t))

}

func remindme(message string, duration time.Duration, m Message, db *Database) {
	time.Sleep(duration)
	respond(message, m, db.ws)
}

func addreaction(m Message, db *Database) string {
    parts := strings.Split(m.Text, "\"")
    
    if len(parts) != 3 {
        return "not enough arguments, need 2"
    }
    
    trigger := parts[1]
    reaction := strings.Trim(parts[2], " :")
    err := db.addReaction(m.User, trigger, reaction)
    if err != nil {
        log.Println(err)
        return "uh oh, that didn't work. Try again?"
    }
    return fmt.Sprintf("ok, I'll react with :%s: to \"%s\"", reaction, trigger)
}