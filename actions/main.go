package main

import (
	"flag"
	"log"
	"time"

	"github.com/sunshineplan/utils"
	"github.com/sunshineplan/utils/database/mongodb"
	"github.com/sunshineplan/utils/mail"
)

var db = mongodb.Config{
	Database:   "feh",
	Collection: "feh",
	Username:   "feh",
	Password:   "feh",
	SRV:        true,
}

var timezone *time.Location
var dialer mail.Dialer
var to string

func main() {
	tz := flag.String("tz", "Local", "Time Zone")
	flag.StringVar(&db.Server, "mongo", "", "MongoDB Server")
	flag.StringVar(&dialer.Host, "server", "smtp.live.com", "SMTP Server")
	flag.IntVar(&dialer.Port, "port", 587, "SMTP Server Port")
	flag.StringVar(&dialer.Account, "account", "", "Mail Account")
	flag.StringVar(&dialer.Password, "password", "", "Mail Account Password")
	flag.StringVar(&to, "to", "", "Backup Account")
	flag.Parse()

	var err error
	timezone, err = time.LoadLocation(*tz)
	if err != nil {
		log.Fatal(err)
	}

	switch flag.Arg(0) {
	case "update":
		if err := update(); err != nil {
			if err == utils.ErrNoMoreRetry {
				log.Print("Event not open.")
			} else {
				log.Fatal(err)
			}
		}
	case "backup":
		backup()
	case "commit":
		commit()
	default:
		log.Fatalln("Unknown argument:", flag.Arg(0))
	}
}
