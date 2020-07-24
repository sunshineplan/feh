package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/sunshineplan/metadata"
	"github.com/sunshineplan/utils/mail"
)

var metadataConfig metadata.Config

var (
	attempts      = uint(3)
	delay         = 10 * time.Second
	lastErrorOnly = true
)

func getMongo() (config mongoConfig) {
	m, err := metadata.Get("feh_mongo", &metadataConfig)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(m, &config); err != nil {
		log.Fatal(err)
	}
	return
}

func getSubscribe() (config mail.Setting) {
	m, err := metadata.Get("feh_subscribe", &metadataConfig)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(m, &config); err != nil {
		log.Fatal(err)
	}
	return
}

func getGithub() (config Github) {
	m, err := metadata.Get("feh_github", &metadataConfig)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(m, &config); err != nil {
		log.Fatal(err)
	}
	return
}
