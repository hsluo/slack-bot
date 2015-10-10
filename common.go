package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

// create slack attachement from Loggly's HTTP alert
func NewAttachment(req *http.Request) (attachment Attachment, err error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return
	}
	alert := LogglyAlert{}
	if err = json.Unmarshal(body, &alert); err != nil {
		return
	}
	var fallback string
	if strings.Contains(alert.RecentHits[0], "#012") {
		fallback = strings.SplitN(alert.RecentHits[0], "#012", 2)[0]
		for i, hit := range alert.RecentHits {
			stackTrace := strings.Split(strings.TrimSpace(hit), "#012")
			if fallback == "" {
				fallback = stackTrace[0]
			}
			splits := strings.Split(stackTrace[0], " ")
			for i, split := range splits {
				if strings.Contains(split, "::") {
					splits[i] = "`" + split + "`"
				}
			}
			stackTrace[0] = strings.Join(splits, " ")
			if len(stackTrace) > 1 {
				stackTrace[1] = ">>> " + stackTrace[1]
			}
			alert.RecentHits[i] = strings.Join(stackTrace, "\n")
		}
	} else {
		fallback = alert.RecentHits[0]
	}
	fields := []Field{
		{Title: "Description", Value: alert.AlertDescription, Short: false},
		{Title: "Query", Value: alert.Query, Short: true},
		{Title: "Num Hits", Value: strconv.Itoa(alert.NumHits), Short: true},
		{Title: "Recent Hits", Value: strings.Join(alert.RecentHits, "\n"), Short: false},
	}
	attachment = Attachment{
		Fallback:  fallback,
		Color:     "warning",
		Title:     alert.AlertName,
		TitleLink: alert.SearchLink,
		Text:      fmt.Sprintf("Edit this alert on <%s|Loggly>", alert.EditAlertLink),
		Fields:    fields,
		MrkdwnIn:  []string{"fields"},
	}
	return attachment, nil
}

func WhatTheCommit(client *http.Client) string {
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Get("http://whatthecommit.com/index.txt")
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	return strings.TrimSpace(string(body))
}

func sendCommitMessage(m Message, outgoing chan<- Message) {
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

func sendCode(m Message, outgoing chan<- Message) {
	m.Text = ack()
	outgoing <- m

	time.Sleep(1 * time.Second)

	m.Text = codeWithAt(m.User)
	outgoing <- m
}

func isImage(m Message) bool {
	return m.SubType == "file_share" &&
		strings.HasPrefix(m.File.Mimetype, "image")
}

// at in the middle of the message is not supported
func isAt(text string) bool {
	return strings.HasPrefix(text, atId) || strings.HasSuffix(text, atId) ||
		strings.HasPrefix(text, alias) || strings.HasSuffix(text, alias)
}

func handleMessage(incoming <-chan Message, outgoing chan<- Message) {
	for msg := range incoming {
		if msg.Type != "message" {
			continue
		}
		if strings.Contains(msg.Text, "谢谢") {
			msg.Text = "不客气 :blush:"
			outgoing <- msg
		} else if isAt(msg.Text) {
			fields := strings.Fields(msg.Text)
			if len(fields) == 1 {
				sendCode(msg, outgoing)
			} else {
				var commit bool
				log.Println(fields)
				for _, f := range fields {
					if isAt(f) {
						continue
					}
					if strings.Contains(f, "commit") {
						commit = true
					}
				}
				if commit {
					sendCommitMessage(msg, outgoing)
				} else {
					if rand.Intn(2) > 0 {
						msg.Text = "呵呵"
					} else {
						msg.Text = "嘻嘻"
					}
					outgoing <- msg
				}
			}
		} else if isImage(msg) {
			sendCode(msg, outgoing)
		}
	}
}

func init() {
	log.Println("common init")
	loc, _ = time.LoadLocation("Asia/Shanghai")
	rand.Seed(time.Now().Unix())
}
