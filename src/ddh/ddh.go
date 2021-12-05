package ddh

import (
	"fmt"
	"math/big"

	"imgfe/pkg/gofe/data"
	"imgfe/pkg/gofe/innerprod/simple"
)

type DDH struct {
	simpleDDH    *simple.DDH
	masterSecKey data.Vector
	masterPubKey data.Vector
	funcKeys     []*big.Int
	y            [][]*big.Int
}

func NewDDH(l int, modulus_length int, bnd uint32) (*DDH, error) {
	bound := big.NewInt(int64(bnd))

	var simpleDDH *simple.DDH
	var err error

	simpleDDH, err = simple.NewDDH(l, modulus_length, bound)

	if err != nil {
		return nil, fmt.Errorf("error during simple inner product creation: %v", err)
	}

	masterSecKey, masterPubKey, err := simpleDDH.GenerateMasterKeys()

	if err != nil {
		return nil, fmt.Errorf("error during master key generation: %v", err)
	}

	return &DDH{
		simpleDDH:    simpleDDH,
		masterSecKey: masterSecKey,
		masterPubKey: masterPubKey,
		funcKeys:     make([]*big.Int, 0),
		y:            make([][]*big.Int, 0),
	}, nil
}

func NewDDHPrecomp(l int, modulus_length int, bnd uint32) (*DDH, error) {
	bound := big.NewInt(int64(bnd))

	var simpleDDH *simple.DDH
	var err error

	simpleDDH, err = simple.NewDDHPrecomp(l, modulus_length, bound) // youl

	if err != nil {
		return nil, fmt.Errorf("error during simple inner product creation: %v", err)
	}

	masterSecKey, masterPubKey, err := simpleDDH.GenerateMasterKeys()

	if err != nil {
		return nil, fmt.Errorf("error during master key generation: %v", err)
	}

	return &DDH{
		simpleDDH:    simpleDDH,
		masterSecKey: masterSecKey,
		masterPubKey: masterPubKey,
		funcKeys:     make([]*big.Int, 0),
		y:            make([][]*big.Int, 0),
	}, nil
}

func (d *DDH) FunctionDDH(mask []uint32) error {
	y := make([]*big.Int, 0)
	for i := range mask {
		y = append(y, big.NewInt(int64(int(mask[i]))))
	}
	d.y = append(d.y, y)

	funcKey, err := d.simpleDDH.DeriveKey(d.masterSecKey, y)
	d.funcKeys = append(d.funcKeys, funcKey)

	if err != nil {
		return fmt.Errorf("error during function key derivation: %v", err)
	} else {
		return nil
	}
}

func (d *DDH) EncryptDDH(row []uint32) (data.Vector, error) {
	x := make([]*big.Int, 0)
	for i := range row {
		x = append(x, big.NewInt(int64(int(row[i]))))
	}

	encryptor := simple.NewDDHFromParams(d.simpleDDH.Params)

	ciphertext, err := encryptor.Encrypt(x, d.masterPubKey)

	if err != nil {
		return nil, fmt.Errorf("error during encryption: %v", err)
	}

	return ciphertext, nil
}

func (d *DDH) GetDDHParams() simple.DDHParams {
	return *d.simpleDDH.Params
}

func (d *DDH) GetYMasks() [][]*big.Int {
	return d.y
}

func (d *DDH) GetFunctionKeys() []*big.Int {
	return d.funcKeys
}

type DDHParamsString struct {
	L     int
	Bound string
	G     string
	P     string
	Q     string
}
