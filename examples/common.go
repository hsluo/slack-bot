package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/hsluo/slack-bot"
)

var (
	credentials slack.Credentials
	bot         slack.Bot
)

func sendCommitMessage(m slack.Message, outgoing chan<- slack.Message) {
	m.Text = WhatTheCommit(nil)
	outgoing <- m
}

func ack() (ret string) {
	ret = "稍等"
	if rand.Intn(2) > 0 {
		ret += "，刚看到"
	}
	h := time.Now().In(loc).Hour()
	if h >= 18 && h <= 20 {
		ret += "，我在地铁上"
	}
	return
}

func codeWithAt(userId string) (ret string) {
	code := rand.Intn(9000) + 1000
	if rand.Intn(2) > 0 {
		ret = fmt.Sprintf("%d <@%s>", code, userId)
	} else {
		ret = fmt.Sprintf("<@%s> %d", userId, code)
	}
	return
}

func sendCode(m slack.Message, outgoing chan<- slack.Message) {
	m.Text = ack()
	outgoing <- m

	time.Sleep(1 * time.Second)

	m.Text = codeWithAt(m.User)
	outgoing <- m
}

func isImage(m slack.Message) bool {
	return m.SubType == "file_share" &&
		strings.HasPrefix(m.File.Mimetype, "image")
}

// at in the middle of the message is not supported
func isAt(text string) bool {
	return strings.HasPrefix(text, atId) || strings.HasSuffix(text, atId) ||
		strings.HasPrefix(text, alias) || strings.HasSuffix(text, alias)
}

func init() {
	log.SetFlags(log.Lshortfile)
	log.Println("common init")
	loc, _ = time.LoadLocation("Asia/Shanghai")
	rand.Seed(time.Now().Unix())
	var err error
	err = slack.LoadCredentials("credentials.json")
	if err != nil {
		log.Fatal(err)
	}
	bot = credentials.Bot
}
