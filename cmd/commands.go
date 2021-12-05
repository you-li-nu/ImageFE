package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"imgfe/src/decryptor"
	"imgfe/src/encryptor"
	"imgfe/src/kernel"
	"imgfe/src/tensor"
)

func run_enc_command(image_file string, cipher_file string, key_file string, key_length int, kernel_types string, kernel_offset int, verbose int) {
	// process kernel
	kernel_types_array := strings.Split(kernel_types, ",")
	var kernel_size int
	switch kernel_types_array[0] {
	case "identity", "gaussian_blur", "box_blur":
		kernel_size = 2*kernel_offset + 1
	case "downsize_partial", "downsize_full":
		kernel_size = kernel_offset + 1
	default:
		panic("unrecognized kernel: " + kernel_types_array[0])
	}

	// read the original image to tensor
	ot, err := encryptor.NewOriginalTensor(image_file)
	if err != nil {
		panic(err.Error())
	}
	if verbose >= 2 {
		err = PrintOriginalImage(image_file, *ot)
		if err != nil {
			panic(err.Error())
		}
	}

	// construct the encryption scheme
	es, err := encryptor.NewEncryptSchemePrecomp(kernel_size*kernel_size, key_length)
	if err != nil {
		panic(err.Error())
	}

	// derive function keys for kernels and write to files
	for _, kernel_type := range kernel_types_array {
		var mask []uint32
		var divisor uint32
		switch kernel_type {
		case "identity":
			mask, divisor, err = kernel.Identity(kernel_size)
			if err != nil {
				panic(err.Error())
			}
			es.DeriveFunctionKey(mask)
			err = WriteLastCredential(es, key_file, kernel_size, kernel_type, key_length, verbose)
		case "gaussian_blur":
			mask, divisor, err = kernel.Gaussian_blur(kernel_size)
			if err != nil {
				panic(err.Error())
			}
			es.DeriveFunctionKey(mask)
			err = WriteLastCredential(es, key_file, kernel_size, kernel_type, key_length, verbose)
		case "box_blur":
			mask, divisor, err = kernel.Box_blur(kernel_size)
			if err != nil {
				panic(err.Error())
			}
			es.DeriveFunctionKey(mask)
			err = WriteLastCredential(es, key_file, kernel_size, kernel_type, key_length, verbose)
		case "downsize_partial":
			mask, divisor, err = kernel.Downsize_partial(kernel_size, 2)
			if err != nil {
				panic(err.Error())
			}
			es.DeriveFunctionKey(mask)
			err = WriteLastCredential(es, key_file, kernel_size, kernel_type, key_length, verbose)
		case "downsize_full":
			for kernel_idx := 0; kernel_idx <= kernel_size*kernel_size-1; kernel_idx++ {
				mask, divisor, err = kernel.Downsize_partial(kernel_size, kernel_idx)
				if err != nil {
					panic(err.Error())
				}
				es.DeriveFunctionKey(mask)
			}
			fmt.Println(es.GetYMasksString())
			err = WriteLastCredentials(es, key_file, kernel_size, kernel_type, key_length, verbose, kernel_size*kernel_size)
		}
		if err != nil {
			panic(err.Error())
		}
		if verbose >= 2 {
			fmt.Println("***Mask and divisor of kernel: " + kernel_type)
			fmt.Println(mask)
			fmt.Println(divisor)
		}
	}
	func_keys := es.GetFunctionKeysString()
	if verbose >= 1 {
		fmt.Println("***Function keys:")
		fmt.Println(func_keys)
	}

	// construct the encrypted tensor
	et := encryptor.NewEncryptTensor(ot.Tensor.W, ot.Tensor.H)
	var pt tensor.Tensor
	switch kernel_types_array[0] {
	case "identity", "gaussian_blur", "box_blur":
		pt = ot.PaddingCenter(kernel_offset, 0)
		if verbose >= 2 {
			err = PrintPaddingImage(image_file, pt)
			if err != nil {
				panic(err.Error())
			}
		}
	case "downsize_partial", "downsize_full":
		pt = ot.CroppingTopLeft(kernel_size, 0)
		if verbose >= 2 {
			err = PrintCroppingImage(image_file, pt)
			if err != nil {
				panic(err.Error())
			}
		}
	}

	if verbose >= 1 {
		fmt.Println("***Encryption started.")
	}
	switch kernel_types_array[0] {
	case "identity", "gaussian_blur", "box_blur":
		et, err = es.EncryptCenter(kernel_offset, pt)
	case "downsize_partial", "downsize_full":
		et, err = es.EncryptTopLeft(kernel_size, pt)
	}
	if err != nil {
		panic(err.Error())
	}
	if verbose >= 1 {
		fmt.Println("***Encryption finished.")
	}

	// write encrypted tensor to file
	err = WriteCipher(et, cipher_file, verbose)
	if err != nil {
		panic(err.Error())
	}
}

func run_dec_command(image_file string, cipher_file string, key_file string, verbose int) {
	//read credentials and construct the decrypted scheme
	if verbose >= 1 {
		fmt.Println("***Read credentials from file: " + key_file)
	}
	credentials_byte_read, err := ioutil.ReadFile(key_file)
	if err != nil {
		panic(err.Error())
	}
	var ec encryptor.EncryptCredentials
	err = json.Unmarshal(credentials_byte_read, &ec)
	if err != nil {
		panic(err.Error())
	}

	es, err := encryptor.NewEncryptSchemePrecomp(ec.Kernel_size*ec.Kernel_size, ec.Key_length)
	if err != nil {
		panic(err.Error())
	}
	params, _, _ := es.Get_Params_Y_FK()
	ds, err := decryptor.NewDecryptSchemePrecomp(params, decryptor.GetYMasksFromString(ec.Y_mask), decryptor.GetFunctionKeysFromString(ec.Func_key), verbose)
	if err != nil {
		panic(err.Error())
	}

	//read cipher and launch decryption
	if verbose >= 1 {
		fmt.Println("***Read cipher from file: " + cipher_file)
	}
	tensor_byte_read, err := ioutil.ReadFile(cipher_file)
	if err != nil {
		panic(err.Error())
	}
	var ets encryptor.EncryptTensorString
	err = json.Unmarshal(tensor_byte_read, &ets)
	if err != nil {
		panic(err.Error())
	}

	if verbose >= 1 {
		fmt.Println("***Decryption started.")
	}
	var t tensor.Tensor
	switch ec.Kernel_type {
	case "identity", "gaussian_blur", "box_blur":
		t, err = ds.DecryptSingle(decryptor.GetEncryptTensorFromString(ets), 0)
	case "downsize_partial":
		t, err = ds.DecryptSingle(decryptor.GetEncryptTensorFromString(ets), 0)
	case "downsize_full":
		t, err = ds.DecryptMultiple(decryptor.GetEncryptTensorFromString(ets), ec, verbose)
	}
	if err != nil {
		panic(err.Error())
	}
	if verbose >= 1 {
		fmt.Println("***Decryption finished.")
	}

	// tensor to output image
	if verbose >= 1 {
		fmt.Println("***Write image to file: " + image_file)
	}
	PrintImage(image_file, t)
}
