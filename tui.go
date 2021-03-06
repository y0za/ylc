package main

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const (
	maxRowCount = 500
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
func (tui *TUI) receiveMessagesLoop(chMsgList chan MessageList) {
	go func() {
		for {
			data, ok := <-chMsgList
			if !ok {
				break
			}
			tui.app.QueueUpdateDraw(func() {
				overflowed := (tui.table.GetRowCount() + len(data.Items)) - maxRowCount
				if overflowed > 0 {
					// remove overflowed rows
					for i := 0; i < overflowed; i++ {
						tui.table.RemoveRow(0)
					}
				}

				for _, mes := range data.Items {
					row := tui.table.GetRowCount()

					if mes.Author != nil {
						ac := authorColor(mes.Author)
						// set message author
						tui.table.SetCell(
							row,
							0, // left column
							tview.NewTableCell(mes.Author.Name).SetMaxWidth(20).SetTextColor(ac),
						)
					}

					// set message text
					tui.table.SetCell(
						row,
						1, // right column
						tview.NewTableCell(mes.Text),
					)
				}
			})
		}
	}()
}

func (tui *TUI) Run(chMsgList chan MessageList) error {
	tui.receiveMessagesLoop(chMsgList)
	tui.app.SetRoot(tui.table, true).SetFocus(tui.table)
	return tui.app.Run()
}

func authorColor(author *Author) tcell.Color {
	switch {
	case author.IsChatOwner:
		return tcell.ColorYellow
	case author.IsChatModerator:
		return tcell.ColorSkyblue
	case author.IsVerified:
		return tcell.ColorOrange
	case author.IsChatSponsor:
		return tcell.ColorLightGreen
	default:
		return tcell.ColorWhite
	}
}
