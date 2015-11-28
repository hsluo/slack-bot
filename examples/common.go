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

type StringSet struct {
	set map[string]struct{}
}

func newStringSet() StringSet {
	return StringSet{set: make(map[string]struct{})}
}

func (v *StringSet) add(key string) {
	v.set[key] = struct{}{}
}

func (v StringSet) contains(key string) bool {
	_, ok := v.set[key]
	return ok
}

func (v StringSet) toSlice() []string {
	s := make([]string, 0, len(v.set))
	for k, _ := range v.set {
		s = append(s, k)
	}
	return s
}

func sendCommitMessage(m slack.Message, outgoing chan<- slack.Message) {
	m.Text = WhatTheCommit(nil)
	outgoing <- m
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

func isImage(m slack.Message) bool {
	return m.SubType == "file_share" &&
		strings.HasPrefix(m.File.Mimetype, "image")
}

func init() {
	log.SetFlags(log.Lshortfile)
	log.Println("common init")
	rand.Seed(time.Now().Unix())
	var err error
	credentials, err = slack.LoadCredentials("credentials.json")
	if err != nil {
		log.Println(err)
	}
	bot = credentials.Bot
}

func main() {}
