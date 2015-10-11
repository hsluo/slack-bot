package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

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
