package example

import (
	"github.com/aggronmagi/walle/app"
	"github.com/aggronmagi/walle/kvstore"
)

type Project struct{
	Base map[string]interface{}
	ConfigCenter kvstore.Store
	Clients map[string]app.Service
}
