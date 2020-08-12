package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/anaskhan96/soup"
	"github.com/avast/retry-go"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Round string
var Round = map[int]string{1: "Round1", 2: "Round2", 3: "Final Round"}

// Scoreboard contains battle score
type Scoreboard struct {
	Event  int
	Round  int
	Hero1  string
	Score1 int
	Hero2  string
	Score2 int
}

func rightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}

func thousandsComma(i int) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", i)
}

// Formatter Scoreboard output string
func (s *Scoreboard) Formatter() string {
	return fmt.Sprintf(
		"%s%15s    VS    %s%15s",
		rightPad2Len(s.Hero1, "　", 24), // 8*3
		thousandsComma(s.Score1),
		rightPad2Len(s.Hero2, "　", 24),
		thousandsComma(s.Score2))
}

func scrape() (event int, round int, fullScoreboard []Scoreboard) {
	var body string
	err := retry.Do(
		func() (err error) {
			body, err = soup.Get("https://support.fire-emblem-heroes.com/voting_gauntlet/current")
			return
		},
		retry.Attempts(attempts),
		retry.Delay(delay),
		retry.LastErrorOnly(lastErrorOnly),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Scoreboard scrape failed. #%d: %s\n", n+1, err)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	doc := soup.HTMLParse(body)
	for _, class := range strings.Split(doc.Find("div", "class", "title-section").Attrs()["class"], " ") {
		if strings.Contains(class, "cover") {
			event, err = strconv.Atoi(strings.Split(class, "-")[1])
			if err != nil {
				log.Fatal(err)
			}
			break
		}
	}
	for _, class := range strings.Split(doc.Find("h2", "class", "title-section").Attrs()["class"], " ") {
		if strings.Contains(class, "tournament") {
			round, err = strconv.Atoi(strings.Split(class, "-")[2])
			if err != nil {
				log.Fatal(err)
			}
			break
		}
	}
	allBattles := doc.FindAll("li", "class", "tournaments-battle")
	for _, battle := range allBattles {
		var scoreboard Scoreboard
		scoreboard.Event = event
		scoreboard.Round = round
		content := battle.FindAll("p")
		scoreboard.Hero1 = content[0].Text()
		scoreboard.Score1, err = strconv.Atoi(strings.Replace(content[1].Text(), ",", "", -1))
		if err != nil {
			log.Fatal(err)
		}
		scoreboard.Hero2 = content[2].Text()
		scoreboard.Score2, err = strconv.Atoi(strings.Replace(content[3].Text(), ",", "", -1))
		if err != nil {
			log.Fatal(err)
		}
		fullScoreboard = append(fullScoreboard, scoreboard)
	}
	return
}
