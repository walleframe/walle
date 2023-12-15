// Generate by wctl plugin(wrpc). DO NOT EDIT.
package wpb

import (
	"github.com/walleframe/walle/network"
	"github.com/walleframe/walle/network/rpc"
	"github.com/walleframe/walle/process"

	"context"
)

// w_svc method uri define
const (
	__WSvcAdd        = "/add"
	__WSvcMul        = "/mul"
	__WSvcRe         = "/re"
	__WSvcCallOneWay = "/call_one_way"
	__WSvcNotifyFunc = "/notify_func"
)

type WSvcService interface {
	// add method
	Add(ctx network.SessionContext, rq *AddRq, rs *AddRs) (err error)
	// mul method
	Mul(ctx network.SessionContext, rq *MulRq, rs *MulRs) (err error)
	// will return error
	Re(ctx network.SessionContext, rq *AddRq, rs *AddRs) (err error)
	// oneway
	CallOneWay(ctx network.SessionContext, rq *AddRq) (err error)
	// notify fun
	NotifyFunc(ctx network.SessionContext, rq *AddRq) (err error)
}

func RegisterWSvcService(router process.Router, s WSvcService) {
	svc := &wWSvcService{svc: s}
	router.Register(__WSvcAdd, svc.Add)
	router.Register(__WSvcMul, svc.Mul)
	router.Register(__WSvcRe, svc.Re)
	router.Register(__WSvcCallOneWay, svc.CallOneWay)
	router.Register(__WSvcNotifyFunc, svc.NotifyFunc)
}

type WSvcClient interface {
	// add method
	Add(ctx context.Context, rq *AddRq, opts ...rpc.CallOption) (rs *AddRs, err error)
	AddAsync(ctx context.Context, rq *AddRq, resp func(ctx process.Context, rs *AddRs, err error), opts ...rpc.AsyncCallOption) (err error)
	// mul method
	Mul(ctx context.Context, rq *MulRq, opts ...rpc.CallOption) (rs *MulRs, err error)
	MulAsync(ctx context.Context, rq *MulRq, resp func(ctx process.Context, rs *MulRs, err error), opts ...rpc.AsyncCallOption) (err error)
	// will return error
	Re(ctx context.Context, rq *AddRq, opts ...rpc.CallOption) (rs *AddRs, err error)
	ReAsync(ctx context.Context, rq *AddRq, resp func(ctx process.Context, rs *AddRs, err error), opts ...rpc.AsyncCallOption) (err error)
	// oneway
	CallOneWay(ctx context.Context, rq *AddRq, opts ...rpc.CallOption) (err error)
	// notify fun
	NotifyFunc(ctx context.Context, rq *AddRq, opts ...rpc.NoticeOption) (err error)
}

type wWSvcService struct {
	svc WSvcService
}

func (s *wWSvcService) Add(c process.Context) {
	ctx := c.(network.SessionContext)
	rq := AddRq{}
	rs := AddRs{}
	err := ctx.Bind(&rq)
	if err != nil {
		ctx.Respond(ctx, err, nil)
		return
	}
	err = s.svc.Add(ctx, &rq, &rs)
	if err != nil {
		ctx.Respond(ctx, err, nil)
		return
	}
	ctx.Respond(ctx, &rs, nil)
	return
}

func (s *wWSvcService) Mul(c process.Context) {
	ctx := c.(network.SessionContext)
	rq := MulRq{}
	rs := MulRs{}
	err := ctx.Bind(&rq)
	if err != nil {
		ctx.Respond(ctx, err, nil)
		return
	}
	err = s.svc.Mul(ctx, &rq, &rs)
	if err != nil {
		ctx.Respond(ctx, err, nil)
		return
	}
	ctx.Respond(ctx, &rs, nil)
	return
}

func (s *wWSvcService) Re(c process.Context) {
	ctx := c.(network.SessionContext)
	rq := AddRq{}
	rs := AddRs{}
	err := ctx.Bind(&rq)
	if err != nil {
		ctx.Respond(ctx, err, nil)
		return
	}
	err = s.svc.Re(ctx, &rq, &rs)
	if err != nil {
		ctx.Respond(ctx, err, nil)
		return
	}
	ctx.Respond(ctx, &rs, nil)
	return
}

func (s *wWSvcService) CallOneWay(c process.Context) {
	ctx := c.(network.SessionContext)
	rq := AddRq{}
	err := ctx.Bind(&rq)
	if err != nil {
		ctx.Respond(ctx, err, nil)
		return
	}
	err = s.svc.CallOneWay(ctx, &rq)
	if err != nil {
		ctx.Respond(ctx, err, nil)
		return
	}
	ctx.Respond(ctx, nil, nil)
	return
}

func (s *wWSvcService) NotifyFunc(c process.Context) {
	ctx := c.(network.SessionContext)
	rq := AddRq{}
	err := ctx.Bind(&rq)
	if err != nil {
		return
	}
	err = s.svc.NotifyFunc(ctx, &rq)
	if err != nil {
		return
	}
	return
}

type wWSvcClient struct {
	cli network.Client
}

func NewWSvcClient(cli network.Client) WSvcClient {
	return &wWSvcClient{
		cli: cli,
	}
}

func (c *wWSvcClient) Add(ctx context.Context, rq *AddRq, opts ...rpc.CallOption) (rs *AddRs, err error) {
	cc := rpc.NewCallOptions(opts...)
	rs = &AddRs{}
	err = c.cli.Call(ctx, __WSvcAdd, rq, rs, cc)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *wWSvcClient) AddAsync(ctx context.Context, rq *AddRq,
	rf func(ctx process.Context, rs *AddRs, err error),
	opts ...rpc.AsyncCallOption) (err error) {
	cc := rpc.NewAsyncCallOptions(opts...)
	err = c.cli.AsyncCall(ctx, __WSvcAdd, rq, func(c process.Context) {
		rs := &AddRs{}
		err := c.Bind(rs)
		if err != nil {
			rf(c, nil, err)
			return
		}
		rf(c, rs, nil)
		return
	}, cc)
	if err != nil {
		return err
	}
	return nil
}

func (c *wWSvcClient) Mul(ctx context.Context, rq *MulRq, opts ...rpc.CallOption) (rs *MulRs, err error) {
	cc := rpc.NewCallOptions(opts...)
	rs = &MulRs{}
	err = c.cli.Call(ctx, __WSvcMul, rq, rs, cc)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *wWSvcClient) MulAsync(ctx context.Context, rq *MulRq,
	rf func(ctx process.Context, rs *MulRs, err error),
	opts ...rpc.AsyncCallOption) (err error) {
	cc := rpc.NewAsyncCallOptions(opts...)
	err = c.cli.AsyncCall(ctx, __WSvcMul, rq, func(c process.Context) {
		rs := &MulRs{}
		err := c.Bind(rs)
		if err != nil {
			rf(c, nil, err)
			return
		}
		rf(c, rs, nil)
		return
	}, cc)
	if err != nil {
		return err
	}
	return nil
}

func (c *wWSvcClient) Re(ctx context.Context, rq *AddRq, opts ...rpc.CallOption) (rs *AddRs, err error) {
	cc := rpc.NewCallOptions(opts...)
	rs = &AddRs{}
	err = c.cli.Call(ctx, __WSvcRe, rq, rs, cc)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *wWSvcClient) ReAsync(ctx context.Context, rq *AddRq,
	rf func(ctx process.Context, rs *AddRs, err error),
	opts ...rpc.AsyncCallOption) (err error) {
	cc := rpc.NewAsyncCallOptions(opts...)
	err = c.cli.AsyncCall(ctx, __WSvcRe, rq, func(c process.Context) {
		rs := &AddRs{}
		err := c.Bind(rs)
		if err != nil {
			rf(c, nil, err)
			return
		}
		rf(c, rs, nil)
		return
	}, cc)
	if err != nil {
		return err
	}
	return nil
}

func (c *wWSvcClient) CallOneWay(ctx context.Context, rq *AddRq, opts ...rpc.CallOption) (err error) {
	cc := rpc.NewCallOptions(opts...)
	err = c.cli.Call(ctx, __WSvcCallOneWay, rq, nil, cc)
	if err != nil {
		return err
	}
	return
}

func (c *wWSvcClient) NotifyFunc(ctx context.Context, rq *AddRq, opts ...rpc.NoticeOption) (err error) {
	cc := rpc.NewNoticeOptions(opts...)
	err = c.cli.Notify(ctx, __WSvcNotifyFunc, rq, cc)
	if err != nil {
		return err
	}
	return
}
