package main

import (
	"encoding/json"
	"fmt"
	"image/png"
	"imgfe/src/encryptor"
	"imgfe/src/tensor"
	"io/ioutil"
	"os"
	"strings"
)

func WriteLastCredential(es encryptor.EncryptScheme, key_file string, kernel_size int, kernel_type string, key_length int, verbose int) error {
	curr_key_file := key_file[:strings.LastIndex(key_file, ".")] + "_" + kernel_type + key_file[strings.LastIndex(key_file, "."):]
	if verbose >= 1 {
		fmt.Println("***Write credentials to file: " + curr_key_file)
	}
	ec := es.GetEncryptCredential(kernel_size, kernel_type, key_length, len(es.GetFunctionKeysString())-1)
	credentials_byte, err := json.Marshal(ec)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(curr_key_file, credentials_byte, 0644)
	if err != nil {
		return err
	}
	return nil
}

func WriteLastCredentials(es encryptor.EncryptScheme, key_file string, kernel_size int, kernel_type string, key_length int, verbose int, num_credentials int) error {
	curr_key_file := key_file[:strings.LastIndex(key_file, ".")] + "_" + kernel_type + key_file[strings.LastIndex(key_file, "."):]
	if verbose >= 1 {
		fmt.Println("***Write credentials to file: " + curr_key_file)
	}
	ec := es.GetLastEncryptCredentials(kernel_size, kernel_type, key_length, num_credentials)
	credentials_byte, err := json.Marshal(ec)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(curr_key_file, credentials_byte, 0644)
	if err != nil {
		return err
	}
	return nil
}

func WriteCipher(et encryptor.EncryptTensor, cipher_file string, verbose int) error {
	if verbose >= 1 {
		fmt.Println("***Write cipher to file: " + cipher_file)
	}
	ets := et.GetEncryptTensorString()
	tensor_byte, err := json.Marshal(ets)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(cipher_file, tensor_byte, 0644)
	if err != nil {
		return err
	}
	return nil
}

func PrintImage(image_file string, t tensor.Tensor) error {
	outfile, err := os.Create(image_file)
	if err != nil {
		return err
	}
	png.Encode(outfile, t.ToImg())
	return nil
}

func PrintOriginalImage(image_file string, ot encryptor.OriginalTensor) error {
	outfilename_original := image_file[:strings.LastIndex(image_file, ".")] + "_tensor" + image_file[strings.LastIndex(image_file, "."):]
	err := PrintImage(outfilename_original, ot.Tensor)
	if err != nil {
		return err
	}
	fmt.Println("***Print the tensor of the original image: " + outfilename_original)
	return nil
}

func PrintPaddingImage(image_file string, pt tensor.Tensor) error {
	outfilename_padding := image_file[:strings.LastIndex(image_file, ".")] + "_padding" + image_file[strings.LastIndex(image_file, "."):]
	err := PrintImage(outfilename_padding, pt)
	if err != nil {
		return err
	}
	fmt.Println("***Print the tensor of the padding image: " + outfilename_padding)
	return nil
}

func PrintCroppingImage(image_file string, pt tensor.Tensor) error {
	outfilename_cropping := image_file[:strings.LastIndex(image_file, ".")] + "_cropping" + image_file[strings.LastIndex(image_file, "."):]
	err := PrintImage(outfilename_cropping, pt)
	if err != nil {
		return err
	}
	fmt.Println("***Print the tensor of the cropping image: " + outfilename_cropping)
	return nil
}
