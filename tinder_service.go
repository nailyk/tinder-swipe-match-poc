package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

//go:embed swipe_and_check_match.lua
var swipeAndCheckMatchFunction string

type TinderService struct {
	ctx    context.Context
	client *redis.Client
}

type SwipeAction string

const (
	SwipeLike    SwipeAction = "like"
	SwipeDislike SwipeAction = "dislike"
)

func NewTinderService(redisClient *redis.Client) *TinderService {
	return &TinderService{
		ctx:    context.Background(),
		client: redisClient,
	}
}

func (ts *TinderService) LoadLuaFunction() error {
	if err := ts.client.Do(ts.ctx, "FUNCTION", "LOAD", "REPLACE", swipeAndCheckMatchFunction).Err(); err != nil {
		return fmt.Errorf("failed to load Lua function: %w", err)
	}
	log.Println("📦 Redis Lua function loaded successfully.")
	return nil
}

func (ts *TinderService) SwipeAndCheckMatch(fromUser, toUser string, swipeAction SwipeAction) (bool, error) {
	key := normalizeSwipeKey(fromUser, toUser)

	cmd := ts.client.FCall(ts.ctx, "swipeAndCheckMatch", []string{key}, fromUser, toUser, string(swipeAction))
	if err := cmd.Err(); err != nil {
		return false, fmt.Errorf("redis FCall error: %w", err)
	}

	val, err := cmd.Int()
	if err != nil {
		return false, fmt.Errorf("invalid response from Lua script: %w", err)
	}

	return val == 1, nil
}

func (ts *TinderService) GetMatches(user string) ([]string, error) {
	matches, err := ts.client.SMembers(ts.ctx, "matches:"+user).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get matches for user %s: %w", user, err)
	}
	return matches, nil
}

func normalizeSwipeKey(user1, user2 string) string {
	if user1 < user2 {
		return "swipes:" + user1 + ":" + user2
	}
	return "swipes:" + user2 + ":" + user1
}
