package main

import (
	"github.com/rivo/tview"
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
				for _, mes := range data.Items {
					row := tui.table.GetRowCount()

					if mes.Author != nil {
						// set message author
						tui.table.SetCell(
							row,
							0, // left column
							tview.NewTableCell(mes.Author.Name).SetMaxWidth(20),
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
