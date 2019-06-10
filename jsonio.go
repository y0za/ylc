package main

import (
	"encoding/json"
	"io"
)

type JSONIO struct {
	out io.Writer
}

func NewJSONIO(out io.Writer) *JSONIO {
	return &JSONIO{
		out: out,
	}
}

func (ji *JSONIO) Run(chMsgList chan MessageList) {
	go func() {
		for {
			data, ok := <-chMsgList
			if !ok {
				break
			}

			s, err := json.Marshal(data)
			if err != nil {
				break
			}
			ji.out.Write(s)
			ji.out.Write([]byte("\n"))
		}
	}()
}
