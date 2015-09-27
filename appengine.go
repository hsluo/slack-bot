// +build appengine

package main

import (
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
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

func init() {
	log.Println("appengine init")
	outgoing = make(chan task)
	go worker(outgoing)

	http.HandleFunc("/hook", handleHook)
}
