package main

import (
	"encoding/json"
	"errors"
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
	Token, UserId, User string
	Client              *http.Client
}

func NewBot(c *http.Client, token string) (b Bot, err error) {
	resp, err := c.PostForm("https://slack.com/api/auth.test", url.Values{"token": {token}})
	if err != nil {
		return
	}
	respAuthTest, err := asJson(resp)
	if err != nil {
		return
	} else if !respAuthTest["ok"].(bool) {
		err = errors.New(respAuthTest["error"].(string))
		return
	} else {
		b = Bot{
			Token:  token,
			UserId: respAuthTest["user_id"].(string),
			User:   respAuthTest["user"].(string),
		}
	}
	return
}

func (b Bot) WithClient(c *http.Client) Bot {
	b.Client = c
	return b
}

func (b Bot) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	data.Add("token", b.Token)
	data.Add("as_user", "true")

	if b.Client == nil {
		b.Client = http.DefaultClient
	}
	resp, err = b.Client.PostForm(url, data)
	if err != nil {
		return
	}
	respJson, err := asJson(resp)
	if err != nil {
		return
	}
	if !respJson["ok"].(bool) {
		err = errors.New(respJson["error"].(string))
	}
	return
}

func (b Bot) ChatPostMessage(data url.Values) (err error) {
	_, err = b.PostForm("https://slack.com/api/chat.postMessage", data)
	return
}

func asJson(resp *http.Response) (m map[string]interface{}, err error) {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	m = make(map[string]interface{})
	err = json.Unmarshal(body, &m)
	return
}

// Calls rtm.start API, return websocket url and bot id
func RtmStart(token string) (wsurl string, id string) {
	resp, err := http.PostForm("https://slack.com/api/rtm.start", url.Values{"token": {token}})
	if err != nil {
		log.Fatal(err)
	}
	respRtmStart, err := asJson(resp)
	if err != nil {
		log.Fatal(err)
	}
	wsurl = respRtmStart["url"].(string)
	id = respRtmStart["self"].(map[string]interface{})["id"].(string)
	return
}

func RtmReceive(ws *websocket.Conn, incoming chan<- Message) {
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

func RtmSend(ws *websocket.Conn, outgoing <-chan Message) {
	for m := range outgoing {
		m.Ts = fmt.Sprintf("%f", float64(time.Now().UnixNano())/1000000000.0)
		log.Printf("send %v", m)
		if err := websocket.JSON.Send(ws, m); err != nil {
			log.Println(err)
		}
	}
}
