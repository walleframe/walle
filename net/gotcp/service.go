package gotcp

import (
	"context"
	"net"

	"github.com/aggronmagi/walle/app"
)

// GoTcpService implement app.Service interface
type GoTcpService struct {
	svr  *GoServer
	name string
	ln   net.Listener
}

func NewService(name string, opt ...ServerOption) app.Service {
	return &GoTcpService{
		name: name,
		svr:  NewServer(opt...),
	}
}

func (svc *GoTcpService) Name() string {
	return svc.name
}
func (svc *GoTcpService) Init() (err error) {
	return svc.svr.Listen("")
}
func (svc *GoTcpService) Start() (err error) {
	go svc.svr.Serve(nil)
	return
}
func (svc *GoTcpService) Stop() {
	svc.svr.Shutdown(context.Background())
	return
}
func (svc *GoTcpService) Finish() {
	return
}
