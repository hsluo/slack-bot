// +build !appengine

package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

var (
	token              string
	botId, atId, alias string
	loc                *time.Location
)

func readCredentials(file string) (token, alias string) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")
	token, alias = lines[0], lines[1]
	log.Println(token, alias)
	return
}

func startServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("listening on " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func init() {
	log.Println("standalone init")
	log.SetFlags(log.Lshortfile)
	token, alias = readCredentials("CREDENTIALS")
}

func main() {
	wsurl, id := RtmStart(token)
	botId = id
	atId = "<@" + botId + ">"
	if alias == "" {
		alias = atId
	} else {
		alias = "@" + alias
	}
	log.Println(wsurl, botId)

	ws, err := websocket.Dial(wsurl, "", "https://api.slack.com/")
	if err != nil {
		log.Fatal(err)
	}

	incoming := make(chan Message)
	outgoing := make(chan Message)

	go RtmReceive(ws, incoming)
	go RtmSend(ws, outgoing)
	go handleMessage(incoming, outgoing)

	startServer()
}
