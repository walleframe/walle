package process

// DataDispatcherFunc 消息分发 - 未解包
type DataDispatcherFunc func(data []byte) (err error)

// PacketUnmarshalFilter
type DataDispatcherFilter func(data []byte, next DataDispatcherFunc) (err error)

func DataDispatcherChain(filters ...DataDispatcherFilter) DataDispatcherFilter {
	return func(data []byte, next DataDispatcherFunc) (err error) {
		chain := func(cur DataDispatcherFilter, next DataDispatcherFunc) DataDispatcherFunc {
			return func(data []byte) (err error) {
				return cur(data, next)
			}
		}
		c := next
		for i := len(filters) - 1; i >= 0; i-- {
			c = chain(filters[i], c)
		}
		return c(data)
	}
}

// DefaultPacketDispatcher default packet dispatch filter
func DefaultDataFilter(data []byte, next DataDispatcherFunc) (err error) {
	return next(data)
}

// PacketDispatcherFunc 消息分发 - 未解包
type PacketDispatcherFunc func(pkg interface{}) (err error)

// PacketUnmarshalFilter
type PacketDispatcherFilter func(pkg interface{}, next PacketDispatcherFunc) (err error)

func PacketDispatcherChain(filters ...PacketDispatcherFilter) PacketDispatcherFilter {
	return func(pkg interface{}, next PacketDispatcherFunc) (err error) {
		chain := func(cur PacketDispatcherFilter, next PacketDispatcherFunc) PacketDispatcherFunc {
			return func(pkg interface{}) (err error) {
				return cur(pkg, next)
			}
		}
		c := next
		for i := len(filters) - 1; i >= 0; i-- {
			c = chain(filters[i], c)
		}
		return c(pkg)
	}
}

// DefaultPacketDispatcher default packet dispatch filter
func DefaultPacketFilter(pkg interface{}, next PacketDispatcherFunc) (err error) {
	return next(pkg)
}
