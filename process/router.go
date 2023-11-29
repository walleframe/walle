package process

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/walleframe/walle/process/errcode"
	"github.com/walleframe/walle/process/packet"
)

//go:generate mockgen -source router.go -destination ../testpkg/mock_process/router.go

// MiddlewareFunc is middleware or router functions
type MiddlewareFunc func(Context)

// RouterFunc router func is MiddlewareFunc
type RouterFunc = MiddlewareFunc

// Router 路由接口
type Router interface {
	// Use 设置中间件，在Use之后注册的接口都会使用此中间件
	Use(m ...MiddlewareFunc)
	// NoRouter 未注册路由的默认处理函数
	NoRouter(rf RouterFunc, mid ...MiddlewareFunc) (err error)
	// Register 注册路由
	Register(uri interface{}, rf RouterFunc, m ...MiddlewareFunc) (err error)
	// GetHandlers 获取处理函数接口
	GetHandlers(p interface{}) (handlers []RouterFunc, err error)
}

var defaultRouter Router = &MixRouter{}

func GetRouter() Router {
	return defaultRouter
}

func SetRouter(r Router) {
	defaultRouter = r
}

type routerNode struct {
	funs []RouterFunc
}

// MixRouter 混合路由. 优先使用RequestID.(内部功能也使用RequestID)
// 路由必须先注册,再使用. 已经开始使用之后,路由不能再修改.(内部无锁)
type MixRouter struct {
	middlewares []MiddlewareFunc
	handlers    map[string]*routerNode
	handlersID  map[uint32]*routerNode
	noCache     []MiddlewareFunc
	startUse    bool
}

// Use 设置全局中间件
func (r *MixRouter) Use(m ...MiddlewareFunc) {
	if len(m) < 1 || r == nil {
		return
	}
	r.middlewares = append(r.middlewares, m...)
}

// NoRouter 未设置路由请求
func (r *MixRouter) NoRouter(rf RouterFunc, mid ...MiddlewareFunc) (err error) {
	if r.noCache != nil {
		err = fmt.Errorf("norouter %w", ErrRouterKeyRepated)
		return
	}
	r.noCache = append(r.noCache, r.middlewares...)
	r.noCache = append(r.noCache, mid...)
	r.noCache = append(r.noCache, rf)
	return
}

func (r *MixRouter) Register(uri interface{}, rf RouterFunc, m ...MiddlewareFunc) (err error) {
	switch v := uri.(type) {
	case string:
		return r.regMethod(v, rf, m...)
	case uint32:
		return r.regRequestID(v, rf, m...)
	case int:
		return r.regRequestID(uint32(v), rf, m...)
	default:
		// 未支持的类型
		return errcode.WrapError(errcode.ErrUnexpectedCode,
			fmt.Errorf("register uri type[%s] not suppert", reflect.TypeOf(uri).Name()),
		)
	}
}

// Method 使用string类型名字注册处理方法
func (r *MixRouter) regMethod(name string, rf RouterFunc, m ...MiddlewareFunc) (err error) {
	if r.handlers == nil {
		r.handlers = make(map[string]*routerNode)
	}
	if _, ok := r.handlers[name]; ok {
		return fmt.Errorf("method %s %w", name, ErrRouterKeyRepated)
	}
	node := new(routerNode)
	node.funs = make([]MiddlewareFunc, len(r.middlewares)+len(m)+1)
	copy(node.funs, r.middlewares)
	if len(m) > 0 {
		copy(node.funs[len(r.middlewares):], m)
	}
	node.funs[len(node.funs)-1] = rf
	r.handlers[name] = node

	return
}

// RequestID 使用请求id注册处理方法
func (r *MixRouter) regRequestID(id uint32, rf RouterFunc, m ...MiddlewareFunc) (err error) {
	if r.handlersID == nil {
		r.handlersID = make(map[uint32]*routerNode)
	}
	if _, ok := r.handlersID[id]; ok {
		return fmt.Errorf("RQID %d %w", id, ErrRouterKeyRepated)
	}
	node := new(routerNode)
	node.funs = make([]MiddlewareFunc, len(r.middlewares)+len(m)+1)
	copy(node.funs, r.middlewares)
	if len(m) > 0 {
		copy(node.funs[len(r.middlewares):], m)
	}
	node.funs[len(node.funs)-1] = rf
	r.handlersID[id] = node

	return
}

// GetHandlers 获取请求对应处理函数
func (r *MixRouter) GetHandlers(in interface{}) (handlers []RouterFunc, err error) {
	if in == nil {
		err = ErrNotFoundRequest
		return
	}
	p := in.(*packet.Packet)
	if r.handlersID != nil && p.MsgID() > 0 {
		if node, ok := r.handlersID[p.MsgID()]; ok {
			return node.funs, nil
		}
	}
	node, ok := r.handlers[p.URI()]
	if !ok {
		if r.noCache == nil {
			err = fmt.Errorf("uri %s %w", p.URI(), ErrRouterNotSupport)
			return
		}
		handlers = r.noCache
		return
	}
	handlers = node.funs
	return
}

// 路由返回通用错误
var (
	ErrRouterKeyRepated = errors.New("router key repeated")
	ErrRouterNotSupport = errors.New("router not found")
	ErrNotFoundRequest  = errors.New("no request packet")
)
