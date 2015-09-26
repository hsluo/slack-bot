// +build appengine

package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

func readCredentials(file string) (hookToken, botToken string) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")
	hookToken, botToken = lines[0], lines[1]
	log.Println(hookToken, botToken)
	return
}

func handleHook(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		return
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
		outgoing <- task{
			context: c,
			url:     ChatPostMessageApi,
			data:    data,
		}
	} else if strings.Contains(text, bot.User) || strings.Contains(text, bot.UserId) {
		d1 := url.Values{"channel": {channel}, "text": {"稍等"}}
		outgoing <- task{
			context: c,
			url:     ChatPostMessageApi,
			data:    d1,
		}
		d2 := url.Values{"channel": {channel}, "text": {"1024 <@" + user_id + ">"}}
		outgoing <- task{
			context: c,
			url:     ChatPostMessageApi,
			data:    d2,
		}
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
	client := urlfetch.Client(c)
	if bot.Token == "" {
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
	HOOK_TOKEN, BOT_TOKEN = readCredentials("CREDENTIALS.appengine")
	outgoing = make(chan task)
	go worker(outgoing)

	http.HandleFunc("/hook", handleHook)
	http.HandleFunc("/_ah/warmup", warmUp)
}
