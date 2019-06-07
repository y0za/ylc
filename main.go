package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	googleClientID     string
	googleClientSecret string
)

func main() {
	godotenv.Load()
	if googleClientID == "" {
		googleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	}
	if googleClientSecret == "" {
		googleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	}

	ctx := context.Background()

	config := NewConfig()

	tm := NewOAuthManager(googleClientID, googleClientSecret, config.TokenStore())
	ts, err := tm.TokenSource(ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Print("Input live id: ")
	var liveID string
	_, err = fmt.Scanf("%s\n", &liveID)
	if err != nil {
		log.Fatalf("failed to scan live id: %v", err)
	}

	err = pollYoutubeLiveChatMessages(ctx, ts, liveID, os.Stdout)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
