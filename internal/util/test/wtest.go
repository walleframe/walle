package test

// FuncCall 用于测试函数调用。
//go:generate mockgen -source wtest.go -package test -destination gentest.go
type FuncCall interface{
	Call(v ...interface{})
}
