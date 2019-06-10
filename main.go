package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

var (
	googleClientID     string
	googleClientSecret string

	liveID = flag.String("liveid", "", "specify target YouTube Live id")
)

func main() {
	godotenv.Load()
	if googleClientID == "" {
		googleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	}
	if googleClientSecret == "" {
		googleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	}

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	config := NewConfig()

	tm := NewOAuthManager(googleClientID, googleClientSecret, config.TokenStore())
	ts, err := tm.TokenSource(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	if *liveID == "" {
		fmt.Print("Input live id: ")
		var lid string
		_, err = fmt.Scanf("%s\n", &lid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to scan live id: %v\n", err)
			return
		}
		liveID = &lid
	}

	chMsg, chErr, err := pollYoutubeLiveChatMessages(ctx, ts, *liveID)

	tui := NewTUI()
	go func() {
		err := tui.Run(chMsg)
		cancel()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case err = <-chErr:
			fmt.Fprintf(os.Stderr, "%v\n", err)
			cancel()
		}
	}
}
