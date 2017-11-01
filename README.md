
# mybot

`mybot` is an working Slack bot written in Go. Fork it and use it to build
your very own cool Slack bot!

Check the [blog post](https://www.opsdash.com/blog/slack-bot-in-golang.html)
for a description of mybot internals.

This particular implementation logs all messages in channels the bot is a part of to an sqlite3 database as well as external files (until deprecation later).
It tries to find the name of the channel and will name the log file and SQL column after the channel,
but if it can't find it (as in the case of DMs or private groups) it will prefix the log
file name with `Private -` and the channel ID.
It also gets stock, random wikipedia pages for games, and maintains a file 
that users can add links or whatever they want to.



## Usage
Compile mybot.go and slack.go and run with:
```mybot api-token admin-user-name /path/to/root/directory/```
The bot will create all files it needs in the root or subdirectories within it. Mind the trailing slash.

## Live Usage
This bot responds to a few commands

`@bot-name help`
* Will link to the readme on github

`@bot-name status`
* Returns uptime and number of messages read

`@bot-name stock stock-to-look-up` 
* Will lookup through Yahoo finance what the current stock is

`@bot-name wiki challenge`
* Will get 2 random wiki pages for the wiki game: https://en.wikipedia.org/wiki/Wikipedia:Wiki_Game

`@bot-name links get`
* The bot will return a random link from its list of links in the /root/directory/links file

`@bot-name links add link-to-add`
* Adds a link to the /root/directory/links file.

`@bot-name watson parameters,parameters text`
* Queries [Watson Tone Analyzer](https://watson-api-explorer.mybluemix.net/apis/tone-analyzer-v3#!/tone/GetTone)

`@bot-name watson help`
* Prints help for using Watson's tone analyzer

`@bot-name schedule free|busy times`
* Describes your schedule to the bot, which stores it for calculating maximum meeting times.
* Only accepts days m-f 8:00-18:00
* Monday, Tuesday, Wednesday, thuRsday, Friday
* Free means you are free during the listed times, busy otherwise. Busy means you are busy during the listed times, free otherwise.
* example: `@bot-name schedule free m 8 t 8 9 10 w 13 14 r 0 f 0`

`@bot-name meeting`
* Returns the best meeting time based on the amount of people expected to attend from their schedules.

`@bot-name poll n`
* Constructs a reaction-based poll from 1 - n, max 10

`@bot-name`
* If the title of a file is the bots name, the bot will download it to the ./files directory

`@bot-name remindme n seconds|minutes|hours|days message`
* Responds after the specified delay the same message
* For time lengths, will accept singular versions and single letter versions
* example: `@bot-name remindme 10 m Did you turn off the stove?`
* Currently does not do fractional units. If you want 6 hours and 30 minutes, you'll have to say 390 minutes

`@bot-name remindall n seconds|minutes|hours|days message`
* Alerts channel after the specified delay the same message
---

## Administrative use
`@bot-name cleanup`
* The admin user can request the bot to save its user and channel translation tables to file, which the bot tries to read from on boot

`@bot-name ban user-name`
* The admin user can ban a user from interacting with the bot. The list of banned users is maintained in the root/directory/
* Probably best to use in a DM

`@bot-name unban user-name`
* Unbans, if they exist
