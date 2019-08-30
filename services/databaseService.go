package services

import (
	"context"
	"github.com/dgoujard/uptimeWorker/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"strconv"
	"time"
)

type DatabaseService struct {
	client *mongo.Client
	databaseName string
}

func CreateDatabaseConnection(config *config.DatabaseConfig) *DatabaseService {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+config.User+":"+config.Password+"@"+config.Server+":"+strconv.Itoa(config.Port)+"/"+config.Database))

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	return &DatabaseService{
		client:client,
		databaseName: config.Database,
	}
}

func (d *DatabaseService) GetSitesLis() (sites []SiteBdd)  {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := d.client.Database(d.databaseName).Collection("sites")
	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result SiteBdd
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		sites = append(sites, result)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	return sites
}