package process

import (
	"github.com/walleframe/walle/process/errcode"
	"go.uber.org/zap"
)

//go:generate gogen imake . -t=Process -r Process=Processer -o processer.go --merge
//go:generate mockgen -source processer.go -destination ../testpkg/mock_process/processer.go

type ProcessFilter func(pkg interface{}) (filter bool)

// Process 通用process 封装
type Process struct {
	// config
	Inner          *InnerOptions
	Opts           *ProcessOptions
	Filter         ProcessFilter
	dispatchData   DataDispatcherFunc
	dispatchPacket PacketDispatcherFunc
}

func NewProcess(inner *InnerOptions, opts *ProcessOptions) Process {
	p := Process{
		Inner: inner,
		Opts:  opts,
	}
	// 防止每次调用转换类型，申请堆
	p.dispatchPacket = p.innerPacket
	p.dispatchData = p.innerData
	return p
}

// OnRead 入口函数。接收数据处理
func (p *Process) OnRead(data []byte) (err error) {
	//p.Opts.FrameLogger.New("proc.read").Info("read size", zap.Int("len", len(data)))
	// dispatch chain
	err = p.Opts.DispatchDataFilter(data, p.innerData)
	if err != nil {
		p.Opts.FrameLogger.New("process.OnRead").Error("dispatch msg failed", zap.Error(err))
	}

	return
}

func (p *Process) innerData(data []byte) (err error) {
	// 解码网络包
	data = p.Opts.PacketEncode.Decode(data)
	// 反序列化网络包
	pkg := p.Opts.PacketPool.Get()
	err = p.Opts.PacketCodec.Unmarshal(data, pkg)
	if err != nil {
		p.Opts.FrameLogger.New("process.innerData").Error("unmarshal packet.Paket failed", zap.Error(err))
		return err
	}

	// rpc 请求回包
	if p.Filter != nil && p.Filter(pkg) {
		return
	}
	// 请求包
	return p.Opts.DispatchPacketFilter(pkg, p.dispatchPacket)
}

func (p *Process) innerPacket(pkg interface{}) (err error) {
	if p.Inner.Router == nil {
		err = errcode.ErrUnexpectedCode
		p.Opts.FrameLogger.New("process.innerPacket").Warn("unexcepted code: not set Router)", zap.Any("pkg", pkg))
		p.Opts.PacketPool.Put(pkg)
		return
	}

	// Request or Notice
	handlers, err := p.Inner.Router.GetHandlers(pkg)
	if err != nil {
		p.Opts.FrameLogger.New("process.innerPacket").Warn("get handler failed", zap.Any("pkg", pkg), zap.Error(err))
		p.Opts.PacketPool.Put(pkg)
		return err
	}

	// load limit
	if p.Opts.LoadLimitFilter(pkg, p.Inner.Load) {
		p.Opts.PacketPool.Put(pkg)
		p.Inner.Load.Dec()
		p.Opts.FrameLogger.New("process.innerPacket").Debug("process load limit", zap.Any("pkg", pkg))
		return
	}

	ctx := p.Inner.ContextPool.NewContext(p.Inner, p.Opts, pkg, handlers, true)
	ctx.Next(ctx)

	return
}
