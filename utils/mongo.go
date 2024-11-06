package utils

import (
	"encoding/json"
	"feh"
	"fmt"
	"log"
	"time"

	"github.com/sunshineplan/database/mongodb"
)

func record(fullScoreboard []feh.Scoreboard, tz *time.Location, db mongodb.Client) (newScoreboard []feh.Scoreboard, err error) {
	for _, scoreboard := range fullScoreboard {
		var res struct{ Round int }
		if err = db.FindOne(
			mongodb.M{
				"event":           scoreboard.Event,
				"scoreboard.hero": mongodb.M{"$all": []string{scoreboard.Hero1, scoreboard.Hero2}},
			},
			&mongodb.FindOneOpt{Projection: mongodb.M{"_id": 0, "round": 1}},
			&res,
		); err == nil {
			scoreboard.Round = res.Round
		} else if err != mongodb.ErrNoDocuments {
			return
		}

		t := time.Now().In(tz).Truncate(24 * time.Hour)
		var r *mongodb.UpdateResult
		r, err = db.UpdateOne(
			mongodb.M{
				"scoreboard": []struct {
					Hero  string `json:"hero"`
					Score int    `json:"score"`
				}{
					{scoreboard.Hero1, scoreboard.Score1},
					{scoreboard.Hero2, scoreboard.Score2},
				},
			},
			mongodb.M{
				"$setOnInsert": struct {
					Event int          `json:"event"`
					Date  mongodb.Time `json:"date"`
					Hour  int          `json:"hour"`
					Round int          `json:"round"`
				}{
					scoreboard.Event,
					db.Time(t),
					t.Hour(),
					scoreboard.Round,
				},
			},
			&mongodb.UpdateOpt{Upsert: true},
		)
		if err != nil {
			return
		}

		if r.UpsertedCount == 1 {
			newScoreboard = append(newScoreboard, scoreboard)
		}
	}

	return
}

type scoreboard struct {
	Date       time.Time
	Hour       int
	Round      int
	Scoreboard []struct {
		Hero  string `json:"hero"`
		Score int    `json:"score"`
	}
}

func result(event int, tz *time.Location, db mongodb.Client) (string, string, error) {
	var detail, summary []scoreboard
	if err := db.Find(
		mongodb.M{"event": event},
		&mongodb.FindOpt{
			Projection: mongodb.M{"_id": 0, "event": 0},
			Sort: struct {
				Round      int `json:"round"`
				Scoreboard int `json:"scoreboard"`
				Date       int `json:"date"`
			}{1, 1, 1},
		},
		&detail,
	); err != nil {
		return "", "", err
	}

	if len(detail) == 0 {
		return "", "", nil
	}

	if err := db.Aggregate([]mongodb.M{
		{"$match": mongodb.M{"event": event}},
		{"$addFields": mongodb.M{"tmp": "$scoreboard"}},
		{"$unwind": "$tmp"},
		{"$group": mongodb.M{
			"_id": mongodb.M{"r": "$round", "h": "$tmp.hero"},
			"s":   mongodb.M{"$max": "$scoreboard"},
			"d":   mongodb.M{"$max": "$date"},
		}},
		{"$group": mongodb.M{"_id": mongodb.M{"d": "$d", "r": "$_id.r", "s": "$s"}}},
		{"$project": struct {
			ID         int    `json:"_id" bson:"_id"`
			Date       string `json:"date" bson:"date"`
			Round      string `json:"round" bson:"round"`
			Scoreboard string `json:"scoreboard" bson:"scoreboard"`
		}{0, "$_id.d", "$_id.r", "$_id.s"}},
		{"$sort": struct {
			Round      int `json:"round" bson:"round"`
			Scoreboard int `json:"scoreboard" bson:"scoreboard"`
		}{1, 1}},
	}, &summary); err != nil {
		return "", "", err
	}

	return converter(detail, true, tz), converter(summary, false, tz), nil
}

func converter(d []scoreboard, showHour bool, tz *time.Location) string {
	var output string
	for index, item := range d {
		var scoreboard interface{}
		if showHour {
			scoreboard = struct {
				Date       string `json:"date"`
				Hour       int    `json:"hour"`
				Round      int    `json:"round"`
				Scoreboard []struct {
					Hero  string `json:"hero"`
					Score int    `json:"score"`
				} `json:"scoreboard"`
			}{
				item.Date.In(tz).Format("2006-01-02"),
				item.Hour,
				item.Round,
				item.Scoreboard,
			}
		} else {
			scoreboard = struct {
				Date       string `json:"date"`
				Round      int    `json:"round"`
				Scoreboard []struct {
					Hero  string `json:"hero"`
					Score int    `json:"score"`
				} `json:"scoreboard"`
			}{
				item.Date.In(tz).Format("2006-01-02"),
				item.Round,
				item.Scoreboard,
			}
		}

		b, err := json.Marshal(scoreboard)
		if err != nil {
			log.Print(err)
		}
		if index < len(d)-1 {
			output = output + string(b) + ",\n"
		} else {
			output = output + string(b)
		}
	}
	return fmt.Sprintf("[%s]", output)
}
