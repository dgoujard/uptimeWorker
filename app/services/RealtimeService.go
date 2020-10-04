package services

import (
	"context"
	"encoding/json"
	"github.com/centrifugal/gocent"
	"github.com/dgoujard/uptimeWorker/config"
)

type RealtimeService struct {
	Client *gocent.Client
}

func CreateRealtimeClient(config *config.RealtimeConfig) *RealtimeService {
	client := gocent.New(gocent.Config{
		Addr: config.Apiurl,
		Key:  config.Apikey,
	})
	return &RealtimeService{Client:client }
}

func (r *RealtimeService) Publish(channel string, message map[string]string) error {
	dataBytes, _ := json.Marshal(message)
	err := r.Client.Publish(context.Background(),channel,dataBytes)
	if err != nil {
		return err
	}
	return nil
}