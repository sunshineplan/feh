package main

import (
	"flag"
	"log"

	"github.com/sunshineplan/utils/database/mongodb"
	"github.com/sunshineplan/utils/mail"
)

var db = mongodb.Config{
	Port:       27017,
	Database:   "feh",
	Collection: "feh",
	Username:   "feh",
	Password:   "feh",
}

var dialer mail.Dialer
var to string

func main() {
	flag.StringVar(&db.Server, "mongo", "", "MongoDB Server")
	flag.StringVar(&dialer.Host, "server", "smtp.live.com", "SMTP Server")
	flag.IntVar(&dialer.Port, "port", 587, "SMTP Server Port")
	flag.StringVar(&dialer.Account, "account", "", "Mail Account")
	flag.StringVar(&dialer.Password, "password", "", "Mail Account Password")
	flag.StringVar(&to, "to", "", "Backup Account")
	flag.Parse()

	switch flag.Arg(0) {
	case "update":
		update()
	case "backup":
		backup()
	case "upload":
		upload()
	default:
		log.Fatalf("Unknown argument: %s", flag.Arg(0))
	}
}
