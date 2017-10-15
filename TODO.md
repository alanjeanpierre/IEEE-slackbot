# To do

* ~~Fix internet cutout connection losses~~ (for now)
* ~~display full schedules~~
* better error handling
* ~~download files either by address (@bot) or by specific channel~~
* add delays for quick looping api queries so i don't get b&
* fix long term disconnects panicing
	* on disconnect, mybot tries to reconnect on an interval, starting a new websocket connection
	* however, the hearteat that detected the disconnect uses that websocket connection
	* if the new connection fails, it may return a nil pointer, so the heartbeat will have a nil pointer dereference

