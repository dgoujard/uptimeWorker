package services

import (
	"context"
	"encoding/json"
	"github.com/centrifugal/gocent"
	"github.com/dgoujard/uptimeWorker/config"
)

type RealtimeService struct {
	client *gocent.Client
}

func CreateRealtimeClient(config *config.RealtimeConfig) *RealtimeService {
	client := gocent.New(gocent.Config{
		Addr: config.Apiurl,
		Key:  config.Apikey,
	})
	return &RealtimeService{client:client }
}

func (r *RealtimeService) Publish(channel string, message map[string]string) error {
	dataBytes, _ := json.Marshal(message)
	err := r.client.Publish(context.Background(),channel,dataBytes)
	if err != nil {
		return err
	}
	return nil
}