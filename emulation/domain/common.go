package domain

import "fmt"

type Tokens uint
type Product uint
type CapacityType string
type Capacity int

func MustGet[K comparable, V any](m map[K]V, k K) V {
	v, ok := m[k]
	if !ok {
		panic(fmt.Errorf("%w: [%T]:[%#v]", ErrNotFound, k, k))
	}
	return v
}
