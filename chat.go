package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type MessageType string

const (
	ChatEnded   MessageType = "chatEndedEvent"
	SuperChat   MessageType = "superChatEvent"
	TextMessage MessageType = "textMessageEvent"
)

type Author struct {
	ChannelID       string `json:"channelId"`
	Name            string `json:"name"`
	ProfileImageURL string `json:"profileImageUrl"`
	IsChatModerator bool   `json:"isChatModerator"`
	IsChatOwner     bool   `json:"isChatOwner"`
	IsChatSponsor   bool   `json:"isChatSponsor"`
	IsVerified      bool   `json:"isVerified"`
}

type Message struct {
	ID     string      `json:"id"`
	Author *Author     `json:"author"`
	Text   string      `json:"text"`
	Type   MessageType `json:"type"`
}

type MessageList struct {
	Items []Message `json:"items"`
}

func pollYoutubeLiveChatMessages(ctx context.Context, tokenSource oauth2.TokenSource, liveID string) (chan MessageList, chan error, error) {
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

	chRequest := make(chan requestMessages)
	chResult := make(chan resultMessages)
	go func() {
		chRequest <- requestMessages{}
	}()

	chMsgList := make(chan MessageList)
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
					chResult <- resultMessages{lcmResp, err}
				}()
			case result := <-chResult:
				if result.err != nil {
					chErr <- fmt.Errorf("failed to get chat messages (id = %s): %v", chatID, result.err)
					break Loop
				}

				chMsgList <- messageListFromResp(result.lcmResp)
				go func() {
					var defaultInterval time.Duration = 5 * time.Second
					interval := time.Duration(result.lcmResp.PollingIntervalMillis) * time.Millisecond
					if interval < defaultInterval {
						interval = defaultInterval
					}
					time.Sleep(interval)
					chRequest <- requestMessages{result.lcmResp.NextPageToken}
				}()
			}
		}

		close(chMsgList)
		close(chErr)
		close(chRequest)
		close(chResult)
	}()

	return chMsgList, chErr, nil
}

type requestMessages struct {
	nextPageToken string
}

type resultMessages struct {
	lcmResp *youtube.LiveChatMessageListResponse
	err     error
}

func messageListFromResp(resp *youtube.LiveChatMessageListResponse) MessageList {
	var mList MessageList
	for _, data := range resp.Items {
		switch data.Snippet.Type {
		case string(ChatEnded), string(SuperChat), string(TextMessage): // ok
		default: // ignore other
			continue
		}

		var author *Author
		if data.AuthorDetails != nil {
			author = &Author{
				ChannelID:       data.AuthorDetails.ChannelId,
				Name:            data.AuthorDetails.DisplayName,
				ProfileImageURL: data.AuthorDetails.ProfileImageUrl,
				IsChatModerator: data.AuthorDetails.IsChatModerator,
				IsChatOwner:     data.AuthorDetails.IsChatOwner,
				IsChatSponsor:   data.AuthorDetails.IsChatSponsor,
				IsVerified:      data.AuthorDetails.IsVerified,
			}
		}
		mList.Items = append(mList.Items, Message{
			ID:     data.Id,
			Author: author,
			Text:   data.Snippet.DisplayMessage,
			Type:   MessageType(data.Snippet.Type),
		})
	}
	return mList
}
