package main

import (
	"log"

	"github.com/sunshineplan/database/mongodb/driver"
	"github.com/sunshineplan/utils/mail"
	"github.com/sunshineplan/utils/metadata"
)

var db driver.Client
var meta metadata.Server

func initMongo() {
	if err := meta.Get("feh_mongo", &db); err != nil {
		log.Fatal(err)
	}
}

func getSubscribe() (dialer *mail.Dialer, to []string) {
	var config struct {
		SMTPServer     string
		SMTPServerPort int
		From, Password string
		To             []string
	}
	if err := meta.Get("feh_subscribe", &config); err != nil {
		log.Fatalln("Failed to get feh_subscribe metadata:", err)
	}
	dialer = new(mail.Dialer)
	dialer.Host = config.SMTPServer
	dialer.Port = config.SMTPServerPort
	dialer.Account = config.From
	dialer.Password = config.Password
	to = config.To
	return
}

func getGithub() (config Github) {
	if err := meta.Get("feh_github", &config); err != nil {
		log.Fatal(err)
	}
	return
}
