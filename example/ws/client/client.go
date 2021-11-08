package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aggronmagi/walle/net/process"

	. "github.com/aggronmagi/walle/net/ws"
)

type rpcRQ struct {
	M int `json:"m"`
	N int `json:"n"`
}

type rpcRS struct {
	V1 int `json:"v1"`
	V2 int `json:"v2"`
}

func main() {
	cli, err := NewClient(fmt.Sprintf("ws://localhost:8080/ws"), http.Header{
		"name": []string{"xxx"},
	}, nil,
		process.WithMsgCodec(process.MessageCodecJSON),
	)
	if err != nil {
		panic(err)
	}
	rq := &rpcRQ{108, 72}

	call := func(uri string) {
		ctx := context.Background()
		rs := &rpcRS{}
		err = cli.Call(ctx, uri, rq, rs, process.NewCallOptions(
			process.WithCallOptionsTimeout(time.Second),
		))
		fmt.Println("call rpc ", uri, err, rs)
	}
	call("f1")
	call("f2")
	call("f3")
}
