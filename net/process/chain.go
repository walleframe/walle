package process

// PacketDispatcherFunc 消息分发 - 未解包
type PacketDispatcherFunc func(data []byte) (err error)

// PacketUnmarshalFilter
type PacketDispatcherFilter func(data []byte, next PacketDispatcherFunc) (err error)

func PacketDispatcherChain(filters ...PacketDispatcherFilter) PacketDispatcherFilter {
	return func(data []byte, next PacketDispatcherFunc) (err error) {
		chain := func(cur PacketDispatcherFilter, next PacketDispatcherFunc) PacketDispatcherFunc {
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
func DefaultPacketFilter(data []byte, next PacketDispatcherFunc) (err error) {
	return next(data)
}
