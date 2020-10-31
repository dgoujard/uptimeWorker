package model

import "gorm.io/gorm"

type LogType struct {
	gorm.Model
	Logs   []Log
	Name string
	Code uint8
}