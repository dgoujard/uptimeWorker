package model

import (
	"time"

	"gorm.io/gorm"
)

type Sites []*Site

type Site struct {
	gorm.Model
	Name         string
	Url        string
	LastStatusCode      uint8
	LastLogDatetime   *time.Time
	SslMonitored   bool
	SslIssuer   string
	SslSubject   string
	SslAlgo   string
	SslExpireDateTime   *time.Time
	SslError   string
	ScreeshotUrl   string
	ScreeshotDateTime   *time.Time
	ScreeshotError   string
	LighthouseUrl   string
	LighthousePerformance   uint8
	LighthouseBestPractice   uint8
	LighthouseSeo   uint8
	LighthousePwa   uint8
	LighthouseDateTime *time.Time
	LighthouseError   string
	IsMonitored   bool
	SslAlertExpirationSended   bool
	NotificationGroupID   bool
	Logs   []Log
	SiteGroupID uint
}

type SiteForm struct {
	Title         string `json:"title" form:"required,max=255"`
	Author        string `json:"author" form:"required,alpha_space,max=255"`
	PublishedDate string `json:"published_date" form:"required,date"`
	ImageUrl      string `json:"image_url" form:"url"`
	Description   string `json:"description"`
}

func (f *SiteForm) ToModel() (*Site, error) {
	/*pubDate, err := time.Parse("2006-01-02", f.PublishedDate)
	if err != nil {
		return nil, err
	}

	return &Site{
		Title:         f.Title,
		Author:        f.Author,
		PublishedDate: pubDate,
		ImageUrl:      f.ImageUrl,
		Description:   f.Description,
	}, nil*/
	return &Site{}, nil
}