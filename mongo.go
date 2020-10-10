package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"time"

	"github.com/sunshineplan/utils/retry"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type mongoConfig struct {
	Server     string
	Port       int
	Database   string
	Collection string
	Username   string
	Password   string
}

func connect() (*mongo.Client, mongoConfig) {
	var client *mongo.Client
	c := getMongo()
	if err := retry.Do(
		func() (err error) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			client, err = mongo.Connect(ctx, options.Client().ApplyURI(
				fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", c.Username, c.Password, c.Server, c.Port, c.Database)))
			if err != nil {
				return
			}
			ctx, cancelPing := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancelPing()
			err = client.Ping(ctx, readpref.Primary())
			return
		}, 3, 10); err != nil {
		log.Fatal(err)
	}
	return client, c
}

func record(fullScoreboard []Scoreboard) []Scoreboard {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, config := connect()
	defer client.Disconnect(ctx)
	collection := client.Database(config.Database).Collection(config.Collection)
	var newScoreboard []Scoreboard
	for _, scoreboard := range fullScoreboard {
		var result bson.M
		if err := collection.FindOne(
			ctx,
			bson.M{"event": scoreboard.Event,
				"scoreboard.hero": bson.M{"$all": bson.A{scoreboard.Hero1, scoreboard.Hero2}}},
			options.FindOne().SetProjection(bson.M{"_id": 0, "round": 1}),
		).Decode(&result); err == nil {
			scoreboard.Round = int(result["round"].(int32))
		} else if err != mongo.ErrNoDocuments {
			log.Fatal(err)
		}

		r, err := collection.UpdateOne(
			ctx,
			bson.M{
				"scoreboard": bson.A{
					bson.D{
						bson.E{Key: "hero", Value: scoreboard.Hero1},
						bson.E{Key: "score", Value: scoreboard.Score1}},
					bson.D{
						bson.E{Key: "hero", Value: scoreboard.Hero2},
						bson.E{Key: "score", Value: scoreboard.Score2}}}},
			bson.M{
				"$setOnInsert": bson.D{
					bson.E{Key: "event", Value: scoreboard.Event},
					bson.E{Key: "date", Value: time.Now().Truncate(24 * time.Hour)},
					bson.E{Key: "hour", Value: time.Now().Hour()},
					bson.E{Key: "round", Value: scoreboard.Round}}},
			options.Update().SetUpsert(true),
		)
		if err != nil {
			log.Fatal(err)
		}
		if r.UpsertedCount == 1 {
			newScoreboard = append(newScoreboard, scoreboard)
		}
	}
	if len(newScoreboard) > 0 {
		return newScoreboard
	}
	return nil
}

func converter(d []bson.D) string {
	var output string
	for index, item := range d {
		for i, e := range item {
			if e.Key == "date" {
				item[i].Value = e.Value.(primitive.DateTime).Time().Format("2006-01-02")
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

func result(event int) (string, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, config := connect()
	defer client.Disconnect(ctx)
	collection := client.Database(config.Database).Collection(config.Collection)
	var detail, summary []bson.D

	opts := options.Find()
	opts.SetProjection(bson.M{"_id": 0, "event": 0})
	opts.SetSort(bson.D{
		bson.E{Key: "round", Value: 1},
		bson.E{Key: "scoreboard", Value: 1},
		bson.E{Key: "date", Value: 1}})
	detailCur, err := collection.Find(ctx, bson.M{"event": event}, opts)
	if err != nil {
		log.Fatal(err)
	}
	defer detailCur.Close(ctx)
	if err = detailCur.All(ctx, &detail); err != nil {
		log.Fatal(err)
	}
	if len(detail) == 0 {
		return "", ""
	}

	var pipeline []interface{}
	pipeline = append(pipeline, bson.M{"$match": bson.M{"event": event}})
	pipeline = append(pipeline, bson.M{"$addFields": bson.M{"tmp": "$scoreboard"}})
	pipeline = append(pipeline, bson.M{"$unwind": "$tmp"})
	pipeline = append(pipeline, bson.M{
		"$group": bson.D{
			bson.E{Key: "_id", Value: bson.D{
				bson.E{Key: "r", Value: "$round"},
				bson.E{Key: "h", Value: "$tmp.hero"}}},
			bson.E{Key: "s", Value: bson.M{"$max": "$scoreboard"}},
			bson.E{Key: "d", Value: bson.M{"$max": "$date"}}}})
	pipeline = append(pipeline, bson.M{
		"$group": bson.M{
			"_id": bson.D{
				bson.E{Key: "d", Value: "$d"},
				bson.E{Key: "r", Value: "$_id.r"},
				bson.E{Key: "s", Value: "$s"}}}})
	pipeline = append(pipeline, bson.M{
		"$project": bson.D{
			bson.E{Key: "_id", Value: 0},
			bson.E{Key: "date", Value: "$_id.d"},
			bson.E{Key: "round", Value: "$_id.r"},
			bson.E{Key: "scoreboard", Value: "$_id.s"}}})
	pipeline = append(pipeline, bson.M{
		"$sort": bson.D{
			bson.E{Key: "round", Value: 1},
			bson.E{Key: "scoreboard", Value: 1}}})
	summaryCur, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}
	defer summaryCur.Close(ctx)
	if err = summaryCur.All(ctx, &summary); err != nil {
		log.Fatal(err)
	}
	return converter(detail), converter(summary)
}

func dump() string {
	tmpfile, err := ioutil.TempFile("", "tmp")
	if err != nil {
		log.Fatal(err)
	}
	tmpfile.Close()

	mongoConfig := getMongo()
	args := []string{}
	args = append(args, fmt.Sprintf("-h%s:%d", mongoConfig.Server, mongoConfig.Port))
	args = append(args, fmt.Sprintf("-d%s", mongoConfig.Database))
	args = append(args, fmt.Sprintf("-c%s", mongoConfig.Collection))
	args = append(args, fmt.Sprintf("-u%s", mongoConfig.Username))
	args = append(args, fmt.Sprintf("-p%s", mongoConfig.Password))
	args = append(args, "--gzip")
	args = append(args, fmt.Sprintf("--archive=%s", tmpfile.Name()))
	cmd := exec.Command("mongodump", args...)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	return tmpfile.Name()
}
