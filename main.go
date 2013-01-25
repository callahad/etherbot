package main

import (
	"flag"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"regexp"
)

var host *string = flag.String("host", "irc.mozilla.org", "IRC server")
var channel *string = flag.String("channel", "#foo", "IRC channel")

func main() {
	flag.Parse()

	// create new IRC connection
	c := irc.SimpleClient("etherbot", "etherbot", "the etherpad robot")
	// c.EnableStateTracking()
	c.AddHandler("connected",
		func(conn *irc.Conn, line *irc.Line) { conn.Join(*channel) })

	// Set up a handler to notify of disconnect events.
	quit := make(chan bool)
	c.AddHandler("disconnected",
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	// Set up handlers for incoming messages
	c.AddHandler("PRIVMSG",
		func(conn *irc.Conn, line *irc.Line) {
			if match, err := regexp.MatchString(`https?://id\.etherpad\.mozilla\.org/.`, line.Args[1]); match && err == nil {
				c.Privmsg(line.Args[0], line.Nick + ": Please make sure that etherpad is public. Thanks!")
			}
		})

	if err := c.Connect(*host); err != nil {
		fmt.Printf("Connection error: %s\n", err)
		return
	}

	<-quit
}
