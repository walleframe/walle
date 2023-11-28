package wpb

// Deprecated: 废弃的生成命令，使用wpb进行pb解析序列化及zap代码生成。
// // 生成pb
// //go:generate wctl gen -c toproto
// // 生成proto的go代码
// //go:generate protoc --gogofaster_out=. --gogofaster_opt=paths=source_relative wpbrpc.proto
// // 生成rpc代码
// //go:generate wctl gen -c wzap -c wrpc

// new generate
//go:generate wctl gen -c wpb -c wrpc
