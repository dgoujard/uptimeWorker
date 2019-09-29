package services

import (
	"context"
	"github.com/dgoujard/uptimeWorker/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (d *DatabaseService) GetNotificationGroup(id string) *NotificationGroup {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := d.client.Database(d.databaseName).Collection("notificationgroups")
	var notificationGroup NotificationGroup
	objId,_ := primitive.ObjectIDFromHex(id)
	err := collection.FindOne(ctx, bson.D{{"_id", objId}}).Decode(&notificationGroup)
	if err != nil {
		log.Fatal(err)
	}
	return &notificationGroup
}

func (d *DatabaseService) UpdateSiteStatus(bdd *SiteBdd, newStatus int,lastLogTimestamp int64)  {
	bdd.Status = newStatus
	d.client.Database(d.databaseName).Collection("sites").FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": bdd.Id},
		bson.M{"$set": bson.D{
			{"status", bdd.Status},
			{"lastlog", lastLogTimestamp},
		},
		},
	)
}

func (d *DatabaseService) AddLogForSite(site *SiteBdd, sitelog *LogBdd) (error) {
	sitelog.Type = site.Account
	sitelog.Site = site.Id
	res, err := d.client.Database(d.databaseName).Collection("logs").InsertOne(
		context.Background(),
		bson.M{
			"code": sitelog.Code,
			"detail": sitelog.Detail,
			"duration": sitelog.Duration,
			"Type":sitelog.Type,
			"Site":sitelog.Site,
			"datetime":sitelog.Datetime,
		},
	)
	if err != nil {
		log.Fatal(err)
		return err
	}
	sitelog.Id = res.InsertedID.(primitive.ObjectID)
	return  nil
}