package services

import "go.mongodb.org/mongo-driver/bson/primitive"

type SiteBdd struct{
	Id primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	Account primitive.ObjectID `json:"Account,omitempty" bson:"Account,omitempty"`
	CreateDatetime int32 `json:"createDatetime,omitempty" bson:"createDatetime,omitempty"`
	Lastlog int32 `json:"lastlog,omitempty" bson:"lastlog,omitempty"` //Ne marche pas
	Name string
	Url string
	Status int
	UptimeId int32 `json:"uptimeId,omitempty" bson:"uptimeId,omitempty"`
}