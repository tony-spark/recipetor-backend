package random

import (
	"math/rand"
	"time"
)

var (
	letters          = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	lowercaseLetters = []rune("abcdefghijklmnopqrstuvwxyz")
	hexSymbols       = []rune("0123456789abcdef")
)

type Generator interface {
	RandomString(length int) string
	RandomEmail() string
	RandomFloat(max float64) float64
	RandomObjectID() string
}

type generator struct {
	r *rand.Rand
}

func NewRandomGenerator() Generator {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return generator{
		r: r,
	}
}

func (g generator) randomString(rs []rune, n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = rs[g.r.Intn(len(rs))]
	}
	return string(b)
}

func (g generator) RandomString(length int) string {
	return g.randomString(letters, length)
}

func (g generator) RandomEmail() string {
	return g.randomString(lowercaseLetters, 8) + "@test.com"
}

func (g generator) RandomFloat(max float64) float64 {
	return g.r.Float64() * max
}

func (g generator) RandomObjectID() string {
	return g.randomString(hexSymbols, 24)
}
