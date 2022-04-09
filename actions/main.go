package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	feh "feh/utils"

	"github.com/sunshineplan/database/mongodb/driver"
	"github.com/sunshineplan/utils/mail"
	"github.com/sunshineplan/utils/retry"
)

var db = driver.Client{
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
	flag.StringVar(&dialer.Server, "server", "", "SMTP Server")
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

	if err := db.Connect(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	switch flag.Arg(0) {
	case "update":
		err = feh.Update(&dialer, []string{to}, timezone, &db)
		if retry.IsNoMoreRetry(err) {
			log.Print(err)
			return
		}
	case "backup":
		err = feh.Backup(&dialer, []string{to}, timezone, &db)
	case "commit":
		err = commit()
	default:
		log.Fatalln("Unknown argument:", flag.Arg(0))
	}
	if err != nil {
		log.Fatal(err)
	}
}

func commit() error {
	event, detail, summary, err := feh.Result(0, timezone, &db)
	if err != nil {
		return err
	}

	err = os.WriteFile("message", []byte(fmt.Sprintf("FEH 投票大戦第%d回", event)), 0666)
	if err != nil {
		return err
	}

	c := make(chan error)
	go func() {
		c <- os.WriteFile(fmt.Sprintf("FEH 投票大戦第%d回.json", event), []byte(detail), 0666)
	}()

	err = os.WriteFile(fmt.Sprintf("FEH 投票大戦第%d回結果一覧.json", event), []byte(summary), 0666)
	if err != nil {
		return err
	}

	return <-c
}
