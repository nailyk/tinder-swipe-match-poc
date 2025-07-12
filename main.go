package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	redisAddr := flag.String("redis", "localhost:6379", "Redis server address")
	numUsers := flag.Int("users", 10, "Number of users to simulate")
	numSwipes := flag.Int("swipes", 100, "Number of total swipes")
	flag.Parse()

	tinderService, err := SetupTinderService(*redisAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🚀 Starting simulation with %d users and %d swipes...\n", *numUsers, *numSwipes)
	start := time.Now()

	runSwipesSimulation(tinderService, *numUsers, *numSwipes)

	fmt.Printf("✅ Simulation completed in %v\n", time.Since(start))
}

func SetupTinderService(redisAddr string) (*TinderService, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	tinderService := NewTinderService(redisClient)

	if err := redisClient.Ping(tinderService.ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %w", redisAddr, err)
	}

	if err := tinderService.LoadLuaFunction(); err != nil {
		return nil, fmt.Errorf("failed to load Redis script: %w", err)
	}

	if err := redisClient.FlushAll(tinderService.ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to flush Redis DB: %w", err)
	}

	return tinderService, nil
}
