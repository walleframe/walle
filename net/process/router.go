package process

import (
	"errors"
	"fmt"

	"github.com/aggronmagi/walle/net/packet"
)

// MiddlewareFunc is middleware or router functions
type MiddlewareFunc func(Context)

// RouterFunc router func is MiddlewareFunc
type RouterFunc = MiddlewareFunc

// Router 路由接口
type Router interface {
	GetHandlers(p *packet.Packet) (handlers []RouterFunc, err error)
}

var _ Router = &MixRouter{}

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
func (r *MixRouter) NoRouter(useGlobalMiddleware bool, rf RouterFunc, mid ...MiddlewareFunc) (err error) {
	if r.noCache != nil {
		err = fmt.Errorf("norouter %w", ErrRouterKeyRepated)
		return
	}
	if useGlobalMiddleware {
		r.noCache = append(r.noCache, r.middlewares...)
	}
	r.noCache = append(r.noCache, mid...)
	r.noCache = append(r.noCache, rf)
	return
}

// Method 使用string类型名字注册处理方法
func (r *MixRouter) Method(name string, rf RouterFunc, m ...MiddlewareFunc) (err error) {
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
func (r *MixRouter) RequestID(id uint32, rf RouterFunc, m ...MiddlewareFunc) (err error) {
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
func (r *MixRouter) GetHandlers(p *packet.Packet) (handlers []RouterFunc, err error) {
	if p == nil {
		err = ErrNotFoundRequest
		return
	}
	if r.handlersID != nil && p.ReservedRq > 0 {
		if node, ok := r.handlersID[p.ReservedRq]; ok {
			return node.funs, nil
		}
	}
	node, ok := r.handlers[p.Uri]
	if !ok {
		if r.noCache == nil {
			err = fmt.Errorf("uri %s %w", p.Uri, ErrRouterNotSupport)
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
