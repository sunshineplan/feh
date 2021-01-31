package feh

import (
	"fmt"
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

// Scrape scrapes fireemblem heroes voting gauntlet informations.
func Scrape() (event int, round int, fullScoreboard []Scoreboard, err error) {
	var body string
	body, err = soup.Get("https://support.fire-emblem-heroes.com/voting_gauntlet/current")
	if err != nil {
		return
	}

	doc := soup.HTMLParse(body)
	for _, class := range strings.Split(doc.Find("div", "class", "title-section").Attrs()["class"], " ") {
		if strings.Contains(class, "cover") {
			event, err = strconv.Atoi(strings.Split(class, "-")[1])
			if err != nil {
				return
			}
			break
		}
	}

	for _, class := range strings.Split(doc.Find("h2", "class", "title-section").Attrs()["class"], " ") {
		if strings.Contains(class, "tournament") {
			round, err = strconv.Atoi(strings.Split(class, "-")[2])
			if err != nil {
				return
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
			return
		}
		scoreboard.Hero2 = content[2].Text()
		scoreboard.Score2, err = strconv.Atoi(strings.Replace(content[3].Text(), ",", "", -1))
		if err != nil {
			return
		}
		fullScoreboard = append(fullScoreboard, scoreboard)
	}

	return
}
