package feh

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/sunshineplan/node"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ErrEventNotOpen is an error when event is not open.
var ErrEventNotOpen = errors.New("event not open yet")

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
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	resp, err := http.Get("https://support.fire-emblem-heroes.com/voting_gauntlet/current")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	s := strings.Split(resp.Request.URL.String(), "/")
	event, err = strconv.Atoi(s[len(s)-1])
	if err != nil {
		return
	}

	doc, err := node.Parse(resp.Body)
	if err != nil {
		return
	}
	re := regexp.MustCompile(`title-tournament-(\d+)`)
	h2 := doc.Find(0, node.Tag("h2"), node.Class(re))
	class, _ := h2.Attrs().Get("class")
	if res := re.FindStringSubmatch(class); len(res) == 0 {
		err = errors.New("round not found")
		return
	} else {
		round, err = strconv.Atoi(res[1])
		if err != nil {
			return
		}
	}

	allBattles := doc.FindAll(0, node.Li, node.Class("tournaments-battle"))
	for _, battle := range allBattles {
		scoreboard := Scoreboard{Event: event, Round: round}
		left := battle.Find(0, node.Div, node.Class(regexp.MustCompile("left"))).Find(0, node.P, node.Class("name"))
		scoreboard.Hero1 = left.GetText()
		scoreboard.Score1, err = strconv.Atoi(strings.ReplaceAll(left.NextSibling().GetText(), ",", ""))
		if err != nil {
			err = ErrEventNotOpen
			return
		}
		right := battle.Find(0, node.Div, node.Class(regexp.MustCompile("right"))).Find(0, node.P, node.Class("name"))
		scoreboard.Hero2 = right.GetText()
		scoreboard.Score2, err = strconv.Atoi(strings.ReplaceAll(right.NextSibling().GetText(), ",", ""))
		if err != nil {
			err = ErrEventNotOpen
			return
		}
		fullScoreboard = append(fullScoreboard, scoreboard)
	}

	return
}
