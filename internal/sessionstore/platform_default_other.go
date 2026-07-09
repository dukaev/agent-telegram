//go:build !darwin || !cgo

package sessionstore

func platformDefaultProvider() string { return MemoryProvider }
