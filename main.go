package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/vharitonsky/iniflags"
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	fmt.Println(
		`  update
        Update Current Fire Emblem Heroes Voting Gauntlet Event Scoreboard to Database.

  backup
        Backup Fire Emblem Heroes Voting Gauntlet Scoreboard Database.

  upload num int(default 0)
        Upload Fire Emblem Heroes Voting Gauntlet Event num's Results to Github Repository.
        Value 0 will treat as current event`)
}

func main() {
	flag.Usage = usage
	flag.StringVar(&MetadataConfig.Server, "server", "", "Metadata Server Address")
	flag.StringVar(&MetadataConfig.VerifyHeader, "header", "", "Verify Header Header Name")
	flag.StringVar(&MetadataConfig.VerifyValue, "value", "", "Verify Header Value")
	iniflags.SetConfigFile("config.ini")
	iniflags.SetAllowMissingConfigFile(true)
	iniflags.Parse()

	switch flag.NArg() {
	case 0:
		Update()
	case 1:
		switch flag.Arg(0) {
		case "update":
			Update()
		case "backup":
			Backup()
		case "upload":
			Upload(0)
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
			Upload(event)
		}
	default:
		log.Fatalf("Unknown arguments: %s", strings.Join(flag.Args(), " "))
	}
}
