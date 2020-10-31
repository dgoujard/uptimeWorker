package model

import (
	"gorm.io/gorm"
	"time"
)

type Logs []*Log

type Log struct {
	gorm.Model
	SiteID uint
	LogTypeID uint
	Code         uint8
	Detail        string
	Duration uint
	Datetime time.Time
	IsIgnored   bool
	IgnoreMessage string
}
