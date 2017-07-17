
# mybot

`mybot` is an working Slack bot written in Go. Fork it and use it to build
your very own cool Slack bot!

Check the [blog post](https://www.opsdash.com/blog/slack-bot-in-golang.html)
for a description of mybot internals.

This particular implementation logs all messages in channels the bot is a part of.
It tries to find the name of the channel and will name the log file after the channel,
but if it can't find it (as in the case of DMs or private channels) it will prefix the log
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
Will link to the readme on github

`@bot-name stock stock-to-look-up` 
Will lookup through Yahoo finance what the current stock is

`@bot-name wiki challenge`
Will get 2 random wiki pages for the wiki game: https://en.wikipedia.org/wiki/Wikipedia:Wiki_Game

`@bot-name links get`
The bot will return a random link from its list of links in the /root/directory/links file

`@bot-name links add link-to-add`
Adds a link to the /root/directory/links file.

## Administrative use
`@bot-name cleanup`
The admin user can request the bot to save its user and channel translation
tables to file, which the bot tries to read from on boot

`@bot-name ban user-name`
The admin user can ban a user from interacting with the bot. The list of banned users is maintained in the root/directory/
Probably best to use in a DM

`@bot-name unban user-name`
Unbans, if they exist