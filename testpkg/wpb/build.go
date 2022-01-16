package wpb

// 生成pb
//go:generate wctl gen -c toproto
// 生成proto的go代码
//go:generate protoc --gogofaster_out=. --gogofaster_opt=paths=source_relative wpbrpc.proto
// 生成rpc代码
//go:generate wctl gen -c wzap -c wrpc
