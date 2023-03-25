package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	feh "feh/utils"

	"github.com/sunshineplan/utils/flags"
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	fmt.Println(`
  update
        Update Current Fire Emblem Heroes Voting Gauntlet Event Scoreboard to Database.

  backup
        Backup Fire Emblem Heroes Voting Gauntlet Scoreboard Database.

  upload num int(default 0)
        Upload Fire Emblem Heroes Voting Gauntlet Event num's Results to Github Repository.
        Value 0 will treat as current event`)
}

func main() {
	self, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	flag.Usage = usage
	flag.StringVar(&meta.Addr, "server", "", "Metadata Server Address")
	flag.StringVar(&meta.Header, "header", "", "Verify Header Header Name")
	flag.StringVar(&meta.Value, "value", "", "Verify Header Value")
	flags.SetConfigFile(filepath.Join(filepath.Dir(self), "config.ini"))
	flags.Parse()

	initMongo()

	switch flag.NArg() {
	case 0, 1:
		switch flag.Arg(0) {
		case "", "update":
			dialer, to := getSubscribe()
			err = feh.Update(dialer, to, time.Local, &db)
			if err == nil {
				log.Print("Update FEH done.")
			}
		case "backup":
			dialer, to := getSubscribe()
			err = feh.Backup(dialer, to, time.Local, &db)
			if err == nil {
				log.Print("Backup FEH done.")
			}
		case "upload":
			err = upload(0)
		default:
			log.Fatalf("Unknown argument: %s", flag.Arg(0))
		}
	case 2:
		if flag.Arg(0) != "upload" {
			log.Fatalf("Unknown arguments: %s", strings.Join(flag.Args(), " "))
		} else {
			var event int
			event, err = strconv.Atoi(flag.Arg(1))
			if err != nil {
				log.Fatalf("Unknown argument: upload %s", flag.Arg(1))
			}
			err = upload(event)
		}
	default:
		log.Fatalf("Unknown arguments: %s", strings.Join(flag.Args(), " "))
	}
	if err != nil {
		log.Fatal(err)
	}
}
