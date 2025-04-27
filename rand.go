package x

import (
	"math/rand"
)

// letters is the character set used for random string generation
const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandInt(args ...int) int {
	if len(args) == 0 {
		return rand.Int()
	} else if len(args) == 1 {
		return rand.Intn(args[0])
	} else {
		return rand.Intn(args[1]-args[0]+1) + args[0]
	}
}

func RandFloat64(args ...float64) float64 {
	if len(args) == 0 {
		return rand.Float64()
	} else if len(args) == 1 {
		return rand.Float64() * args[0]
	} else {
		return rand.Float64()*(args[1]-args[0]) + args[0]
	}
}

func RandFloat32(args ...float32) float32 {
	if len(args) == 0 {
		return rand.Float32()
	} else if len(args) == 1 {
		return rand.Float32() * args[0]
	} else {
		return rand.Float32()*(args[1]-args[0]) + args[0]
	}
}

func RandInt64(args ...int64) int64 {
	if len(args) == 0 {
		return rand.Int63()
	} else if len(args) == 1 {
		return rand.Int63n(args[0])
	} else {
		return rand.Int63n(args[1]-args[0]+1) + args[0]
	}
}

func RandInt32(args ...int32) int32 {
	if len(args) == 0 {
		return rand.Int31()
	} else if len(args) == 1 {
		return rand.Int31n(args[0])
	} else {
		return rand.Int31n(args[1]-args[0]+1) + args[0]
	}
}

func RandUint64(args ...uint64) uint64 {
	if len(args) == 0 {
		return rand.Uint64()
	} else if len(args) == 1 {
		return rand.Uint64() % args[0]
	} else {
		return rand.Uint64()%(args[1]-args[0]+1) + args[0]
	}
}

func RandUint32(args ...uint32) uint32 {
	if len(args) == 0 {
		return rand.Uint32()
	} else if len(args) == 1 {
		return rand.Uint32() % args[0]
	} else {
		return rand.Uint32()%(args[1]-args[0]+1) + args[0]
	}
}

func RandBool() bool {
	return rand.Intn(2) == 0
}

// RandString generates a random string of the specified length.
// If length is less than 0, it will return an empty string.
func RandString(length int) string {
	if length <= 0 {
		return ""
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
