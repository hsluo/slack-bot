package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/hsluo/slack-bot"
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

var (
	exRe = regexp.MustCompile(`\w+::\w+`)
)

// create slack attachement from Loggly's HTTP alert
func NewAttachment(req *http.Request) (attachment slack.Attachment, err error) {
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
			stackTrace[0] = exRe.ReplaceAllStringFunc(stackTrace[0], func(match string) string {
				return "`" + match + "`"
			})
			if len(stackTrace) > 1 {
				stackTrace[1] = ">>> " + stackTrace[1]
			}
			alert.RecentHits[i] = strings.Join(stackTrace, "\n")
		}
	} else {
		fallback = alert.RecentHits[0]
	}
	fields := []slack.Field{
		{Title: "Description", Value: alert.AlertDescription, Short: false},
		{Title: "Query", Value: alert.Query, Short: true},
		{Title: "Num Hits", Value: strconv.Itoa(alert.NumHits), Short: true},
		{Title: "Recent Hits", Value: strings.Join(alert.RecentHits, "\n"), Short: false},
	}
	attachment = slack.Attachment{
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
