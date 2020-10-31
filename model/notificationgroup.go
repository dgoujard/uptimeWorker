package model

import "gorm.io/gorm"

type NotificationGroup struct {
	gorm.Model
	Name         string
	Sites   []Site
}