package main

import (
	"log"

	"github.com/sunshineplan/database/mongodb/driver"
	"github.com/sunshineplan/metadata"
	"github.com/sunshineplan/utils/mail"
)

var db driver.Client
var meta metadata.Server

func initMongo() {
	if err := meta.Get("feh_mongo", &db); err != nil {
		log.Fatal(err)
	}
	if err := db.Connect(); err != nil {
		log.Fatal(err)
	}
}

func getSubscribe() (dialer *mail.Dialer, to mail.Receipts) {
	var config struct {
		SMTPServer     string
		SMTPServerPort int
		From, Password string
		To             mail.Receipts
	}
	if err := meta.Get("feh_subscribe", &config); err != nil {
		log.Fatalln("Failed to get feh_subscribe metadata:", err)
	}
	dialer = new(mail.Dialer)
	dialer.Server = config.SMTPServer
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
