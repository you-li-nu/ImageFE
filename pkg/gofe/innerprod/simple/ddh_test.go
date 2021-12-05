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

package simple_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"imgfe/pkg/gofe/data"
	"imgfe/pkg/gofe/innerprod/simple"
	"imgfe/pkg/gofe/sample"
)

type ddhTestParam struct {
	name          string
	modulusLength int
	precomputed   bool
}

func testSimpleDDHFromParam(t *testing.T, param ddhTestParam) {
	l := 3
	bound := new(big.Int).Exp(big.NewInt(2), big.NewInt(10), nil)
	sampler := sample.NewUniformRange(new(big.Int).Add(new(big.Int).Neg(bound), big.NewInt(1)), bound)

	var simpleDDH *simple.DDH
	var err error
	if param.precomputed {
		simpleDDH, err = simple.NewDDHPrecomp(l, param.modulusLength, bound)
	} else {
		simpleDDH, err = simple.NewDDH(l, param.modulusLength, bound)
	}
	if err != nil {
		t.Fatalf("Error during simple inner product creation: %v", err)
	}

	masterSecKey, masterPubKey, err := simpleDDH.GenerateMasterKeys()

	if err != nil {
		t.Fatalf("Error during master key generation: %v", err)
	}

	y, err := data.NewRandomVector(l, sampler)

	if err != nil {
		t.Fatalf("Error during random generation: %v", err)
	}

	funcKey, err := simpleDDH.DeriveKey(masterSecKey, y)

	if err != nil {
		t.Fatalf("Error during key derivation: %v", err)
	}

	x, err := data.NewRandomVector(l, sampler)

	if err != nil {
		t.Fatalf("Error during random generation: %v", err)
	}

	// simulate the instantiation of encryptor (which should be given masterPubKey)
	encryptor := simple.NewDDHFromParams(simpleDDH.Params)
	xyCheck, err := x.Dot(y)

	if err != nil {
		t.Fatalf("Error during inner product calculation")
	}
	ciphertext, err := encryptor.Encrypt(x, masterPubKey)

	if err != nil {
		t.Fatalf("Error during encryption: %v", err)
	}

	decryptor := simple.NewDDHFromParams(simpleDDH.Params)

	// if decryption will be done for multiple ciphertexts pre-computation of
	// dlog values can speed thing up
	// the following precomputes a table of size O(2^precomputedBits)
	// which can be used if the solution is in range [-2^(2*precomputed Bits), 2^(2*precomputed Bits)]
	precomputedBits := 12
	err = decryptor.PrecomputeDlog(precomputedBits)
	if err != nil {
		t.Fatalf("Error during precomputation: %v", err)
	}
	xy, err := decryptor.Decrypt(ciphertext, funcKey, y)

	if err != nil {
		t.Fatalf("Error during decryption: %v", err)
	}

	assert.Equal(t, xy, xyCheck, "Original and decrypted values should match")
}

func TestSimple_DDH(t *testing.T) {
	params := []ddhTestParam{{name: "random", modulusLength: 512, precomputed: false},
		{name: "precomputed", modulusLength: 2048, precomputed: true}}

	for _, param := range params {
		t.Run(param.name, func(t *testing.T) {
			testSimpleDDHFromParam(t, param)
		})
	}
}
