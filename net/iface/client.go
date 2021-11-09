package iface

import (
	"github.com/aggronmagi/walle/net/process"
)

type Client interface {
	Link
}

type ClientContext interface {
	process.Context
	Link
}
