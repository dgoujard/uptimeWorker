package services

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SiteBdd struct{
	Id primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	Account primitive.ObjectID `json:"Account,omitempty" bson:"Account,omitempty"`
	NotificationGroup primitive.ObjectID `json:"NotificationGroup,omitempty" bson:"NotificationGroup,omitempty"`
	CreateDatetime int32 `json:"createDatetime,omitempty" bson:"createDatetime,omitempty"`
	Lastlog int32 `json:"lastlog,omitempty" bson:"lastlog,omitempty"` //Ne marche pas
	Name string
	Url string
	Status int
	UptimeId int32 `json:"uptimeId,omitempty" bson:"uptimeId,omitempty"`
}

type LogBdd struct{
	Id primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	Datetime int64 `json:"datetime,omitempty" bson:"datetime,omitempty"`
	Site primitive.ObjectID `json:"Site,omitempty" bson:"Site,omitempty"`
	Type primitive.ObjectID `json:"Type,omitempty" bson:"Type,omitempty"`
	Duration int `json:"duration,omitempty" bson:"duration,omitempty"`
	Code int `json:"code,omitempty" bson:"code,omitempty"` //Ne marche pas
	Detail string
}

type NotificationGroup struct{
	Id primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	Name string
	Cibles []NotificationCible `json:"cibles,omitempty" bson:"cibles,omitempty"`
}

type NotificationCible struct {
	Type string
	Cible string
}