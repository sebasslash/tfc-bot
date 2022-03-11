package store

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/sebasslash/tfc-bot/models"
)

var DB *Redis

type Redis struct {
	*redis.Client
}

func (r *Redis) init() error {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		return fmt.Errorf("REDIS_HOST env var not set")
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		return fmt.Errorf("REDIS_PORT env var not set")
	}

	pwd := os.Getenv("REDIS_PASSWORD")
	if pwd == "" {
		return fmt.Errorf("REDIS_PASSWORD env var not set")
	}

	r.Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: pwd,
		DB:       0,
	})

	return nil
}

func (r *Redis) CreateConfigurationKey(ctx context.Context, key *models.ConfigurationKey, configurationID string) error {
	k := fmt.Sprintf("%s-%s", key.ChannelID, key.WorkspaceID)
	_, err := r.Client.SetNX(ctx, k, configurationID, 0).Result()
	return err
}

func (r *Redis) ReadConfigurationKey(ctx context.Context, key *models.ConfigurationKey) (string, error) {
	k := fmt.Sprintf("%s-%s", key.ChannelID, key.WorkspaceID)
	ncID, err := r.Client.Get(ctx, k).Result()
	if err != nil {
		return "", err
	}

	return ncID, nil
}

func (r *Redis) AddConfiguration(ctx context.Context, configurationID, channelID string) error {
	_, err := r.Client.SetNX(ctx, configurationID, channelID, 0).Result()
	return err
}

func (r *Redis) GetConfiguration(ctx context.Context, configurationID string) (string, error) {
	channelID, err := r.Client.Get(ctx, configurationID).Result()
	if err != nil {
		return "", err
	}

	return channelID, nil
}

func (r *Redis) RemoveConfiguration(ctx context.Context, configurationID string) error {
	_, err := r.Client.Del(ctx, configurationID).Result()
	return err
}

func (r *Redis) PublishNotification(ctx context.Context, notification *models.Notification) error {
	err := r.Client.Publish(ctx, notification.WorkspaceID, notification).Err()
	return err
}

func (r *Redis) SubNotificationChannel(ctx context.Context, workspaceID string) (<-chan *redis.Message, error) {
	pubsub := r.Subscribe(ctx, workspaceID)

	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, err
	}

	return pubsub.Channel(), nil
}

func Create() {
	DB = &Redis{}
	err := DB.init()
	if err != nil {
		panic(err)
	}
}
