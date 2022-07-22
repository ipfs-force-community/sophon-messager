package mtypes

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeFromGo(t *testing.T) {
	a := big.NewInt(1000)
	res := SafeFromGo(a)
	assert.Equal(t, NewInt(1000), res)

	var nilVal *big.Int
	res = SafeFromGo(nilVal)
	assert.Equal(t, NewInt(0), res)
}
