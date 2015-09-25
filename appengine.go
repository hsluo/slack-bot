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

var (
	BOT_TOKEN, HOOK_TOKEN string
	bot                   Bot
	botId, atId, alias    string
	loc                   *time.Location
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

	client := urlfetch.Client(c)

	channel := req.PostFormValue("channel_id")
	text := req.PostFormValue("text")

	if strings.Contains(text, "commit") {
		data := url.Values{
			"channel": {channel},
			"text":    {WhatTheCommit(client)},
		}
		bot.WithClient(client).ChatPostMessage(data)
	}
}

func warmUp(rw http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)
	client := urlfetch.Client(c)
	if bot.Token == "" {
		bot, err := NewBot(client, BOT_TOKEN)
		if err != nil {
			c.Errorf("%v", err)
		} else {
			c.Infof("new bot: %#v", bot)
		}
	}
}

func init() {
	log.Println("appengine init")
	HOOK_TOKEN, BOT_TOKEN = readCredentials("CREDENTIALS.appengine")

	http.HandleFunc("/hook", handleHook)
	http.HandleFunc("/_ah/warmup", warmUp)
}
