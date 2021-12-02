package zaplog

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogFields 日志字段
type LogFields struct {
	entity *LogEntities
}

// Any takes a key and an arbitrary value and chooses the best way to represent
// them as a field, falling back to a reflection-based approach only if
// necessary.
//
// Since byte/uint8 and rune/int32 are aliases, Any can't differentiate between
// them. To minimize surprises, []byte values are treated as binary blobs, byte
// values are treated as uint8, and runes are always treated as integers.
func (fields *LogFields) Any(key string, value interface{}) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Any(key, value))
	return fields
}

// Binary constructs a field that carries an opaque binary blob.
//
// Binary data is serialized in an encoding-appropriate format. For example,
// zap's JSON encoder base64-encodes binary blobs. To log UTF-8 encoded text,
// use ByteString.
func (fields *LogFields) Binary(key string, val []byte) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Binary(key, val))
	return fields
}

// Bool constructs a field that carries a bool.
func (fields *LogFields) Bool(key string, val bool) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Bool(key, val))
	return fields
}

// Boolp constructs a field that carries a *bool. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Boolp(key string, val *bool) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Boolp(key, val))
	return fields
}

// ByteString constructs a field that carries UTF-8 encoded text as a []byte.
// To log opaque binary blobs (which aren't necessarily valid UTF-8), use
// Binary.
func (fields *LogFields) ByteString(key string, val []byte) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.ByteString(key, val))
	return fields
}

// Complex128 constructs a field that carries a complex number. Unlike most
// numeric fields, this costs an allocation (to convert the complex128 to
// interface{}).
func (fields *LogFields) Complex128(key string, val complex128) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Complex128(key, val))
	return fields
}

// Complex128p constructs a field that carries a *complex128. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Complex128p(key string, val *complex128) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Complex128p(key, val))
	return fields
}

// Complex64 constructs a field that carries a complex number. Unlike most
// numeric fields, this costs an allocation (to convert the complex64 to
// interface{}).
func (fields *LogFields) Complex64(key string, val complex64) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Complex64(key, val))
	return fields
}

// Complex64p constructs a field that carries a *complex64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Complex64p(key string, val *complex64) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Complex64p(key, val))
	return fields
}

// Float64 constructs a field that carries a float64. The way the
// floating-point value is represented is encoder-dependent, so marshaling is
// necessarily lazy.
func (fields *LogFields) Float64(key string, val float64) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Float64(key, val))
	return fields
}

// Float64p constructs a field that carries a *float64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Float64p(key string, val *float64) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Float64p(key, val))
	return fields
}

// Float32 constructs a field that carries a float32. The way the
// floating-point value is represented is encoder-dependent, so marshaling is
// necessarily lazy.
func (fields *LogFields) Float32(key string, val float32) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Float32(key, val))
	return fields
}

// Float32p constructs a field that carries a *float32. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Float32p(key string, val *float32) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Float32p(key, val))
	return fields
}

// Int constructs a field with the given key and value.
func (fields *LogFields) Int(key string, val int) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Int(key, val))
	return fields
}

// Intp constructs a field that carries a *int. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Intp(key string, val *int) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Intp(key, val))
	return fields
}

// Int64 constructs a field with the given key and value.
func (fields *LogFields) Int64(key string, val int64) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Int64(key, val))
	return fields
}

// Int64p constructs a field that carries a *int64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Int64p(key string, val *int64) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Int64p(key, val))
	return fields
}

// Int32 constructs a field with the given key and value.
func (fields *LogFields) Int32(key string, val int32) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Int32(key, val))
	return fields
}

// Int32p constructs a field that carries a *int32. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Int32p(key string, val *int32) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Int32p(key, val))
	return fields
}

// Int16 constructs a field with the given key and value.
func (fields *LogFields) Int16(key string, val int16) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Int16(key, val))
	return fields
}

// Int16p constructs a field that carries a *int16. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Int16p(key string, val *int16) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Int16p(key, val))
	return fields
}

// Int8 constructs a field with the given key and value.
func (fields *LogFields) Int8(key string, val int8) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Int8(key, val))
	return fields
}

// Int8p constructs a field that carries a *int8. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Int8p(key string, val *int8) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Int8p(key, val))
	return fields
}

// String constructs a field with the given key and value.
func (fields *LogFields) String(key string, val string) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.String(key, val))
	return fields
}

// Stringp constructs a field that carries a *string. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Stringp(key string, val *string) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Stringp(key, val))
	return fields
}

// Uint constructs a field with the given key and value.
func (fields *LogFields) Uint(key string, val uint) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uint(key, val))
	return fields
}

// Uintp constructs a field that carries a *uint. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Uintp(key string, val *uint) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uintp(key, val))
	return fields
}

// Uint64 constructs a field with the given key and value.
func (fields *LogFields) Uint64(key string, val uint64) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uint64(key, val))
	return fields
}

// Uint64p constructs a field that carries a *uint64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Uint64p(key string, val *uint64) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uint64p(key, val))
	return fields
}

// Uint32 constructs a field with the given key and value.
func (fields *LogFields) Uint32(key string, val uint32) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uint32(key, val))
	return fields
}

// Uint32p constructs a field that carries a *uint32. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Uint32p(key string, val *uint32) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uint32p(key, val))
	return fields
}

// Uint16 constructs a field with the given key and value.
func (fields *LogFields) Uint16(key string, val uint16) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uint16(key, val))
	return fields
}

// Uint16p constructs a field that carries a *uint16. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Uint16p(key string, val *uint16) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uint16p(key, val))
	return fields
}

// Uint8 constructs a field with the given key and value.
func (fields *LogFields) Uint8(key string, val uint8) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uint8(key, val))
	return fields
}

// Uint8p constructs a field that carries a *uint8. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Uint8p(key string, val *uint8) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uint8p(key, val))
	return fields
}

// Uintptr constructs a field with the given key and value.
func (fields *LogFields) Uintptr(key string, val uintptr) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uintptr(key, val))
	return fields
}

// Uintptrp constructs a field that carries a *uintptr. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Uintptrp(key string, val *uintptr) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uintptrp(key, val))
	return fields
}

// Reflect constructs a field with the given key and an arbitrary object. It uses
// an encoding-appropriate, reflection-based function to lazily serialize nearly
// any object into the logging context, but it's relatively slow and
// allocation-heavy. Outside tests, Any is always a better choice.
//
// If encoding fails (e.g., trying to serialize a map[int]string to JSON), Reflect
// includes the error message in the final log output.
func (fields *LogFields) Reflect(key string, val interface{}) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Reflect(key, val))
	return fields
}

// Namespace creates a named, isolated scope within the logger's context. All
// subsequent fields will be added to the new namespace.
//
// This helps prevent key collisions when injecting loggers into sub-components
// or third-party libraries.
func (fields *LogFields) Namespace(key string) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Namespace(key))
	return fields
}

// Stringer constructs a field with the given key and the output of the value's
// String method. The Stringer's String method is called lazily.
func (fields *LogFields) Stringer(key string, val fmt.Stringer) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Stringer(key, val))
	return fields
}

// Time constructs a Field with the given key and value. The encoder
// controls how the time is serialized.
func (fields *LogFields) Time(key string, val time.Time) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Time(key, val))
	return fields
}

// Timep constructs a field that carries a *time.Time. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Timep(key string, val *time.Time) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Timep(key, val))
	return fields
}

// Stack constructs a field that stores a stacktrace of the current goroutine
// under provided key. Keep in mind that taking a stacktrace is eager and
// expensive (relatively speaking); this function both makes an allocation and
// takes about two microseconds.
func (fields *LogFields) Stack(key string) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Stack(key))
	return fields
}

// Duration constructs a field with the given key and value. The encoder
// controls how the duration is serialized.
func (fields *LogFields) Duration(key string, val time.Duration) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Duration(key, val))
	return fields
}

// Durationp constructs a field that carries a *time.Duration. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func (fields *LogFields) Durationp(key string, val *time.Duration) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Durationp(key, val))
	return fields
}

// Object constructs a field with the given key and ObjectMarshaler. It
// provides a flexible, but still type-safe and efficient, way to add map- or
// struct-like user-defined types to the logging context. The struct's
// MarshalLogObject method is called lazily.
func (fields *LogFields) Object(key string, val zapcore.ObjectMarshaler) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Object(key, val))
	return fields
}

// Array constructs a field with the given key and ArrayMarshaler. It provides
// a flexible, but still type-safe and efficient, way to add array-like types
// to the logging context. The struct's MarshalLogArray method is called lazily.
func (fields *LogFields) Array(key string, val zapcore.ArrayMarshaler) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Array(key, val))
	return fields
}

// Bools constructs a field that carries a slice of bools.
func (fields *LogFields) Bools(key string, bs []bool) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Bools(key, bs))
	return fields
}

// ByteStrings constructs a field that carries a slice of []byte, each of which
// must be UTF-8 encoded text.
func (fields *LogFields) ByteStrings(key string, bss [][]byte) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.ByteStrings(key, bss))
	return fields
}

// Complex128s constructs a field that carries a slice of complex numbers.
func (fields *LogFields) Complex128s(key string, nums []complex128) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Complex128s(key, nums))
	return fields
}

// Complex64s constructs a field that carries a slice of complex numbers.
func (fields *LogFields) Complex64s(key string, nums []complex64) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Complex64s(key, nums))
	return fields
}

// Durations constructs a field that carries a slice of time.Durations.
func (fields *LogFields) Durations(key string, ds []time.Duration) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Durations(key, ds))
	return fields
}

// Float64s constructs a field that carries a slice of floats.
func (fields *LogFields) Float64s(key string, nums []float64) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Float64s(key, nums))
	return fields
}

// Float32s constructs a field that carries a slice of floats.
func (fields *LogFields) Float32s(key string, nums []float32) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Float32s(key, nums))
	return fields
}

// Ints constructs a field that carries a slice of integers.
func (fields *LogFields) Ints(key string, nums []int) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Ints(key, nums))
	return fields
}

// Int64s constructs a field that carries a slice of integers.
func (fields *LogFields) Int64s(key string, nums []int64) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Int64s(key, nums))
	return fields
}

// Int32s constructs a field that carries a slice of integers.
func (fields *LogFields) Int32s(key string, nums []int32) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Int32s(key, nums))
	return fields
}

// Int16s constructs a field that carries a slice of integers.
func (fields *LogFields) Int16s(key string, nums []int16) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Int16s(key, nums))
	return fields
}

// Int8s constructs a field that carries a slice of integers.
func (fields *LogFields) Int8s(key string, nums []int8) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Int8s(key, nums))
	return fields
}

// Strings constructs a field that carries a slice of strings.
func (fields *LogFields) Strings(key string, ss []string) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Strings(key, ss))
	return fields
}

// Times constructs a field that carries a slice of time.Times.
func (fields *LogFields) Times(key string, ts []time.Time) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Times(key, ts))
	return fields
}

// Uints constructs a field that carries a slice of unsigned integers.
func (fields *LogFields) Uints(key string, nums []uint) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uints(key, nums))
	return fields
}

// Uint64s constructs a field that carries a slice of unsigned integers.
func (fields *LogFields) Uint64s(key string, nums []uint64) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uint64s(key, nums))
	return fields
}

// Uint32s constructs a field that carries a slice of unsigned integers.
func (fields *LogFields) Uint32s(key string, nums []uint32) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uint32s(key, nums))
	return fields
}

// Uint16s constructs a field that carries a slice of unsigned integers.
func (fields *LogFields) Uint16s(key string, nums []uint16) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uint16s(key, nums))
	return fields
}

// Uint8s constructs a field that carries a slice of unsigned integers.
func (fields *LogFields) Uint8s(key string, nums []uint8) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uint8s(key, nums))
	return fields
}

// Uintptrs constructs a field that carries a slice of pointer addresses.
func (fields *LogFields) Uintptrs(key string, us []uintptr) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Uintptrs(key, us))
	return fields
}

// Errors constructs a field that carries a slice of errors.
func (fields *LogFields) Errors(key string, errs []error) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Errors(key, errs))
	return fields
}

// Error is shorthand for the common idiom NamedError("error", err).
func (fields *LogFields) Error(err error) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.Error(err))
	return fields
}

// NamedError constructs a field that lazily stores err.Error() under the
// provided key. Errors which also implement fmt.Formatter (like those produced
// by github.com/pkg/errors) will also have their verbose representation stored
// under key+"Verbose". If passed a nil error, the field is a no-op.
//
// For the common case in which the key is simply "error", the Error function
// is shorter and less repetitive.
func (fields *LogFields) NamedError(key string, err error) *LogFields {
	if fields == nil {
		return fields
	}
	fields.entity.fields = append(fields.entity.fields, zap.NamedError(key, err))
	return fields
}
