package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func pollYoutubeLiveChatMessages(ctx context.Context, tokenSource oauth2.TokenSource, liveID string, out io.Writer) error {
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

	cRequest := make(chan RequestMessages)
	cResult := make(chan ResultMessages)
	go func() {
		cRequest <- RequestMessages{}
	}()

Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		case req := <-cRequest:
			go func() {
				lcmResp, err := lcmlc.Context(ctx).PageToken(req.nextPageToken).Do()
				cResult <- ResultMessages{lcmResp, err}
			}()
		case result := <-cResult:
			if result.err != nil {
				return fmt.Errorf("failed to get chat messages (id = %s): %v", chatID, result.err)
			}
			for _, mes := range result.lcmResp.Items {
				fmt.Fprintf(out, "%s: %s\n", mes.AuthorDetails.DisplayName, mes.Snippet.DisplayMessage)
			}
			go func() {
				var defaultInterval time.Duration = 5 * time.Second
				interval := time.Duration(result.lcmResp.PollingIntervalMillis) * time.Millisecond
				if interval < defaultInterval {
					interval = defaultInterval
				}
				time.Sleep(interval)
				cRequest <- RequestMessages{result.lcmResp.NextPageToken}
			}()
		}
	}

	return nil
}

type RequestMessages struct {
	nextPageToken string
}

type ResultMessages struct {
	lcmResp *youtube.LiveChatMessageListResponse
	err     error
}
