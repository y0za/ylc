package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func pollYoutubeLiveChatMessages(ctx context.Context, tokenSource oauth2.TokenSource, liveID string) (chan *youtube.LiveChatMessageListResponse, chan error, error) {
	ys, err := youtube.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create youtube service client: %v", err)
	}
	vs := youtube.NewVideosService(ys)
	lblc := vs.List("liveStreamingDetails")
	vResp, err := lblc.Context(ctx).Id(liveID).Do()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get live data (id = %s): %v", liveID, err)
	}
	if len(vResp.Items) == 0 || vResp.Items[0].LiveStreamingDetails == nil {
		return nil, nil, fmt.Errorf("failed to get live data (id = %s)", liveID)
	}

	chatID := vResp.Items[0].LiveStreamingDetails.ActiveLiveChatId
	lcms := youtube.NewLiveChatMessagesService(ys)
	lcmlc := lcms.List(chatID, "id,snippet,authorDetails")

	chRequest := make(chan RequestMessages)
	chResult := make(chan ResultMessages)
	go func() {
		chRequest <- RequestMessages{}
	}()

	chMsg := make(chan *youtube.LiveChatMessageListResponse)
	chErr := make(chan error)

	go func() {
	Loop:
		for {
			select {
			case <-ctx.Done():
				break Loop
			case req := <-chRequest:
				go func() {
					lcmResp, err := lcmlc.Context(ctx).PageToken(req.nextPageToken).Do()
					chResult <- ResultMessages{lcmResp, err}
				}()
			case result := <-chResult:
				if result.err != nil {
					chErr <- fmt.Errorf("failed to get chat messages (id = %s): %v", chatID, result.err)
					break Loop
				}
				chMsg <- result.lcmResp
				go func() {
					var defaultInterval time.Duration = 5 * time.Second
					interval := time.Duration(result.lcmResp.PollingIntervalMillis) * time.Millisecond
					if interval < defaultInterval {
						interval = defaultInterval
					}
					time.Sleep(interval)
					chRequest <- RequestMessages{result.lcmResp.NextPageToken}
				}()
			}
		}

		close(chMsg)
		close(chErr)
		close(chRequest)
		close(chResult)
	}()

	return chMsg, chErr, nil
}

type RequestMessages struct {
	nextPageToken string
}

type ResultMessages struct {
	lcmResp *youtube.LiveChatMessageListResponse
	err     error
}
