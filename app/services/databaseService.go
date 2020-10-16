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
	"sync"
	"time"
)

type DatabaseService struct {
	Client *mongo.Client
	DatabaseName string
	LogtypesMap map[int]primitive.ObjectID
	LogtypesMapMux sync.Mutex
}

func (d *DatabaseService) getLogtypesAvailable()  {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := d.Client.Database(d.DatabaseName).Collection("logtypes")
	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result LogTypesBdd
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		d.LogtypesMap[result.TypeId] = result.Id
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
}
func CreateDatabaseConnection(config *config.DatabaseConfig) *DatabaseService {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+config.User+":"+config.Password+"@"+config.Server+":"+strconv.Itoa(config.Port)+"/"+config.Database))
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	return &DatabaseService{
		Client:client,
		DatabaseName: config.Database,
		LogtypesMap:make(map[int]primitive.ObjectID),
	}
}

func (d *DatabaseService) GetSitesList(withPaused bool) (sites []SiteBdd)  {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := d.Client.Database(d.DatabaseName).Collection("sites")
	var filter = bson.M{}
	if !withPaused {
		filter = bson.M{"status": bson.M{"$ne": SiteStatusPause}}
	}
	cur, err := collection.Find(ctx, filter)
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
		//if(result.Name == "Afficheurs CCI (S6)"){
			sites = append(sites, result)
		//}
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	return sites
}

func (d *DatabaseService) GetLastSiteLog(id primitive.ObjectID, lastDownLog bool) (logSite *LogBdd) {
	d.LogtypesMapMux.Lock()
	if len(d.LogtypesMap) == 0 { //Recuperation des types si je n'en ai pas déjà eu besoin avant
		d.getLogtypesAvailable()
	}
	d.LogtypesMapMux.Unlock()

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := d.Client.Database(d.DatabaseName).Collection("logs")

	var logType primitive.ObjectID
	if lastDownLog {
		logType = d.LogtypesMap[LogTypeStatusDown]
	}else{
		logType = d.LogtypesMap[LogTypeStatusUp]
	}
	err := collection.FindOne(ctx,
		bson.M{
		"Site": id,
		"Type":logType,
		}, &options.FindOneOptions{Sort:bson.D{{"datetime",-1}}}).Decode(&logSite)
	if err != nil {
		log.Fatal(err)
	}
	return logSite
}

func (d *DatabaseService) GetNotificationGroup(id string) *NotificationGroup {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := d.Client.Database(d.DatabaseName).Collection("notificationgroups")
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
	d.Client.Database(d.DatabaseName).Collection("sites").FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": bdd.Id},
		bson.M{"$set": bson.D{
			{"status", bdd.Status},
			{"lastlog", lastLogTimestamp},
		},
		},
	)
}

func (d *DatabaseService) AddLogForSite(site *SiteBdd, sitelog *LogBdd, isDown bool) error {
	d.LogtypesMapMux.Lock()
	if len(d.LogtypesMap) == 0 { //Recuperation des types si je n'en ai pas déjà eu besoin avant
		d.getLogtypesAvailable()
	}
	d.LogtypesMapMux.Unlock()
	if isDown {
		sitelog.Type = d.LogtypesMap[LogTypeStatusDown]
	}else{
		sitelog.Type = d.LogtypesMap[LogTypeStatusUp]
	}
	sitelog.Site = site.Id

	res, err := d.Client.Database(d.DatabaseName).Collection("logs").InsertOne(
		context.Background(),
		bson.M{
			"code": sitelog.Code,
			"detail": sitelog.Detail,
			"duration": sitelog.Duration,
			"Type":sitelog.Type,
			"Site":sitelog.Site,
			"datetime":sitelog.Datetime,
			"takeIntoAccount": sitelog.TakeIntoAccount,
		},
	)
	if err != nil {
		log.Fatal(err)
		return err
	}
	sitelog.Id = res.InsertedID.(primitive.ObjectID)
	return  nil
}