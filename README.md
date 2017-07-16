
# mybot

`mybot` is an working Slack bot written in Go. Fork it and use it to build
your very own cool Slack bot!

Check the [blog post](https://www.opsdash.com/blog/slack-bot-in-golang.html)
for a description of mybot internals.

This particular implementation logs all messages in channels the bot is a part of.
It also gets stock, random wikipedia pages for games, and maintains a file 
that users can add links or whatever they want to.

## Usage
Compile mybot.go and slack.go and run with:
```mybot api-token admin-user-name /path/to/root/directory/```
The bot will create all files it needs in the root or subdirectories within it. Mind the trailing slash.