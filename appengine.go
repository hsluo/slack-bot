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

	channel := req.PostFormValue("channel_id")
	text := req.PostFormValue("text")

	client := urlfetch.Client(c)
	data := url.Values{}

	if strings.Contains(text, "commit") {
		data.Add("channel", channel)
		data.Add("text", WhatTheCommit(client))
		c.Infof("%v", data)
		err := bot.WithClient(client).ChatPostMessage(data)
		if err != nil {
			c.Errorf("%v", err)
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

	http.HandleFunc("/hook", handleHook)
	http.HandleFunc("/_ah/warmup", warmUp)
}
