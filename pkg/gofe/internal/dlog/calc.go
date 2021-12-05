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

package dlog

import (
	"crypto/sha1"
	"fmt"
	"math/big"

	"imgfe/pkg/bn256"
)

// MaxBound limits the interval of values that are checked when
// computing discrete logarithms. It prevents time and memory
// exhaustive computation for practical purposes.
// If Calc is configured to use a boundary value > MaxBound,
// it will be automatically adjusted to MaxBound.
var MaxBound = new(big.Int).Exp(big.NewInt(2), big.NewInt(48), nil)

// Calc represents a discrete logarithm calculator.
type Calc struct{}

// NewCalc generates a new discrete logarithm calculator.
func NewCalc() *Calc {
	return &Calc{}
}

// CalcZp represents a calculator for discrete logarithms
// that operates in the Zp group of integers modulo prime p.
type CalcZp struct {
	p              *big.Int
	bound          *big.Int
	m              *big.Int
	Precomp        map[string]*big.Int
	precompMaxBits int
	neg            bool
}

// InZp builds parameters needed to calculate a discrete
// logarithm in Z_p group.
func (*Calc) InZp(p, order *big.Int) (*CalcZp, error) {
	one := big.NewInt(1)
	var bound *big.Int
	if p == nil {
		return nil, fmt.Errorf("group modulus p cannot be nil")
	}

	if order == nil {
		if !p.ProbablyPrime(20) {
			return nil, fmt.Errorf("group modulus p must be prime")
		}
		bound = new(big.Int).Sub(p, one)
	} else {
		bound = order
	}

	m := new(big.Int).Sqrt(bound)
	m.Add(m, one)

	return &CalcZp{
		p:              p,
		bound:          bound,
		m:              m,
		neg:            false,
		Precomp:        nil,
		precompMaxBits: 0,
	}, nil
}

// WithBound sets a bound for the calculator of the discrete logarithm.
func (c *CalcZp) WithBound(bound *big.Int) *CalcZp {
	if bound != nil && bound.Cmp(MaxBound) < 0 && bound.Sign() > 0 {
		m := new(big.Int).Sqrt(bound)
		m.Add(m, big.NewInt(1))

		return &CalcZp{
			bound:          bound,
			m:              m,
			p:              c.p,
			neg:            c.neg,
			Precomp:        c.Precomp,
			precompMaxBits: c.precompMaxBits,
		}
	}
	return c
}

// WithNeg sets that the result should be searched also among
// negative integers.
func (c *CalcZp) WithNeg() *CalcZp {
	return &CalcZp{
		bound:          c.bound,
		m:              c.m,
		p:              c.p,
		neg:            true,
		Precomp:        c.Precomp,
		precompMaxBits: c.precompMaxBits,
	}
}

// Precompute precomputes small steps for the discrete logarithm
// search. The resulting precomputation table is of size 2^maxBits.
func (c *CalcZp) Precompute(maxBits int, g *big.Int) error {
	if maxBits < 2 {
		return fmt.Errorf("maxBits should be at least 1")
	}

	one := big.NewInt(1)
	sh := sha1.New()
	// big.Int cannot be a key, thus we use a stringified bytes representation of the integer
	T := make(map[string]*big.Int)
	x := big.NewInt(1)

	for i := big.NewInt(0); i.BitLen() <= maxBits; i.Add(i, one) {
		sh.Write([]byte(x.String()))
		T[string(sh.Sum(nil)[:10])] = new(big.Int).Set(i)
		sh.Reset()
		x = new(big.Int).Mod(new(big.Int).Mul(x, g), c.p)
	}

	c.Precomp = T
	c.precompMaxBits = maxBits
	return nil
}

// BabyStepGiantStep uses the baby-step giant-step method to
// compute the discrete logarithm in the Zp group. If c.neg is
// set to true it searches for the answer within [-bound, bound].
// It does so by running two goroutines, one for negative
// answers and one for positive. If c.neg is set to false
// only one goroutine is started, searching for the answer
// within [0, bound].
func (c *CalcZp) BabyStepGiantStep(h, g *big.Int) (*big.Int, error) {
	// create goroutines calculating positive and possibly negative
	// result if c.neg is set to true
	retChan := make(chan *big.Int)
	errChan := make(chan error)
	if c.precompMaxBits == 0 {
		go c.runBabyStepGiantStepIterative(h, g, retChan, errChan)
	} else {
		go c.runBabyStepGiantStep(h, g, retChan, errChan)
	}
	if c.neg {
		hInv := new(big.Int).ModInverse(h, c.p)
		if c.precompMaxBits == 0 {
			go c.runBabyStepGiantStepIterative(hInv, g, retChan, errChan)
		} else {
			go c.runBabyStepGiantStep(hInv, g, retChan, errChan)
		}
	}

	// catch a value when the first routine finishes
	ret := <-retChan
	err := <-errChan
	// prevent the situation when one routine exhausted all possibilities
	// before the second found the solution
	if c.neg && err != nil {
		ret = <-retChan
		err = <-errChan
	}
	// if both routines give an error, return an error
	if err != nil {
		return nil, err
	}
	// based on ret decide which routine gave the answer, thus if
	// answer is negative
	if c.neg && h.Cmp(new(big.Int).Exp(g, ret, c.p)) != 0 {
		ret.Neg(ret)
	}

	return ret, nil
}

// runBabyStepGiantStep implements the baby-step giant-step method to
// compute the discrete logarithm in the Zp group. It is meant to be run
// as a goroutine.
//
// The function searches for x, where h = g^x mod p. If the solution was not found
// within the provided bound, it returns an error.
func (c *CalcZp) runBabyStepGiantStep(h, g *big.Int, retChan chan *big.Int, errChan chan error) {
	one := big.NewInt(1)

	m := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(c.precompMaxBits)), nil)
	// g^-m
	z := new(big.Int).ModInverse(g, c.p)
	z.Exp(z, m, c.p)
	x := new(big.Int).Set(h)
	sh := sha1.New()
	for i := big.NewInt(0); i.Cmp(m) < 0; i.Add(i, one) {
		sh.Write([]byte(x.String()))
		e, ok := c.Precomp[string(sh.Sum(nil)[:10])]
		sh.Reset()
		if ok {
			retChan <- new(big.Int).Add(new(big.Int).Mul(i, m), e)
			errChan <- nil
			return
		}
		x = new(big.Int).Mod(new(big.Int).Mul(x, z), c.p)
	}
	retChan <- nil
	errChan <- fmt.Errorf("failed to find the discrete logarithm within bound " + new(big.Int).Mul(m, m).String())
}

// runBabyStepGiantStepIterative implements the baby-step giant-step method to
// compute the discrete logarithm in the Zp group. It is meant to be run
// as a goroutine.
//
// The function searches for x, where h = g^x mod p. If the solution was not found
// within the provided bound, it returns an error. In contrast to the usual
// implementation of the method, this one proceeds iteratively, meaning that
// smaller the solution is, faster the algorithm finishes.
func (c *CalcZp) runBabyStepGiantStepIterative(h, g *big.Int, retChan chan *big.Int, errChan chan error) {
	one := big.NewInt(1)
	two := big.NewInt(2)

	// big.Int cannot be a key, thus we use a stringified bytes representation of the integer
	T := make(map[string]*big.Int)
	// prepare values for the loop
	x := big.NewInt(1)
	y := new(big.Int).Set(h)
	z := new(big.Int).ModInverse(g, c.p)
	z.Exp(z, two, c.p)

	bits := int64(c.m.BitLen())

	T[string(x.Bytes())] = big.NewInt(0)
	x.Mod(x.Mul(x, g), c.p)
	j := big.NewInt(0)
	giantStep := new(big.Int)
	bound := new(big.Int)
	for i := int64(0); i < bits; i++ {
		// iteratively increasing giant step up to maximal value c.m
		giantStep.Exp(two, big.NewInt(i+1), nil)
		if giantStep.Cmp(c.m) > 0 {
			giantStep.Set(c.m)
			z.ModInverse(g, c.p)
			z.Exp(z, c.m, c.p)
		}
		// for the selected giant step, add all the needed small steps
		for k := new(big.Int).Exp(two, big.NewInt(i), nil); k.Cmp(giantStep) < 0; k.Add(k, one) {
			T[string(x.Bytes())] = new(big.Int).Set(k)
			x = x.Mod(x.Mul(x, g), c.p)
		}
		// make giant steps and search for the solution
		bound.Exp(two, big.NewInt(2*(i+1)), nil)
		for ; j.Cmp(bound) < 0; j.Add(j, giantStep) {
			if e, ok := T[string(y.Bytes())]; ok {
				retChan <- new(big.Int).Add(j, e)
				errChan <- nil
				return
			}
			y.Mod(y.Mul(y, z), c.p)
		}
		z.Mul(z, z)
		z.Mod(z, c.p)
	}

	retChan <- nil
	errChan <- fmt.Errorf("failed to find the discrete logarithm within bound")
}

// CalcBN256 represents a calculator for discrete logarithms
// that operates in the BN256 group.
type CalcBN256 struct {
	bound          *big.Int
	m              *big.Int
	Precomp        map[string]*big.Int
	precompMaxBits int
	neg            bool
}

// InBN256 builds parameters needed to calculate a discrete
// logarithm in a pairing BN256 group.
func (*Calc) InBN256() *CalcBN256 {
	m := new(big.Int).Sqrt(MaxBound)
	m.Add(m, big.NewInt(1))
	return &CalcBN256{
		bound: MaxBound,
		m:     m,
		neg:   false,
	}
}

// WithBound sets a bound for the calculator of the discrete logarithm.
func (c *CalcBN256) WithBound(bound *big.Int) *CalcBN256 {
	if bound != nil && bound.Cmp(MaxBound) < 0 {
		m := new(big.Int).Sqrt(bound)
		m.Add(m, big.NewInt(1))

		return &CalcBN256{
			bound:          bound,
			m:              m,
			Precomp:        c.Precomp,
			precompMaxBits: c.precompMaxBits,
			neg:            c.neg,
		}
	}
	return c
}

// WithNeg sets that the result should be searched also among
// negative integers.
func (c *CalcBN256) WithNeg() *CalcBN256 {
	return &CalcBN256{
		bound:          c.bound,
		m:              c.m,
		Precomp:        c.Precomp,
		precompMaxBits: c.precompMaxBits,
		neg:            true,
	}
}

// Precompute precomputes small steps for the discrete logarithm
// search. The resulting precomputation table is of size 2^maxBits.
func (c *CalcBN256) Precompute(maxBits int) error {
	if maxBits < 2 {
		return fmt.Errorf("maxBits should be at least 1")
	}
	g := new(bn256.GT).ScalarBaseMult(big.NewInt(1))

	one := big.NewInt(1)
	sh := sha1.New()
	// big.Int cannot be a key, thus we use a stringified bytes representation of the integer
	T := make(map[string]*big.Int)
	x := bn256.GetGTOne()

	for i := big.NewInt(0); i.BitLen() <= maxBits; i.Add(i, one) {
		sh.Write([]byte(x.String()))
		T[string(sh.Sum(nil)[:10])] = new(big.Int).Set(i)
		sh.Reset()
		x = new(bn256.GT).Add(x, g)
	}

	c.Precomp = T
	c.precompMaxBits = maxBits
	return nil
}

// BabyStepGiantStepStd implements the baby-step giant-step method to
// compute the discrete logarithm in the BN256.GT group.
//
// It searches for a solution <= bound. If bound argument is nil,
// the bound is automatically set to the hard coded MaxBound.
//
// The function returns x, where h = g^x in BN256.GT group where operations
// are written as multiplications. If the solution was not found
// within the provided bound, it returns an error.
func (c *CalcBN256) BabyStepGiantStepStd(h, g *bn256.GT) (*big.Int, error) {
	one := big.NewInt(1)

	// first part of the method can be reused so we
	// Precompute it and save it for later

	if c.Precomp == nil {
		maxbits := c.m.BitLen() + 1
		_ = c.Precompute(maxbits)
	}

	// z = g^-m
	gm := new(bn256.GT).ScalarMult(g, c.m)
	z := new(bn256.GT).Neg(gm)
	x := new(bn256.GT).Set(h)
	for i := big.NewInt(0); i.Cmp(c.m) < 0; i.Add(i, one) {
		if e, ok := c.Precomp[x.String()]; ok {
			return new(big.Int).Add(new(big.Int).Mul(i, c.m), e), nil
		}
		x.Add(x, z)
	}

	return nil, fmt.Errorf("failed to find discrete logarithm within bound")
}

// BabyStepGiantStep uses the baby-step giant-step method to
// compute the discrete logarithm in the BN256.GT group. If c.neg is
// set to true it searches for the answer within [-bound, bound].
// It does so by running two goroutines, one for negative
// answers and one for positive. If c.neg is set to false
// only one goroutine is started, searching for the answer
// within [0, bound].
func (c *CalcBN256) BabyStepGiantStep(h, g *bn256.GT) (*big.Int, error) {
	// create goroutines calculating positive and possibly negative
	// result if c.neg is set to true
	retChan := make(chan *big.Int, 2)
	errChan := make(chan error, 2)
	quit := make(chan bool, 2)
	go c.runBabyStepGiantStepIterative(h, g, retChan, errChan, quit)
	if c.neg {
		hInv := new(bn256.GT).Neg(h)
		go c.runBabyStepGiantStepIterative(hInv, g, retChan, errChan, quit)
	}

	// catch a value when the first routine finishes
	ret := <-retChan
	err := <-errChan

	// prevent the situation when one routine exhausted all possibilities
	// before the second found the solution
	if c.neg && err != nil {
		ret = <-retChan
		err = <-errChan
	}
	if c.neg {
		quit <- true
	}
	// if both routines give an error, return an error
	if err != nil {
		return nil, err
	}
	// based on ret decide which routine gave the answer, thus if
	// answer is negative
	if c.neg && h.String() != new(bn256.GT).ScalarMult(g, ret).String() {
		ret.Neg(ret)
	}

	return ret, nil
}

// runBabyStepGiantStepIterative implements the baby-step giant-step method to
// compute the discrete logarithm in the BN256.GT group. It is meant to be run
// as a goroutine.
//
// The function searches for x, where h = g^x in BN256.GT group where operations
// are written as multiplications. If the solution was not found
// within the provided bound, it returns an error. In contrast to the usual
// implementation of the method, this one proceeds iteratively, meaning that
// smaller the solution is, faster the algorithm finishes.
func (c *CalcBN256) runBabyStepGiantStepIterative(h, g *bn256.GT, retChan chan *big.Int, errChan chan error, quit chan bool) {
	one := big.NewInt(1)
	two := big.NewInt(2)

	var startBits int
	if c.Precomp == nil {
		_ = c.Precompute(2)
	}

	startBits = c.precompMaxBits

	// prepare values for the loop
	y := new(bn256.GT).Set(h)
	j := big.NewInt(0)

	// define first giant step
	giantStep := new(big.Int)
	giantStep.Exp(big.NewInt(2), big.NewInt(int64(startBits)), nil)
	z := new(bn256.GT).Neg(g)
	z.ScalarMult(z, giantStep)

	bound := new(big.Int).Exp(two, big.NewInt(2*int64(startBits)), nil)
	sh := sha1.New()
	for ; j.Cmp(bound) < 0; j.Add(j, giantStep) {
		select {
		case <-quit:
			return
		default:
			sh.Write([]byte(y.String()))
			e, ok := c.Precomp[string(sh.Sum(nil)[:10])]
			sh.Reset()
			if ok {
				retChan <- new(big.Int).Add(j, e)
				errChan <- nil
				return
			}
			y.Add(y, z)
		}
	}
	z.Add(z, z)
	x := new(bn256.GT).ScalarMult(g, new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(startBits)), nil))

	T := make(map[string]*big.Int)
	for k, v := range c.Precomp {
		T[k] = v
	}

	bits := int64(c.m.BitLen())
	for i := int64(startBits); i < bits; i++ {
		select {
		case <-quit:
			return
		default:
			// iteratively increasing giant step up to maximal value c.m
			giantStep.Exp(two, big.NewInt(i+1), nil)
			if giantStep.Cmp(c.m) > 0 {
				giantStep.Set(c.m)
				z.Neg(g)
				z.ScalarMult(z, c.m)
			}
			// for the selected giant step, add all the needed small steps
			for k := new(big.Int).Exp(two, big.NewInt(i), nil); k.Cmp(giantStep) < 0; k.Add(k, one) {
				select {
				case <-quit:
					return
				default:
					sh.Write([]byte(x.String()))
					T[string(sh.Sum(nil)[:10])] = new(big.Int).Set(k)
					sh.Reset()
					x = new(bn256.GT).Add(x, g)
				}
			}
			// make giant steps and search for the solution
			bound.Exp(giantStep, two, nil)
			for ; j.Cmp(bound) < 0; j.Add(j, giantStep) {
				select {
				case <-quit:
					return
				default:
					sh.Write([]byte(y.String()))
					e, ok := T[string(sh.Sum(nil)[:10])]
					sh.Reset()
					if ok {
						retChan <- new(big.Int).Add(j, e)
						errChan <- nil
						return
					}
					y.Add(y, z)
				}
			}
			z.Add(z, z)
		}
	}
	retChan <- nil
	errChan <- fmt.Errorf("failed to find the discrete logarithm within bound")
}
