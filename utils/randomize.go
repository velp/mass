package utils

import (
	"math/rand"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
)

func RandomString(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}

type RandomUint32Range struct {
	collection []uint32
	current    uint32
	total      uint32
}

func NewRandomUint32Range(min, max uint32) RandomUint32Range {
	numRange := RandomUint32Range{
		current: 0,
	}
	// Generate numbers
	numbers := make([]uint32, max-min+1)
	for i := range numbers {
		numbers[i] = min + uint32(i)
	}
	// Shuffle numbers
	r := rand.New(rand.NewSource(time.Now().Unix()))
	numRange.collection = make([]uint32, len(numbers))
	perm := r.Perm(len(numbers))
	for i, randIndex := range perm {
		numRange.collection[i] = numbers[randIndex]
	}
	numRange.total = uint32(len(numbers))
	return numRange
}

func (r *RandomUint32Range) Next() uint32 {
	r.current++
	if r.current == r.total {
		r.current = 0
	}
	return r.collection[r.current]
}
