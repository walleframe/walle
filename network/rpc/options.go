package rpc

import (
	"time"

	"github.com/aggronmagi/walle/process"
	"github.com/aggronmagi/walle/process/metadata"
)

// AsyncResponseFilter 异步请回回复调用
type AsyncResponseFilter func(ctx process.Context, req, rsp interface{})

// CallOption rpc call options
//go:generate gogen option -n CallOption -f Call -o option.call.go
func walleCallOption() interface{} {
	return map[string]interface{}{
		// rpc call timeout
		"Timeout": time.Duration(0),
		// metadata
		"Metadata": metadata.MD(nil),
	}
}

// CallOption rpc call options
//go:generate gogen option -n AsyncCallOption -f Async -o option.async.go
func walleAsyncCallOption() interface{} {
	return map[string]interface{}{
		// rpc call timeout
		"Timeout": time.Duration(0),
		// metadata
		"Metadata": metadata.MD(nil),
		// response filter. NOTE: req only valid in Filter func.
		"ResponseFilter": AsyncResponseFilter(func(ctx process.Context, req, rsp interface{}) {
			ctx.Next(ctx)
		}),
		"WaitFilter": func(await func()) {
			go await()
		},
	}
}

// NoticeOption oneway rpc
//go:generate gogen option -n NoticeOption -f Notice -o option.notify.go
func walleNoticeCallOption() interface{} {
	return map[string]interface{}{
		// send message timeout
		"Timeout": time.Duration(0),
		// metadata
		"Metadata": metadata.MD(nil),
	}
}
