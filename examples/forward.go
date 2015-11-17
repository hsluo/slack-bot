package main

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"
)

const KEY = "fwdtable"

var fwdtable map[string][]string

func handleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	toChan := r.PostFormValue("channel_id")
	text := r.PostFormValue("text")
	splits := strings.SplitN(text, " ", 2)
	fromChan, query := splits[0], splits[1]

	client := urlfetch.Client(ctx)
	channels, err := bot.WithClient(client).ChannelsList()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	log.Debugf(ctx, "%v", channels)

	for i := range channels {
		if fromChan != channels[i].Name {
			continue
		}

		if fwdtable == nil {
			fwdtable = make(map[string][]string)
			_, err := memcache.JSON.Get(ctx, KEY, &fwdtable)
			if err != nil && err != memcache.ErrCacheMiss {
				log.Errorf(ctx, "memcache get error: %v", err)
				return
			}
		}

		if _, ok := fwdtable[channels[i].Id]; !ok {
			fwdtable[channels[i].Id] = []string{toChan, query}
			if err := memcache.JSON.Set(ctx, &memcache.Item{
				Key:    KEY,
				Object: fwdtable,
			}); err != nil {
				log.Errorf(ctx, "memcache set error: %v", err)
			} else {
				fmt.Fprintf(w, "registered chan=%s query=%s", fromChan, query)
			}
		}

		return
	}
}

// forward is a cron job that fetches the latest messages from registered channels
// and forwards them to other channels.
// It should be implemented using RTM API, but it's not possible on GAE.
func forward(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)

	if fwdtable == nil {
		fwdtable = make(map[string][]string)
		_, err := memcache.JSON.Get(ctx, KEY, &fwdtable)
		if err != nil {
			if err != memcache.ErrCacheMiss {
				log.Errorf(ctx, "memcache get error: %v", err)
			}
			return
		}
	}

	for fromChan, v := range fwdtable {
		oldest := time.Now().Add(-15 * time.Second).Unix()
		messages, err := bot.WithClient(client).ChannelsHistory(url.Values{
			"channel": {fromChan},
			"oldest":  {strconv.FormatInt(oldest, 10)},
		})
		if err != nil {
			log.Errorf(ctx, "channels history error: %v", err)
			continue
		}
		for i := range messages {
			if matched, err := regexp.MatchString(v[1], messages[i].Text); matched {
				bot.WithClient(client).ChatPostMessage(url.Values{
					"channel": {v[0]},
					"text":    {messages[i].Text},
				})
				log.Debugf(ctx, "%s", messages[i].Text)
			} else if err != nil {
				log.Errorf(ctx, "regexp error: %v", err)
			}
		}
	}
}
