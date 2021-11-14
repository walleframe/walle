package gotcp

import (
	"context"
	"net"

	"github.com/aggronmagi/walle/app"
)

// GNetService implement app.Service interface
type GNetService struct {
	svr  *GoServer
	name string
	ln   net.Listener
}

func NewService(name string, opt ...ServerOption) app.Service {
	return &GNetService{
		name: name,
		svr:  NewServer(opt...),
	}
}

func (svc *GNetService) Name() string {
	return svc.name
}
func (svc *GNetService) Init() (err error) {
	return svc.svr.Listen("")
}
func (svc *GNetService) Start() (err error) {
	go svc.svr.Serve(nil)
	return
}
func (svc *GNetService) Stop() {
	svc.svr.Shutdown(context.Background())
	return
}
func (svc *GNetService) Finish() {
	return
}
