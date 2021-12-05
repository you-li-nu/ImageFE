package decryptor

import (
	"fmt"
	"imgfe/src/ddh"
	"imgfe/src/encryptor"
	"imgfe/src/tensor"
	"math/big"

	"imgfe/pkg/gofe/data"
	"imgfe/pkg/gofe/innerprod/simple"

	"github.com/cheggaaa/pb"
)

type DecryptScheme struct {
	params    simple.DDHParams
	ymasks    [][]*big.Int
	funcKeys  []*big.Int
	decryptor *simple.DDH
}

func NewDecryptScheme(params simple.DDHParams, ymasks [][]*big.Int, fk []*big.Int) (DecryptScheme, error) {
	return DecryptScheme{
		params:    params,
		ymasks:    ymasks,
		funcKeys:  fk,
		decryptor: nil,
	}, nil
}

func NewDecryptSchemePrecomp(params simple.DDHParams, ymasks [][]*big.Int, fk []*big.Int, verbose int) (DecryptScheme, error) {
	decryptor := simple.NewDDHFromParams(&params)
	precomputedBits := 10
	if verbose >= 1 {
		fmt.Println("***Precomputation started.")
		fmt.Println("***Precomputed bits:", precomputedBits)
	}
	err := decryptor.PrecomputeDlog(precomputedBits)
	if verbose >= 1 {
		fmt.Println("***Precomputation finished.")
	}
	if err != nil {
		return DecryptScheme{}, fmt.Errorf("error during precomputation: %v", err)
	}
	return DecryptScheme{
		params:    params,
		ymasks:    ymasks,
		funcKeys:  fk,
		decryptor: decryptor,
	}, nil
}

func (ds *DecryptScheme) DecryptSingle(et encryptor.EncryptTensor, key_idx int) (tensor.Tensor, error) {
	bar := pb.StartNew(et.H)
	new_t := tensor.GetZeroTensor(et.W, et.H)
	for i := 0; i < new_t.H; i++ {
		for j := 0; j < new_t.W; j++ {
			//fmt.Println(i, j) // youl
			idx := new_t.GetIdx(i, j)

			r, err := ds.DecryptDDH(et.R[idx], key_idx)
			if err != nil {
				return tensor.Tensor{}, err
			}
			new_t.R[idx] = r

			g, err := ds.DecryptDDH(et.G[idx], key_idx)
			if err != nil {
				return tensor.Tensor{}, err
			}
			new_t.G[idx] = g

			b, err := ds.DecryptDDH(et.B[idx], key_idx)
			if err != nil {
				return tensor.Tensor{}, err
			}
			new_t.B[idx] = b

			a, err := ds.DecryptDDH(et.A[idx], key_idx)
			if err != nil {
				return tensor.Tensor{}, err
			}
			new_t.A[idx] = a
		}
		bar.Increment()
	}
	bar.Finish()
	return new_t, nil
}

func (ds *DecryptScheme) DecryptMultiple(et encryptor.EncryptTensor, ec encryptor.EncryptCredentials, verbose int) (tensor.Tensor, error) {
	var tensors []tensor.Tensor
	for kernel_idx := 0; kernel_idx <= ec.Num_kernels-1; kernel_idx++ {
		if verbose >= 1 {
			fmt.Println("***Decrypt with kernel", kernel_idx)
		}
		new_tensor, err := ds.DecryptSingle(et, kernel_idx)
		if err != nil {
			return tensor.Tensor{}, err
		}
		tensors = append(tensors, new_tensor)
	}
	return tensor.MergeMultipleTensors(tensors, ec.Kernel_size)
}

func (scheme *DecryptScheme) DecryptDDH(ciphertext data.Vector, key_idx int) (uint32, error) {
	xy, err := scheme.decryptor.Decrypt(ciphertext, scheme.funcKeys[key_idx], scheme.ymasks[key_idx])
	if err != nil {
		return 0, fmt.Errorf("error during decryption: %v", err)
	}
	return uint32(xy.Int64()), nil
}

func GetDDHParamsFromString(params ddh.DDHParamsString) simple.DDHParams {
	var ret_params simple.DDHParams
	ret_params.L = params.L
	bound := new(big.Int)
	bound.SetString(params.Bound, 10)
	ret_params.Bound = bound
	g := new(big.Int)
	g.SetString(params.G, 10)
	ret_params.G = g
	p := new(big.Int)
	p.SetString(params.P, 10)
	ret_params.P = p
	q := new(big.Int)
	q.SetString(params.Q, 10)
	ret_params.Q = q
	return ret_params
}

func GetYMasksFromString(ymasks [][]string) [][]*big.Int {
	var ret_ymasks [][]*big.Int
	for _, vi := range ymasks {
		var tmp_ymasks []*big.Int
		for _, vj := range vi {
			y := new(big.Int)
			y.SetString(vj, 10)
			tmp_ymasks = append(tmp_ymasks, y)
		}
		ret_ymasks = append(ret_ymasks, tmp_ymasks)
	}
	return ret_ymasks
}

func GetFunctionKeysFromString(func_keys []string) []*big.Int {
	var ret_keys []*big.Int
	for _, v := range func_keys {
		fk := new(big.Int)
		fk.SetString(v, 10)
		ret_keys = append(ret_keys, fk)
	}
	return ret_keys
}

func GetEncryptTensorFromString(et encryptor.EncryptTensorString) encryptor.EncryptTensor {
	var ret_tensor encryptor.EncryptTensor
	ret_tensor.W = et.W
	ret_tensor.H = et.H
	for _, vi := range et.R {
		var tmp_slice data.Vector
		for _, vj := range vi {
			d := new(big.Int)
			d.SetString(vj, 10)
			tmp_slice = append(tmp_slice, d)
		}
		ret_tensor.R = append(ret_tensor.R, tmp_slice)
	}
	for _, vi := range et.G {
		var tmp_slice data.Vector
		for _, vj := range vi {
			d := new(big.Int)
			d.SetString(vj, 10)
			tmp_slice = append(tmp_slice, d)
		}
		ret_tensor.G = append(ret_tensor.G, tmp_slice)
	}
	for _, vi := range et.B {
		var tmp_slice data.Vector
		for _, vj := range vi {
			d := new(big.Int)
			d.SetString(vj, 10)
			tmp_slice = append(tmp_slice, d)
		}
		ret_tensor.B = append(ret_tensor.B, tmp_slice)
	}
	for _, vi := range et.A {
		var tmp_slice data.Vector
		for _, vj := range vi {
			d := new(big.Int)
			d.SetString(vj, 10)
			tmp_slice = append(tmp_slice, d)
		}
		ret_tensor.A = append(ret_tensor.A, tmp_slice)
	}
	return ret_tensor
}
