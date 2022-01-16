package wpb

import (
	"github.com/aggronmagi/walle/network"
	"github.com/aggronmagi/walle/process/errcode"
)

type WPBSvc struct{}

var _ WSvcService = (*WPBSvc)(nil)

func (*WPBSvc) Add(ctx network.SessionContext, rq *AddRq, rs *AddRs) (err error) {
	for _, v := range rq.Params {
		rs.Value += v
	}
	return
}
func (*WPBSvc) Mul(ctx network.SessionContext, rq *MulRq, rs *MulRs) (err error) {
	rs.R = rq.A * rq.B
	return
}

func (*WPBSvc) Re(ctx network.SessionContext, rq *AddRq, rs *AddRs) (err error) {
	err = errcode.ErrUnkwon
	return
}

func (*WPBSvc) CallOneWay(ctx network.SessionContext, rq *AddRq) (err error) {
	return
}

// notify fun
func (*WPBSvc) NotifyFunc(ctx network.SessionContext, rq *AddRq) (err error) {
	return
}
