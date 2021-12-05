/*
 * Copyright (c) 2018 XLAB d.o.o
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package data

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"imgfe/pkg/gofe/sample"
)

func TestVector(t *testing.T) {
	l := 3
	bound := new(big.Int).Exp(big.NewInt(2), big.NewInt(20), big.NewInt(0))
	sampler := sample.NewUniform(bound)

	x, err := NewRandomVector(l, sampler)
	if err != nil {
		t.Fatalf("Error during random generation: %v", err)
	}

	y, err := NewRandomVector(l, sampler)
	if err != nil {
		t.Fatalf("Error during random generation: %v", err)
	}

	add := x.Add(y)
	mul, err := x.Dot(y)

	if err != nil {
		t.Fatalf("Error during vector multiplication: %v", err)
	}

	modulo := int64(104729)
	mod := x.Mod(big.NewInt(modulo))

	innerProd := big.NewInt(0)
	for i := 0; i < 3; i++ {
		assert.Equal(t, new(big.Int).Add(x[i], y[i]), add[i], "coordinates should sum correctly")
		innerProd = innerProd.Add(innerProd, new(big.Int).Mul(x[i], y[i]))
		assert.Equal(t, new(big.Int).Mod(x[i], big.NewInt(modulo)), mod[i], "coordinates should mod correctly")
	}

	assert.Equal(t, innerProd, mul, "inner product should calculate correctly")

	sampler = sample.NewUniform(big.NewInt(256))
	var key [32]byte
	for i := range key {
		r, _ := sampler.Sample()
		key[i] = byte(r.Int64())
	}

	_, err = NewRandomDetVector(100, big.NewInt(5), &key)
	assert.Equal(t, err, nil)
}

func TestVector_MulAsPolyInRing(t *testing.T) {
	p1 := Vector{big.NewInt(0), big.NewInt(1), big.NewInt(2)}
	p2 := Vector{big.NewInt(2), big.NewInt(1), big.NewInt(0)}
	prod, _ := p1.MulAsPolyInRing(p2)

	assert.Equal(t, prod, Vector{big.NewInt(-2), big.NewInt(2), big.NewInt(5)})
}

func TestVecor_Tensor(t *testing.T) {
	v1 := Vector{big.NewInt(1), big.NewInt(2)}

	v2 := Vector{big.NewInt(1), big.NewInt(2)}

	prodExpected := Vector{big.NewInt(1), big.NewInt(2), big.NewInt(2), big.NewInt(4)}
	prod := v1.Tensor(v2)

	assert.Equal(t, prodExpected, prod, "tensor product of vectors does not work correctly")
}
