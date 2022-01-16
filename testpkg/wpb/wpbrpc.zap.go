// Generate by wctl plugin(wzap). DO NOT EDIT.
package wpb

import (
	"strconv"

	"go.uber.org/zap/zapcore"
)

// test message
func (x *TestMsg) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt32("V1", x.V1)
	enc.AddString("V2", x.V2)
	return nil
}

type ZapArrayTestMsg []*TestMsg

func (x ZapArrayTestMsg) MarshalLogArray(ae zapcore.ArrayEncoder) error {
	for _, v := range x {
		ae.AppendObject(v)
	}
	return nil
}

// multiply rq
func (x *MulRq) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt32("A", x.A)
	enc.AddInt32("B", x.B)
	return nil
}

type ZapArrayMulRq []*MulRq

func (x ZapArrayMulRq) MarshalLogArray(ae zapcore.ArrayEncoder) error {
	for _, v := range x {
		ae.AppendObject(v)
	}
	return nil
}

// multiply result
func (x *MulRs) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt32("R", x.R)
	return nil
}

type ZapArrayMulRs []*MulRs

func (x ZapArrayMulRs) MarshalLogArray(ae zapcore.ArrayEncoder) error {
	for _, v := range x {
		ae.AppendObject(v)
	}
	return nil
}

// add request
func (x *AddRq) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddArray("Params", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, v := range x.Params {
			ae.AppendInt64(v)
		}
		return nil
	}))
	return nil
}

type ZapArrayAddRq []*AddRq

func (x ZapArrayAddRq) MarshalLogArray(ae zapcore.ArrayEncoder) error {
	for _, v := range x {
		ae.AppendObject(v)
	}
	return nil
}

// add reply
func (x *AddRs) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("Value", x.Value)
	return nil
}

type ZapArrayAddRs []*AddRs

func (x ZapArrayAddRs) MarshalLogArray(ae zapcore.ArrayEncoder) error {
	for _, v := range x {
		ae.AppendObject(v)
	}
	return nil
}
func (x *ANtf) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddObject("F1", zapcore.ObjectMarshalerFunc(func(oe zapcore.ObjectEncoder) error {
		for k, v := range x.F1 {
			oe.AddInt32(strconv.FormatInt(int64(k), 10), v)
		}
		return nil
	}))
	return nil
}

type ZapArrayANtf []*ANtf

func (x ZapArrayANtf) MarshalLogArray(ae zapcore.ArrayEncoder) error {
	for _, v := range x {
		ae.AppendObject(v)
	}
	return nil
}
func (x *BNtf) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddObject("F1", x.F1)
	enc.AddArray("F2", ZapArrayANtf(x.F2))
	return nil
}

type ZapArrayBNtf []*BNtf

func (x ZapArrayBNtf) MarshalLogArray(ae zapcore.ArrayEncoder) error {
	for _, v := range x {
		ae.AppendObject(v)
	}
	return nil
}
