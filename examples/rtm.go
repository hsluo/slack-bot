// +build !appengine

package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hsluo/slack-bot"

	"golang.org/x/net/websocket"
)

var (
	token       string
	botId, atId string
	loc         *time.Location
)

func startServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("listening on " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleMessage(incoming <-chan slack.Message, outgoing chan<- slack.Message) {
	for msg := range incoming {
		if msg.Type != "message" {
			continue
		}
		if strings.Contains(msg.Text, "谢谢") {
			msg.Text = "不客气 :blush:"
			outgoing <- msg
		} else if isAt(msg.Text) {
			fields := strings.Fields(msg.Text)
			if len(fields) == 1 {
				sendCode(msg, outgoing)
			} else {
				var commit bool
				log.Println(fields)
				for _, f := range fields {
					if isAt(f) {
						continue
					}
					if strings.Contains(f, "commit") {
						commit = true
					}
				}
				if commit {
					sendCommitMessage(msg, outgoing)
				} else {
					if rand.Intn(2) > 0 {
						msg.Text = "呵呵"
					} else {
						msg.Text = "嘻嘻"
					}
					outgoing <- msg
				}
			}
		} else if isImage(msg) {
			sendCode(msg, outgoing)
		}
	}
}

func init() {
	log.Println("standalone init")
}

func main() {
	credentials, err := slack.LoadCredentials("credentials.json")
	wsurl, id := slack.RtmStart(credentials.Bot.Token)
	botId = id
	atId = "<@" + botId + ">"
	log.Println(wsurl, botId)

	ws, err := websocket.Dial(wsurl, "", "https://api.slack.com/")
	if err != nil {
		log.Fatal(err)
	}

	incoming := make(chan slack.Message)
	outgoing := make(chan slack.Message)

	go slack.RtmReceive(ws, incoming)
	go slack.RtmSend(ws, outgoing)
	go handleMessage(incoming, outgoing)

	startServer()
}
