package main

import (
	"log"
	"time"

	"github.com/avast/retry-go"
	"github.com/sunshineplan/metadata"
)

// MetadataConfig is metadata server config
var MetadataConfig = new(metadata.Config)

var (
	// Attempts is default retry attempts
	Attempts = uint(3)
	// Delay is default retry delay
	Delay = 10 * time.Second
	// LastErrorOnly return all errors if false
	LastErrorOnly = true
)

// GetMongo get mongo config
func GetMongo() (config MongoConfig) {
	var c interface{}
	err := retry.Do(
		func() (err error) {
			c, err = metadata.Get("feh_mongo", MetadataConfig)
			return
		},
		retry.Attempts(Attempts),
		retry.Delay(Delay),
		retry.LastErrorOnly(LastErrorOnly),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Failed to get metadata feh_mongo. #%d: %s\n", n+1, err)
		}),
	)
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
	var c interface{}
	err := retry.Do(
		func() (err error) {
			c, err = metadata.Get("feh_subscribe", MetadataConfig)
			return
		},
		retry.Attempts(Attempts),
		retry.Delay(Delay),
		retry.LastErrorOnly(LastErrorOnly),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Failed to get metadata feh_subscribe. #%d: %s\n", n+1, err)
		}),
	)
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
	var c interface{}
	err := retry.Do(
		func() (err error) {
			c, err = metadata.Get("feh_github", MetadataConfig)
			return
		},
		retry.Attempts(Attempts),
		retry.Delay(Delay),
		retry.LastErrorOnly(LastErrorOnly),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Failed to get metadata feh_github. #%d: %s\n", n+1, err)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	config.User = c.(map[string]interface{})["user"].(string)
	config.Repository = c.(map[string]interface{})["repo"].(string)
	config.Token = c.(map[string]interface{})["token"].(string)
	config.Path = c.(map[string]interface{})["path"].(string)
	return
}
