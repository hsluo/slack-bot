package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"
)

const KEY = "fwdtable"

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
		fwdtable := make(map[string][]string)
		if _, err := memcache.JSON.Get(ctx, KEY, &fwdtable); err == memcache.ErrCacheMiss {
			if _, ok := fwdtable[channels[i].Id]; !ok {
				fwdtable[channels[i].Id] = []string{toChan, query}
				if err := memcache.JSON.Set(ctx, &memcache.Item{
					Key:    KEY,
					Object: fwdtable,
				}); err != nil {
					log.Errorf(ctx, "memcache set error: %v", err)
				} else {
					fmt.Fprintln(w, "added")
				}
			}
		} else if err != nil {
			log.Errorf(ctx, "memcache get error: %v", err)
		} else {
			fwdtable[channels[i].Id] = []string{toChan, query}
			if err := memcache.JSON.Set(ctx, &memcache.Item{
				Key:    KEY,
				Object: fwdtable,
			}); err != nil {
				log.Errorf(ctx, "memcache set error: %v", err)
			} else {
				fmt.Fprintln(w, "updated")
			}
		}
		return
	}
}

func forward(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	//client := urlfetch.Client(ctx)

	fwdtable := make(map[string][]string)
	_, err := memcache.JSON.Get(ctx, KEY, &fwdtable)
	if err == memcache.ErrCacheMiss {
		return
	} else if err != nil {
		log.Errorf(ctx, "memcache get error: %v", err)
	}

	for fromChan, v := range fwdtable {
		oldest := time.Now().Add(-10 * time.Second).Unix()
		messages, err := bot.ChannelsHistory(url.Values{
			"channel": {fromChan},
			"oldest":  {strconv.FormatInt(oldest, 10)},
		})
	}
}
