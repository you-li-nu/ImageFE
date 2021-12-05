package encryptor

import (
	"fmt"
	"image"
	"imgfe/src/ddh"
	"imgfe/src/image_decoder"
	"imgfe/src/tensor"
	"math/big"

	"imgfe/pkg/gofe/data"
	"imgfe/pkg/gofe/innerprod/simple"

	"github.com/cheggaaa/pb"
)

type OriginalTensor struct {
	Img    image.Image
	Tensor tensor.Tensor
}

type EncryptTensor struct {
	W, H       int
	R, G, B, A []data.Vector // one ciphertext per grid
}

type EncryptScheme struct {
	scheme *ddh.DDH
}

func NewOriginalTensor(filename string) (*OriginalTensor, error) {
	img, err := image_decoder.GetImg(filename)
	if err != nil {
		return nil, err
	}
	if err == nil && (img.Bounds().Max.X <= 0 || img.Bounds().Max.Y <= 0) {
		return nil, fmt.Errorf("empty image")
	}
	tensor := tensor.NewTensor(img)
	//_, _, _, a := img.At(0, 0).RGBA()
	return &OriginalTensor{
		Img:    img,
		Tensor: tensor,
	}, nil
}

func NewEncryptTensor(w, h int) EncryptTensor {
	var R []data.Vector
	var G []data.Vector
	var B []data.Vector
	var A []data.Vector
	return EncryptTensor{
		W: w,
		H: h,
		R: R,
		G: G,
		B: B,
		A: A,
	}
}

func NewEncryptScheme(l int, modulus_length int) (EncryptScheme, error) {
	d, err := ddh.NewDDH(l, modulus_length, 4096)
	if err != nil {
		return EncryptScheme{}, err
	}
	return EncryptScheme{
		scheme: d,
	}, nil
}

func NewEncryptSchemePrecomp(l int, modulus_length int) (EncryptScheme, error) {
	d, err := ddh.NewDDHPrecomp(l, modulus_length, 4096)
	if err != nil {
		return EncryptScheme{}, err
	}
	return EncryptScheme{
		scheme: d,
	}, nil
}

func (o *OriginalTensor) PaddingCenter(offset int, filler uint32) tensor.Tensor {
	return o.Tensor.PaddingCenter(offset, filler)
}

func (o *OriginalTensor) PaddingTopLeft(dist int, filler uint32) tensor.Tensor {
	return o.Tensor.PaddingTopLeft(dist, filler)
}

func (o *OriginalTensor) CroppingTopLeft(dist int, filler uint32) tensor.Tensor {
	return o.Tensor.CroppingTopLeft(dist, filler)
}

func (es *EncryptScheme) DeriveFunctionKey(mask []uint32) {
	es.scheme.FunctionDDH(mask)
}

func (es *EncryptScheme) EncryptCenter(dist int, t tensor.Tensor) (EncryptTensor, error) {
	bar := pb.StartNew(t.H)
	et := NewEncryptTensor(t.W-2*dist, t.H-2*dist)
	for i := 0; i < t.H; i++ {
		for j := 0; j < t.W; j++ {
			idx := t.GetIdx(i, j)
			if !t.IsNonEdgeCenter(dist, idx) {
				continue
			}
			//fmt.Println(i, j) // youl
			sub, err := t.SubTensorCenter(dist, i, j)
			if err != nil {
				return EncryptTensor{}, err
			}

			r, err := es.scheme.EncryptDDH(sub.R)
			if err != nil {
				return EncryptTensor{}, err
			}
			et.R = append(et.R, r)

			g, err := es.scheme.EncryptDDH(sub.G)
			if err != nil {
				return EncryptTensor{}, err
			}
			et.G = append(et.G, g)

			b, err := es.scheme.EncryptDDH(sub.B)
			if err != nil {
				return EncryptTensor{}, err
			}
			et.B = append(et.B, b)

			a, err := es.scheme.EncryptDDH(sub.A)
			if err != nil {
				return EncryptTensor{}, err
			}
			et.A = append(et.A, a)
		}
		bar.Increment()
	}
	bar.Finish()
	return et, nil
}

func (es *EncryptScheme) EncryptTopLeft(dist int, t tensor.Tensor) (EncryptTensor, error) {
	bar := pb.StartNew(t.H)
	et := NewEncryptTensor(t.W/dist, t.H/dist)
	for i := 0; i < t.H; i++ {
		for j := 0; j < t.W; j++ {
			idx := t.GetIdx(i, j)
			if !t.IsTopLeft(dist, idx) {
				continue
			}
			//fmt.Println(i, j) // youl
			sub, err := t.SubTensorTopLeft(dist, i, j)
			if err != nil {
				return EncryptTensor{}, err
			}

			r, err := es.scheme.EncryptDDH(sub.R)
			if err != nil {
				return EncryptTensor{}, err
			}
			et.R = append(et.R, r)

			g, err := es.scheme.EncryptDDH(sub.G)
			if err != nil {
				return EncryptTensor{}, err
			}
			et.G = append(et.G, g)

			b, err := es.scheme.EncryptDDH(sub.B)
			if err != nil {
				return EncryptTensor{}, err
			}
			et.B = append(et.B, b)

			a, err := es.scheme.EncryptDDH(sub.A)
			if err != nil {
				return EncryptTensor{}, err
			}
			et.A = append(et.A, a)
		}
		bar.Increment()
	}
	bar.Finish()
	return et, nil
}

func (es *EncryptScheme) Get_Params_Y_FK() (simple.DDHParams, [][]*big.Int, []*big.Int) {
	return es.scheme.GetDDHParams(), es.scheme.GetYMasks(), es.scheme.GetFunctionKeys()
}

func (es *EncryptScheme) GetDDHParamsString() ddh.DDHParamsString {
	var ret_params ddh.DDHParamsString
	params := es.scheme.GetDDHParams()
	ret_params.L = params.L
	ret_params.Bound = params.Bound.String()
	ret_params.G = params.G.String()
	ret_params.P = params.P.String()
	ret_params.Q = params.Q.String()
	return ret_params
}

type EncryptCredentials struct {
	Num_kernels int
	Kernel_size int
	Kernel_type string
	Key_length  int
	Y_mask      [][]string
	Func_key    []string
}

func (es *EncryptScheme) GetEncryptCredential(kernel_size int, kernel_type string, key_length int, key_idx int) EncryptCredentials {
	var ret_credentials EncryptCredentials
	ret_credentials.Num_kernels = 1
	ret_credentials.Kernel_size = kernel_size
	ret_credentials.Kernel_type = kernel_type
	ret_credentials.Key_length = key_length
	ret_credentials.Y_mask = [][]string{es.GetYMasksString()[key_idx]}
	ret_credentials.Func_key = []string{es.GetFunctionKeysString()[key_idx]}
	return ret_credentials
}

func (es *EncryptScheme) GetLastEncryptCredentials(kernel_size int, kernel_type string, key_length int, num_credentials int) EncryptCredentials {
	var ret_credentials EncryptCredentials
	ret_credentials.Num_kernels = num_credentials
	ret_credentials.Kernel_size = kernel_size
	ret_credentials.Kernel_type = kernel_type
	ret_credentials.Key_length = key_length
	ret_credentials.Y_mask = es.GetYMasksString()[len(es.GetYMasksString())-num_credentials:]
	ret_credentials.Func_key = es.GetFunctionKeysString()[len(es.GetYMasksString())-num_credentials:]
	return ret_credentials
}

func (es *EncryptScheme) GetYMasksString() [][]string {
	var ret_ymasks [][]string
	ymasks := es.scheme.GetYMasks()
	for _, vi := range ymasks {
		var tmp_ymasks []string
		for _, vj := range vi {
			tmp_ymasks = append(tmp_ymasks, vj.String())
		}
		ret_ymasks = append(ret_ymasks, tmp_ymasks)
	}
	return ret_ymasks
}

func (es *EncryptScheme) GetFunctionKeysString() []string {
	var ret_keys []string
	func_keys := es.scheme.GetFunctionKeys()
	for _, v := range func_keys {
		ret_keys = append(ret_keys, v.String())
	}
	return ret_keys
}

type EncryptTensorString struct {
	W, H       int
	R, G, B, A [][]string
}

func (et *EncryptTensor) GetEncryptTensorString() EncryptTensorString {
	var ret_tensor EncryptTensorString
	ret_tensor.W = et.W
	ret_tensor.H = et.H
	for _, vi := range et.R {
		var tmp_slice []string
		for _, vj := range vi {
			tmp_slice = append(tmp_slice, vj.String())
		}
		ret_tensor.R = append(ret_tensor.R, tmp_slice)
	}
	for _, vi := range et.G {
		var tmp_slice []string
		for _, vj := range vi {
			tmp_slice = append(tmp_slice, vj.String())
		}
		ret_tensor.G = append(ret_tensor.G, tmp_slice)
	}
	for _, vi := range et.B {
		var tmp_slice []string
		for _, vj := range vi {
			tmp_slice = append(tmp_slice, vj.String())
		}
		ret_tensor.B = append(ret_tensor.B, tmp_slice)
	}
	for _, vi := range et.A {
		var tmp_slice []string
		for _, vj := range vi {
			tmp_slice = append(tmp_slice, vj.String())
		}
		ret_tensor.A = append(ret_tensor.A, tmp_slice)
	}
	return ret_tensor
}
