package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

var (
	rndMu sync.Mutex
	rnd   = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func randIntn(n int) int {
	rndMu.Lock()
	defer rndMu.Unlock()
	return rnd.Intn(n)
}

func randFloat64() float64 {
	rndMu.Lock()
	defer rndMu.Unlock()
	return rnd.Float64()
}

func runSwipesSimulation(tinder *TinderService, numUsers, numSwipes int) {
	users := make([]string, numUsers)
	for i := range numUsers {
		users[i] = fmt.Sprintf("user%d", i+1)
	}

	var wg sync.WaitGroup

	for range numSwipes {
		wg.Add(1)
		go func() {
			defer wg.Done()

			fromIdx := randIntn(numUsers)
			toIdx := randIntn(numUsers)

			for toIdx == fromIdx {
				toIdx = randIntn(numUsers)
			}

			from := users[fromIdx]
			to := users[toIdx]

			swipe := SwipeDislike
			swipeEmoji := "❌"
			if randFloat64() < 0.8 {
				swipe = SwipeLike
				swipeEmoji = "❤️"
			}

			matched, err := tinder.SwipeAndCheckMatch(from, to, swipe)
			if err != nil {
				log.Printf("⚠️ Error: 👤 %s -> 👤 %s (%s): %v", from, to, swipe, err)
				return
			}

			log.Printf("👤 %s 👉 %s %s 👤 %s\n", from, swipeEmoji, swipe, to)

			if matched {
				log.Printf("💘 MATCH! 👤 %s ❤️ 👤 %s\n", from, to)
			}
		}()
	}

	wg.Wait()

	log.Println("\n📊 --- MATCH SUMMARY ---")
	for _, user := range users {
		matches := tinder.client.SMembers(tinder.ctx, "matches:"+user).Val()
		log.Printf("👤 %s matched with %d users\n", user, len(matches))
	}
}
