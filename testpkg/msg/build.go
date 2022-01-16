package msg

// 生成proto的go代码
//go:generate protoc --gogofaster_out=. --gogofaster_opt=paths=source_relative msg.proto
// 生成zap代码
//go:generate wctl gen -c wzap

