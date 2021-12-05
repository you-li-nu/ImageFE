package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	enc_command := flag.NewFlagSet("enc", flag.ExitOnError)
	enc_image_file := enc_command.String("image_file", "../workspace/cart.png", "path to the input image")
	enc_cipher_file := enc_command.String("cipher_file", "../workspace/cart_cipher.json", "path to the output cipher")
	enc_key_file := enc_command.String("key_file", "../workspace/func_key.json", "path to the output functional key file")
	enc_key_length := enc_command.Int("key_length", 256, "length of the prime number used in Diffie Hellman encryption, (128/256/512/1024/1536/2048/2560/3072/4096)")
	enc_kernel_type := enc_command.String("kernel_type", "identity,gaussian_blur,box_blur", "[identity/gaussian_blur/box_blur], [downsize_full/downsize_partial]")
	enc_kernel_offset := enc_command.Int("kernel_offset", 1, "offset of the kernel, (1/2)")
	enc_verbose := enc_command.Int("verbose", 1, "verbosity, (0/1/2)")

	dec_command := flag.NewFlagSet("dec", flag.ExitOnError)
	dec_image_file := dec_command.String("image_file", "../workspace/cart2.png", "path to the output image")
	dec_cipher_file := dec_command.String("cipher_file", "../workspace/cart_cipher.json", "path to the input cipher")
	dec_key_file := dec_command.String("key_file", "../workspace/func_key_identity.json", "path to the input functional key file")
	dec_verbose := dec_command.Int("verbose", 1, "verbosity, (0/1/2)")

	if len(os.Args) < 2 {
		fmt.Println("expected 'enc' or 'dec' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {

	case "enc":
		enc_command.Parse(os.Args[2:])
		fmt.Println("subcommand 'enc'")
		fmt.Println("  image_file:", *enc_image_file)
		fmt.Println("  cipher_file:", *enc_cipher_file)
		fmt.Println("  key_file:", *enc_key_file)
		fmt.Println("  key_length:", *enc_key_length)
		fmt.Println("  kernel_type:", *enc_kernel_type)
		fmt.Println("  kernel_offset:", *enc_kernel_offset)
		fmt.Println("  tail:", enc_command.Args())
		run_enc_command(*enc_image_file, *enc_cipher_file, *enc_key_file, *enc_key_length, *enc_kernel_type, *enc_kernel_offset, *enc_verbose)

	case "dec":
		dec_command.Parse(os.Args[2:])
		fmt.Println("subcommand 'enc'")
		fmt.Println("  image_file:", *dec_image_file)
		fmt.Println("  cipher_file:", *dec_cipher_file)
		fmt.Println("  key_file:", *dec_key_file)
		fmt.Println("  tail:", dec_command.Args())
		run_dec_command(*dec_image_file, *dec_cipher_file, *dec_key_file, *dec_verbose)

	default:
		fmt.Println("expected 'enc' or 'dec' subcommands")
		os.Exit(1)
	}
}
