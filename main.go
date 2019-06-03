package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

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

	config := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes: []string{
			"https://www.googleapis.com/auth/youtube",
			"https://www.googleapis.com/auth/youtube.readonly",
			"https://www.googleapis.com/auth/youtube.force-ssl",
		},
	}

	ctx := context.Background()
	token, err := requestOAuthToken(ctx, config)
	if err != nil {
		log.Fatalf("%v", err)
	}

	err = getYoutubeLiveChatMessages(ctx, config.TokenSource(ctx, token))
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func requestOAuthToken(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	fmt.Println("Access to this URL and get auth code.")
	fmt.Println(config.AuthCodeURL(""))
	fmt.Print("Input auth code: ")
	var code string
	_, err := fmt.Scanf("%s\n", &code)
	if err != nil {
		return nil, fmt.Errorf("failed to scan auth code: %v", err)
	}

	return config.Exchange(ctx, code)
}

func getYoutubeLiveChatMessages(ctx context.Context, tokenSource oauth2.TokenSource) error {
	fmt.Print("Input live id: ")
	var liveID string
	_, err := fmt.Scanf("%s\n", &liveID)
	if err != nil {
		return fmt.Errorf("failed to scan live id: %v", err)
	}

	ys, err := youtube.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return fmt.Errorf("failed to create youtube service client: %v", err)
	}
	vs := youtube.NewVideosService(ys)
	lblc := vs.List("liveStreamingDetails")
	vResp, err := lblc.Context(ctx).Id(liveID).Do()
	if err != nil {
		return fmt.Errorf("failed to get live data (id = %s): %v", liveID, err)
	}
	if len(vResp.Items) == 0 || vResp.Items[0].LiveStreamingDetails == nil {
		return fmt.Errorf("failed to get live data (id = %s)", liveID)
	}

	chatID := vResp.Items[0].LiveStreamingDetails.ActiveLiveChatId
	lcms := youtube.NewLiveChatMessagesService(ys)
	lcmlc := lcms.List(chatID, "id,snippet,authorDetails")
	lcmResp, err := lcmlc.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get chat messages (id = %s): %v", chatID, err)
	}

	for _, mes := range lcmResp.Items {
		fmt.Printf("%s: %s\n", mes.AuthorDetails.DisplayName, mes.Snippet.DisplayMessage)
	}
	return nil
}
