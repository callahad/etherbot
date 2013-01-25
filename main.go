package main

import (
	"code.google.com/p/cookiejar"
	"errors"
	"flag"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"net/http"
	"regexp"
)

var host *string = flag.String("host", "irc.mozilla.org", "IRC server")
var channel *string = flag.String("channel", "#foo", "IRC channel")
var nick *string = flag.String("nick", "etherbot", "IRC nickname")

func isPrivate(s string) bool {
	client := &http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			if r.URL.Path == "/ep/account/sign-in" {
				return errors.New("Asked to sign in. Pad is not public.")
			}
			return nil
		},
		Jar: cookiejar.NewJar(false),
	}

	_, err := client.Get(s)

	return (err != nil)
}

func main() {
	flag.Parse()

	// create new IRC connection
	c := irc.SimpleClient(*nick, "etherbot", "the etherpad robot")
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
				re := regexp.MustCompile(`https?://id\.etherpad\.mozilla\.org/\S+`)
				pad := re.FindString(line.Args[1])
				if len(pad) > 0 {
					go func() {
						if isPrivate(pad) {
							c.Privmsg(line.Args[0], line.Nick+": Please make sure that etherpad is public. Thanks!")
						}
					}()
				}
			}
		})

	if err := c.Connect(*host); err != nil {
		fmt.Printf("Connection error: %s\n", err)
		return
	}

	<-quit
}
