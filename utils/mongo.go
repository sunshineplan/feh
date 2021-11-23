package utils

import (
	"context"
	"feh"
	"fmt"
	"log"
	"time"

	"github.com/sunshineplan/database/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func record(fullScoreboard []feh.Scoreboard, tz *time.Location, db *mongodb.Config) (newScoreboard []feh.Scoreboard, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var client *mongo.Client
	client, err = db.Open()
	if err != nil {
		return
	}
	defer client.Disconnect(ctx)

	collection := client.Database(db.Database).Collection(db.Collection)
	for _, scoreboard := range fullScoreboard {
		var result bson.M
		if err = collection.FindOne(
			ctx,
			bson.M{
				"event":           scoreboard.Event,
				"scoreboard.hero": bson.M{"$all": bson.A{scoreboard.Hero1, scoreboard.Hero2}},
			},
			options.FindOne().SetProjection(bson.M{"_id": 0, "round": 1}),
		).Decode(&result); err == nil {
			scoreboard.Round = int(result["round"].(int32))
		} else if err != mongo.ErrNoDocuments {
			return
		}

		t := time.Now().In(tz)
		var r *mongo.UpdateResult
		r, err = collection.UpdateOne(
			ctx,
			bson.M{
				"scoreboard": bson.A{
					bson.D{
						bson.E{Key: "hero", Value: scoreboard.Hero1},
						bson.E{Key: "score", Value: scoreboard.Score1},
					},
					bson.D{
						bson.E{Key: "hero", Value: scoreboard.Hero2},
						bson.E{Key: "score", Value: scoreboard.Score2},
					},
				},
			},
			bson.M{
				"$setOnInsert": bson.D{
					bson.E{Key: "event", Value: scoreboard.Event},
					bson.E{Key: "date", Value: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, tz)},
					bson.E{Key: "hour", Value: t.Hour()},
					bson.E{Key: "round", Value: scoreboard.Round},
				},
			},
			options.Update().SetUpsert(true),
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

func converter(d []bson.D, tz *time.Location) string {
	var output string
	for index, item := range d {
		for i, e := range item {
			if e.Key == "date" {
				item[i].Value = e.Value.(primitive.DateTime).Time().In(tz).Format("2006-01-02")
				break
			}
		}
		b, err := bson.MarshalExtJSON(item, false, true)
		if err != nil {
			log.Println(err)
		}
		if index < len(d)-1 {
			output = output + string(b) + ",\n"
		} else {
			output = output + string(b)
		}
	}
	return fmt.Sprintf("[%s]", output)
}

func result(event int, tz *time.Location, db *mongodb.Config) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := db.Open()
	if err != nil {
		return "", "", err
	}
	defer client.Disconnect(ctx)

	collection := client.Database(db.Database).Collection(db.Collection)
	var detail, summary []bson.D

	opts := options.Find()
	opts.SetProjection(bson.M{"_id": 0, "event": 0})
	opts.SetSort(bson.D{
		bson.E{Key: "round", Value: 1},
		bson.E{Key: "scoreboard", Value: 1},
		bson.E{Key: "date", Value: 1},
	})
	detailCur, err := collection.Find(ctx, bson.M{"event": event}, opts)
	if err != nil {
		return "", "", err
	}
	defer detailCur.Close(ctx)

	if err := detailCur.All(ctx, &detail); err != nil {
		return "", "", err
	}

	if len(detail) == 0 {
		return "", "", nil
	}

	summaryCur, err := collection.Aggregate(ctx, []bson.M{
		{"$match": bson.M{"event": event}},
		{"$addFields": bson.M{"tmp": "$scoreboard"}},
		{"$unwind": "$tmp"},
		{"$group": bson.D{
			bson.E{
				Key: "_id", Value: bson.D{
					bson.E{Key: "r", Value: "$round"},
					bson.E{Key: "h", Value: "$tmp.hero"},
				},
			},
			bson.E{Key: "s", Value: bson.M{"$max": "$scoreboard"}},
			bson.E{Key: "d", Value: bson.M{"$max": "$date"}},
		}},
		{"$group": bson.M{
			"_id": bson.D{
				bson.E{Key: "d", Value: "$d"},
				bson.E{Key: "r", Value: "$_id.r"},
				bson.E{Key: "s", Value: "$s"},
			},
		}},
		{"$project": bson.D{
			bson.E{Key: "_id", Value: 0},
			bson.E{Key: "date", Value: "$_id.d"},
			bson.E{Key: "round", Value: "$_id.r"},
			bson.E{Key: "scoreboard", Value: "$_id.s"},
		}},
		{"$sort": bson.D{
			bson.E{Key: "round", Value: 1},
			bson.E{Key: "scoreboard", Value: 1},
		}},
	})
	if err != nil {
		return "", "", err
	}
	defer summaryCur.Close(ctx)

	if err := summaryCur.All(ctx, &summary); err != nil {
		return "", "", err
	}

	return converter(detail, tz), converter(summary, tz), nil
}
