package x

import "golang.org/x/exp/constraints"

func Max[T constraints.Ordered](args ...T) T {
	if len(args) == 0 {
		panic("Max: no arguments")
	}

	max := args[0]

	for _, arg := range args[1:] {
		if arg > max {
			max = arg
		}
	}
	return max
}

func Min[T constraints.Ordered](args ...T) T {
	if len(args) == 0 {
		panic("Min: no arguments")
	}

	min := args[0]

	for _, arg := range args[1:] {
		if arg < min {
			min = arg
		}
	}
	return min
}

type Nullable interface {
	~*any | []any | map[any]any | chan any | func()
}

// FirstNotNil returns the first non-nil value from the arguments
func FirstNotNil(args ...any) any {
	for _, arg := range args {
		if arg != nil {
			return arg
		}
	}

	return nil
}

func LastNotNil[T Nullable](args ...T) T {
	for i := len(args) - 1; i >= 0; i-- {
		if args[i] != nil {
			return args[i]
		}
	}

	var zero T
	return zero
}
