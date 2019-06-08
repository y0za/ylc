package main

import (
	"github.com/rivo/tview"
	"google.golang.org/api/youtube/v3"
)

type TUI struct {
	app   *tview.Application
	table *tview.Table
}

func NewTUI() *TUI {
	return &TUI{
		app:   tview.NewApplication(),
		table: tview.NewTable(),
	}
}

// ReceiveMessagesLoop receives new messages and appends them to table.
func (tui *TUI) receiveMessagesLoop(chMsg chan *youtube.LiveChatMessageListResponse) {
	go func() {
		for {
			data, ok := <-chMsg
			if !ok {
				break
			}
			tui.app.QueueUpdateDraw(func() {
				for _, mes := range data.Items {
					row := tui.table.GetRowCount()

					// set message author
					tui.table.SetCell(
						row,
						0, // left column
						tview.NewTableCell(mes.AuthorDetails.DisplayName).SetMaxWidth(20),
					)

					// set message text
					tui.table.SetCell(
						row,
						1, // right column
						tview.NewTableCell(mes.Snippet.DisplayMessage),
					)
				}
			})
		}
	}()
}

func (tui *TUI) Run(chMsg chan *youtube.LiveChatMessageListResponse) error {
	tui.receiveMessagesLoop(chMsg)
	tui.app.SetRoot(tui.table, true).SetFocus(tui.table)
	return tui.app.Run()
}
