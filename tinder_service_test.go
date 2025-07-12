package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupWithTestContainer(t *testing.T) *TinderService {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(10 * time.Second),
	}

	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start Redis container: %v", err)
	}

	t.Cleanup(func() {
		_ = redisC.Terminate(ctx)
	})

	host, err := redisC.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get Redis container host: %v", err)
	}

	port, err := redisC.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("failed to get Redis container port: %v", err)
	}

	redisAddr := host + ":" + port.Port()

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	tinderService := NewTinderService(redisClient)

	if err := redisClient.Ping(tinderService.ctx).Err(); err != nil {
		t.Fatalf("failed to ping Redis: %v", err)
	}

	if err := tinderService.LoadLuaFunction(); err != nil {
		t.Fatalf("failed to load Lua function: %v", err)
	}

	return tinderService
}

func TestSwipeMatch(t *testing.T) {
	tinderService := setupWithTestContainer(t)

	// Alice likes Bob first — no match expected yet
	matched, err := tinderService.SwipeAndCheckMatch("alice", "bob", SwipeLike)
	assert.NoError(t, err)
	assert.False(t, matched, "expected no match when only one side swiped like")

	// Bob likes Alice — this should create a match
	matched, err = tinderService.SwipeAndCheckMatch("bob", "alice", SwipeLike)
	assert.NoError(t, err)
	assert.True(t, matched, "expected a match when both like each other")

	// Verify matches are recorded for both users
	aliceMatches, err := tinderService.GetMatches("alice")
	assert.NoError(t, err)
	assert.Contains(t, aliceMatches, "bob")

	bobMatches, err := tinderService.GetMatches("bob")
	assert.NoError(t, err)
	assert.Contains(t, bobMatches, "alice")
}

func TestSwipeNoMatch(t *testing.T) {
	tinderService := setupWithTestContainer(t)

	// Charlie likes Dana, Dana dislikes Charlie — no match expected
	matched, err := tinderService.SwipeAndCheckMatch("charlie", "dana", SwipeLike)
	assert.NoError(t, err)
	assert.False(t, matched, "expected no match after first swipe")

	matched, err = tinderService.SwipeAndCheckMatch("dana", "charlie", SwipeDislike)
	assert.NoError(t, err)
	assert.False(t, matched, "expected no match when second swipe is dislike")

	// Confirm no matches exist for either user
	charlieMatches, err := tinderService.GetMatches("charlie")
	assert.NoError(t, err)
	assert.NotContains(t, charlieMatches, "dana")

	danaMatches, err := tinderService.GetMatches("dana")
	assert.NoError(t, err)
	assert.NotContains(t, danaMatches, "charlie")
}

func TestConcurrentSwipeAtomicity(t *testing.T) {
	tinderService := setupWithTestContainer(t)

	var wg sync.WaitGroup
	wg.Add(2)

	var (
		matchedA bool
		errA     error
		matchedB bool
		errB     error
	)

	go func() {
		defer wg.Done()
		matchedA, errA = tinderService.SwipeAndCheckMatch("Alice", "Bob", SwipeLike)
	}()

	go func() {
		defer wg.Done()
		matchedB, errB = tinderService.SwipeAndCheckMatch("Bob", "Alice", SwipeLike)
	}()

	wg.Wait()

	assert.NoError(t, errA)
	assert.NoError(t, errB)
	assert.True(t, matchedA || matchedB, "expected at least one swipe to detect a match due to atomic Redis Lua script")

	aMatches, err := tinderService.GetMatches("Alice")
	assert.NoError(t, err)
	bMatches, err := tinderService.GetMatches("Bob")
	assert.NoError(t, err)

	assert.Contains(t, aMatches, "Bob", "Alice should have Bob in matches")
	assert.Contains(t, bMatches, "Alice", "Bob should have Alice in matches")
}
