// +build appengine

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"

	"github.com/hsluo/slack-bot"

	"google.golang.org/appengine"
	l "google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type task struct {
	context context.Context
	method  string
	data    url.Values
}

var (
	botId, atId string
	outgoing    chan task
)

func handleHook(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		return
	}

	token := req.PostFormValue("token")
	if token != credentials.HookToken {
		return
	}

	replyHook(req)
}

func replyHook(req *http.Request) {
	c := appengine.NewContext(req)
	l.Infof(c, "%v", req.Form)

	channel := req.PostFormValue("channel_id")
	text := req.PostFormValue("text")
	user_id := req.PostFormValue("user_id")

	client := urlfetch.Client(c)
	data := url.Values{"channel": {channel}}

	if strings.Contains(text, "commit") {
		data.Add("text", WhatTheCommit(client))
		outgoing <- task{context: c, method: "chat.postMessage", data: data}
	} else if strings.Contains(text, bot.User) ||
		strings.Contains(text, bot.UserId) {
		d1 := url.Values{"channel": {channel}, "text": {"稍等"}}
		outgoing <- task{context: c, method: "chat.postMessage", data: d1}

		text := codeWithAt(user_id)
		d2 := url.Values{"channel": {channel}, "text": {text}}
		outgoing <- task{context: c, method: "chat.postMessage", data: d2}
	} else if strings.Contains(text, "谢谢") {
		data.Add("text", "不客气 :blush:")
		outgoing <- task{context: c, method: "chat.postMessage", data: data}
	} else {
		if rand.Intn(2) > 0 {
			data.Add("text", "呵呵")
		} else {
			data.Add("text", "嘻嘻")
		}
		outgoing <- task{context: c, method: "chat.postMessage", data: data}
	}
}

func worker(outgoing chan task) {
	for task := range outgoing {
		_, err := bot.WithClient(urlfetch.Client(task.context)).PostForm(task.method, task.data)
		if err != nil {
			l.Errorf(task.context, "%s\n%v", err, task.data)
		}
	}
}

func standUpAlert(rw http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)
	url := credentials.SlackbotUrl
	if url == "" {
		l.Errorf(c, "no slackbot URL provided")
		return
	}
	url = fmt.Sprintf("%s&channel=%%23%s", url, "general")
	client := urlfetch.Client(c)
	client.Post(url, "text/plain", strings.NewReader("stand up"))
}

func replyCommit(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(rw, WhatTheCommit(urlfetch.Client(appengine.NewContext(req))))
}

func init() {
	log.Println("appengine init")
	outgoing = make(chan task)
	go worker(outgoing)

	http.HandleFunc("/hook", handleHook)
	http.HandleFunc("/alerts/standup", standUpAlert)
	http.HandleFunc("/cmds/whatthecommit",
		slack.ValidateCommand(http.HandlerFunc(replyCommit), credentials.Commands))
}

func main() {}
