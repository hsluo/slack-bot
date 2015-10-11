// +build appengine

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hsluo/slack-bot"

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
	bot                   slack.Bot
	botId, atId, alias    string
	loc                   *time.Location
	outgoing              chan task
)

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
		outgoing <- task{context: c, url: slack.ChatPostMessageApi, data: data}
	} else if strings.Contains(text, bot.User) ||
		strings.Contains(text, bot.UserId) {
		d1 := url.Values{"channel": {channel}, "text": {"稍等"}}
		outgoing <- task{context: c, url: slack.ChatPostMessageApi, data: d1}

		text := codeWithAt(user_id)
		d2 := url.Values{"channel": {channel}, "text": {text}}
		outgoing <- task{context: c, url: slack.ChatPostMessageApi, data: d2}
	} else if strings.Contains(text, "谢谢") {
		data.Add("text", "不客气 :blush:")
		outgoing <- task{context: c, url: slack.ChatPostMessageApi, data: data}
	} else {
		if rand.Intn(2) > 0 {
			data.Add("text", "呵呵")
		} else {
			data.Add("text", "嘻嘻")
		}
		outgoing <- task{context: c, url: slack.ChatPostMessageApi, data: data}
	}
}

func worker(outgoing chan task) {
	for task := range outgoing {
		_, err := bot.WithClient(urlfetch.Client(task.context)).PostForm(task.url, task.data)
		if err != nil {
			task.context.Errorf("%s\n%v", err, task.data)
		}
	}
}

func warmUp(rw http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)

	if bot.UserId == "" {
		client := urlfetch.Client(c)
		newbot, err := slack.NewBot(client, slack.Creds.BotToken)
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

func logglyAlert(rw http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)

	attachment, err := NewAttachment(req)
	if err != nil {
		c.Errorf("%s", err)
		return
	}

	if BOT_TOKEN == "" {
		warmUp(rw, req)
	}
	bytes, err := json.Marshal([]slack.Attachment{attachment})
	if err != nil {
		c.Errorf("%s", err)
		return
	}
	data := url.Values{}
	data.Add("channel", "#loggly")
	data.Add("attachments", string(bytes))
	data.Add("as_user", "false")
	outgoing <- task{context: c, url: slack.ChatPostMessageApi, data: data}
}

func replyCommit(rw http.ResponseWriter, req *http.Request) {
	if !slack.ValidateCommand(req) {
		return
	}
	fmt.Fprintln(rw, WhatTheCommit(urlfetch.Client(appengine.NewContext(req))))
}

func init() {
	log.Println("appengine init")
	outgoing = make(chan task)
	go worker(outgoing)

	http.HandleFunc("/hook", handleHook)
	http.HandleFunc("/alerts/standup", standUpAlert)
	http.HandleFunc("/loggly", logglyAlert)
	http.HandleFunc("/cmds/whatthecommit", replyCommit)
}
