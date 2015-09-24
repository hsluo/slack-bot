package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/websocket"
)

type Message struct {
	Type    string     `json:"type"`
	SubType string     `json:"subtype"`
	Channel string     `json:"channel"`
	User    string     `json:"user"`
	Text    string     `json:"text"`
	Ts      string     `json:"ts"`
	File    FileObject `json:"file"`
}

type FileObject struct {
	Mimetype   string `json:"mimetype"`
	Filetype   string `json:"filetype"`
	PrettyType string `json:"pretty_type"`
}

type Bot struct {
	token string
}

func (b *Bot) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	data.Add("token", b.token)
	return http.PostForm(url, data)
}

// Calls rtm.start API, return websocket url and bot id
func rtmStart(token string) (wsurl string, id string) {
	resp, err := http.PostForm("https://slack.com/api/rtm.start", url.Values{"token": {token}})
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	respRtmStart := make(map[string]interface{})
	err = json.Unmarshal(body, &respRtmStart)
	if err != nil {
		log.Fatal(err)
	}
	wsurl = respRtmStart["url"].(string)
	id = respRtmStart["self"].(map[string]interface{})["id"].(string)
	return
}

func rtmReceive(ws *websocket.Conn, incoming chan<- Message) {
	for {
		var m Message
		if err := websocket.JSON.Receive(ws, &m); err != nil {
			log.Println(err)
		} else {
			log.Printf("read %v", m)
			incoming <- m
		}
	}
}

func rtmSend(ws *websocket.Conn, outgoing <-chan Message) {
	for m := range outgoing {
		m.User = botId
		m.Ts = fmt.Sprintf("%f", float64(time.Now().UnixNano())/1000000000.0)
		log.Printf("send %v", m)
		if err := websocket.JSON.Send(ws, m); err != nil {
			log.Println(err)
		}
	}
}
