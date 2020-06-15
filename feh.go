package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/anaskhan96/soup"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Round string
var Round = map[int]string{1: "Round1", 2: "Round2", 3: "Final Round"}

// Scoreboard contains battle score
type Scoreboard struct {
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

// Scrape Fire Emblem Heroes Voting Gauntlet Scoreboard
func Scrape() (event int, round int, fullScoreboard []Scoreboard, status int) {
	resp, err := soup.Get("https://support.fire-emblem-heroes.com/voting_gauntlet/current")
	if err != nil {
		log.Fatal(err)
	}
	doc := soup.HTMLParse(resp)
	for _, class := range strings.Split(doc.Find("div", "class", "title-section").Attrs()["class"], " ") {
		if strings.Contains(class, "cover") {
			event, _ = strconv.Atoi(strings.Split(class, "-")[1])
			break
		}
	}
	for _, class := range strings.Split(doc.Find("h2", "class", "title-section").Attrs()["class"], " ") {
		if strings.Contains(class, "tournament") {
			round, _ = strconv.Atoi(strings.Split(class, "-")[2])
			break
		}
	}
	allBattles := doc.FindAll("li", "class", "tournaments-battle")
	for _, battle := range allBattles {
		scoreboard := new(Scoreboard)
		content := battle.FindAll("p")
		scoreboard.Hero1 = content[0].Text()
		scoreboard.Score1, _ = strconv.Atoi(strings.Replace(content[1].Text(), ",", "", -1))
		scoreboard.Hero2 = content[2].Text()
		scoreboard.Score2, _ = strconv.Atoi(strings.Replace(content[3].Text(), ",", "", -1))
		fullScoreboard = append(fullScoreboard, *scoreboard)
	}
	if len(doc.FindAll("div", "class", "tournaments-art-win")) == 0 {
		status = 1
	} else {
		status = 0 // Event Not Open
	}
	return
}
