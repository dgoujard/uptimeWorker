package services

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const SiteStatusUp = 2
const SiteStatusDown = 9
const SiteStatusPause = 0

const LogTypeStatusUp = 2
const LogTypeStatusDown = 1
const LogTypeStatusPause = 99

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
	SslMonitored bool `json:"ssl_monitored,omitempty" bson:"ssl_monitored,omitempty"`
	SslAlertExpireSended bool `json:"ssl_alertExpireSended,omitempty" bson:"ssl_alertExpireSended,omitempty"`
	SslError string `json:"ssl_error,omitempty" bson:"ssl_error,omitempty"`
	SslSubject string `json:"ssl_subject,omitempty" bson:"ssl_subject,omitempty"`
	SslIssuer string `json:"ssl_issuer,omitempty" bson:"ssl_issuer,omitempty"`
	SslAlgo string `json:"ssl_algo,omitempty" bson:"ssl_algo,omitempty"`
	SslExpireDatetime int32 `json:"ssl_expireDatetime,omitempty" bson:"ssl_expireDatetime,omitempty"`
}

type LogBdd struct{
	Id primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	Datetime int64 `json:"datetime,omitempty" bson:"datetime,omitempty"`
	Site primitive.ObjectID `json:"Site,omitempty" bson:"Site,omitempty"`
	Type primitive.ObjectID `json:"Type,omitempty" bson:"Type,omitempty"`
	Duration int `json:"duration,omitempty" bson:"duration,omitempty"`
	Code int `json:"code,omitempty" bson:"code,omitempty"` //Ne marche pas
	Detail string
	TakeIntoAccount  bool `json:"takeIntoAccount,omitempty" bson:"takeIntoAccount,omitempty"`
	Comment string `json:"comment,omitempty" bson:"comment,omitempty"`
}
type LogTypesBdd struct{
	Id primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	TypeId int `json:"logTypeId,omitempty" bson:"logTypeId,omitempty"`
	Label string `json:"logTypeLabel,omitempty" bson:"logTypeLabel,omitempty"`
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