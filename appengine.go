// +build appengine

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"appengine"
	"appengine/urlfetch"
)

type task struct {
	context appengine.Context
	url     string
	data    url.Values
}

var (
	BOT_TOKEN, HOOK_TOKEN string
	bot                   Bot
	botId, atId, alias    string
	loc                   *time.Location
	outgoing              chan task
)

// load credentials from env
func loadCredentials(c appengine.Context) (hookToken, botToken string) {
	hookToken = os.Getenv("HOOK_TOKEN")
	botToken = os.Getenv("BOT_TOKEN")
	if hookToken == "" || botToken == "" {
		c.Errorf("%s", "cannot find credentials")
		os.Exit(1)
	}
	return
}

func handleHook(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		return
	}

	if HOOK_TOKEN == "" || BOT_TOKEN == "" {
		warmUp(rw, req)
	}

	token := req.PostFormValue("token")
	if token != HOOK_TOKEN {
		return
	}

	reply(req)
}

func reply(req *http.Request) {
	c := appengine.NewContext(req)
	c.Infof("%v", req.Form)

	channel := req.PostFormValue("channel_id")
	text := req.PostFormValue("text")
	user_id := req.PostFormValue("user_id")

	client := urlfetch.Client(c)
	data := url.Values{"channel": {channel}}

	if strings.Contains(text, "commit") {
		data.Add("text", WhatTheCommit(client))
		outgoing <- task{context: c, url: ChatPostMessageApi, data: data}
	} else if strings.Contains(text, bot.User) ||
		strings.Contains(text, bot.UserId) {
		d1 := url.Values{"channel": {channel}, "text": {"稍等"}}
		outgoing <- task{context: c, url: ChatPostMessageApi, data: d1}

		text := codeWithAt(user_id)
		d2 := url.Values{"channel": {channel}, "text": {text}}
		outgoing <- task{context: c, url: ChatPostMessageApi, data: d2}
	} else if strings.Contains(text, "谢谢") {
		data.Add("text", "不客气 :blush:")
		outgoing <- task{context: c, url: ChatPostMessageApi, data: data}
	} else {
		if rand.Intn(2) > 0 {
			data.Add("text", "呵呵")
		} else {
			data.Add("text", "嘻嘻")
		}
		outgoing <- task{context: c, url: ChatPostMessageApi, data: data}
	}
}

func worker(outgoing chan task) {
	for task := range outgoing {
		task.context.Infof("%v", task.data)
		_, err := bot.WithClient(urlfetch.Client(task.context)).PostForm(task.url, task.data)
		if err != nil {
			task.context.Errorf("%v", err)
		}
	}
}

func warmUp(rw http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)
	HOOK_TOKEN, BOT_TOKEN = loadCredentials(c)

	if bot.UserId == "" {
		client := urlfetch.Client(c)
		newbot, err := NewBot(client, BOT_TOKEN)
		if err != nil {
			c.Errorf("%v", err)
		} else {
			bot = newbot
			c.Infof("current bot: %#v", bot)
		}
	}
}

func standUpAlert(rw http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)
	url := os.Getenv("SLACKBOT_URL")
	if url == "" {
		c.Errorf("no slackbot URL provided")
		return
	}
	url = fmt.Sprintf("%s&channel=%%23%s", url, "general")
	client := urlfetch.Client(c)
	client.Post(url, "text/plain", strings.NewReader("stand up"))
}

type LogglyAlert struct {
	AlertName        string   `json:"alert_name"`
	AlertDescription string   `json:"alert_description"`
	EditAlertLink    string   `json:"edit_alert_link"`
	SourceGroup      string   `json:"source_group"`
	StartTime        string   `json:"start_time"`
	EndTime          string   `json:"end_time"`
	SearchLink       string   `json:"search_link"`
	Query            string   `json:"query"`
	NumHits          int      `json:"num_hits"`
	RecentHits       []string `json:"recent_hits"`
	OwnerUsername    string   `json:"owner_username"`
}

func logglyAlert(rw http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		c.Errorf("%s", err)
		return
	}
	if BOT_TOKEN == "" {
		warmUp(rw, req)
	}

	alert := LogglyAlert{}
	if err := json.Unmarshal(body, &alert); err != nil {
		c.Errorf("%s\n%s", err, string(body))
		return
	}

	fields := []Field{
		{
			Title: "Query",
			Value: alert.Query,
			Short: true,
		}, {
			Title: "Num Hits",
			Value: strconv.Itoa(alert.NumHits),
			Short: true,
		}, {
			Title: "Recent Hits",
			Value: strings.Join(alert.RecentHits, "\n"),
			Short: false,
		},
	}
	attachments := []Attachment{
		{
			Fallback:   alert.AlertName,
			Color:      "warning",
			Pretext:    alert.AlertDescription,
			AuthorName: alert.OwnerUsername,
			Title:      alert.AlertName,
			TitleLink:  alert.SearchLink,
			Text:       alert.AlertDescription,
			Fields:     fields,
		},
	}
	bytes, err := json.Marshal(attachments)
	if err != nil {
		c.Errorf("%s", err)
		return
	}
	data := url.Values{}
	data.Add("channel", "#loggly")
	data.Add("attachments", string(bytes))
	data.Add("as_user", "false")
	outgoing <- task{context: c, url: ChatPostMessageApi, data: data}
}

func init() {
	log.Println("appengine init")
	outgoing = make(chan task)
	go worker(outgoing)

	http.HandleFunc("/hook", handleHook)
	http.HandleFunc("/alerts/standup", standUpAlert)
	http.HandleFunc("/loggly", logglyAlert)
}
