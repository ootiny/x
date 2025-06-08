package x

import (
	"encoding/base64"
	"math/rand"
	"strings"

	"github.com/google/uuid"
)

func UUID() string {
	raw := uuid.New()
	return base64.RawURLEncoding.EncodeToString(raw[:])
}

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

func RandString(length int) string {
	if length <= 0 {
		return ""
	}

	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const letterIdxBits = 6
	const letterIdxMask = 1<<letterIdxBits - 1
	const letterIdxMax = 63 / letterIdxBits

	var sb strings.Builder
	sb.Grow(length)

	for i, cache, remain := 0, rand.Int63(), letterIdxMax; i < length; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}

		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i++
		}

		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func RandFileName(length int) string {
	// 随机生成一个文件名，文件名由数字和小写字母组成，长度为length
	// 文件名不能以数字开头

	if length <= 0 {
		return ""
	}

	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	const letterIdxBits = 6
	const letterIdxMask = 1<<letterIdxBits - 1
	const letterIdxMax = 63 / letterIdxBits

	var sb strings.Builder
	sb.Grow(length)

	for i, cache, remain := 0, rand.Int63(), letterIdxMax; i < length; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}

		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i++
		}

		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}
