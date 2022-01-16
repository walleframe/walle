package process

//go:generate mockgen -source atomic_num.go -destination ../testpkg/mock_process/atomic_num.go

type AtomicNumber interface {
	// Load atomically loads the wrapped value.
	Load() int64
	// Add atomically adds to the wrapped int64 and returns the new value.
	Add(n int64) int64
	// Sub atomically subtracts from the wrapped int64 and returns the new value.
	Sub(n int64) int64
	// Inc atomically increments the wrapped int64 and returns the new value.
	Inc() int64
	// Dec atomically decrements the wrapped int64 and returns the new value.
	Dec() int64
	// Store atomically stores the passed value.
	Store(n int64)
}
