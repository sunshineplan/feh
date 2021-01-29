package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/vharitonsky/iniflags"
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
	iniflags.SetConfigFile(filepath.Join(filepath.Dir(self), "config.ini"))
	iniflags.SetAllowMissingConfigFile(true)
	iniflags.Parse()

	initMongo()

	switch flag.NArg() {
	case 0:
		update()
	case 1:
		switch flag.Arg(0) {
		case "update":
			update()
		case "backup":
			backup()
		case "upload":
			upload(0)
		default:
			log.Fatalf("Unknown argument: %s", flag.Arg(0))
		}
	case 2:
		if flag.Arg(0) != "upload" {
			log.Fatalf("Unknown arguments: %s", strings.Join(flag.Args(), " "))
		} else {
			event, err := strconv.Atoi(flag.Arg(1))
			if err != nil {
				log.Fatalf("Unknown argument: upload %s", flag.Arg(1))
			}
			upload(event)
		}
	default:
		log.Fatalf("Unknown arguments: %s", strings.Join(flag.Args(), " "))
	}
}
