package main

import (
	"log"

	"github.com/sunshineplan/metadata"
)

// MetadataConfig is metadata server config
var MetadataConfig = new(metadata.Config)

// GetMongo get mongo config
func GetMongo() (config MongoConfig) {
	c, err := metadata.Get("feh_mongo", MetadataConfig)
	if err != nil {
		log.Fatal(err)
	}
	config.Server = c.(map[string]interface{})["server"].(string)
	config.Port = int(c.(map[string]interface{})["port"].(float64))
	config.Database = c.(map[string]interface{})["database"].(string)
	config.Collection = c.(map[string]interface{})["collection"].(string)
	config.Username = c.(map[string]interface{})["username"].(string)
	config.Password = c.(map[string]interface{})["password"].(string)
	return
}

// GetSubscribe get subscribe info
func GetSubscribe() (config Subscribe) {
	c, err := metadata.Get("feh_subscribe", MetadataConfig)
	if err != nil {
		log.Fatal(err)
	}
	config.Sender = c.(map[string]interface{})["sender"].(string)
	config.Password = c.(map[string]interface{})["password"].(string)
	config.SMTPServer = c.(map[string]interface{})["smtp_server"].(string)
	config.SMTPServerPort = int(c.(map[string]interface{})["smtp_server_port"].(float64))
	config.Subscriber = c.(map[string]interface{})["subscriber"].(string)
	return
}

// GetGithub get github info
func GetGithub() (config Github) {
	c, err := metadata.Get("feh_github", MetadataConfig)
	if err != nil {
		log.Fatal(err)
	}
	config.User = c.(map[string]interface{})["user"].(string)
	config.Repository = c.(map[string]interface{})["repo"].(string)
	config.Token = c.(map[string]interface{})["token"].(string)
	config.Path = c.(map[string]interface{})["path"].(string)
	return
}
