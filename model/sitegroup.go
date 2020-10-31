package model

import "gorm.io/gorm"

type SiteGroup struct {
	gorm.Model
	Name         string
	Sites   []Site
}