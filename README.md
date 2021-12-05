# ImageFE

Functional encryption for images.

## Introduction

In the traditional cryptography framework, a decryptor either recovers the *entire* plaintext from the ciphertext or recovers *nothing*. FE allows fine-grained control over the amount of information each specific decryptor can reveal from the ciphertext. 

ImageFE adopts a simplified FE scheme. There are two participants, the trusted encryptor and the decryptor. The encryptor *i)* generates a universal ciphertext from a given plaintext; *ii)* generates a set of different function keys. The type and amount of information a decryptor can reveal is determined by the function key it receives.

## Methodology

Efficient FE only exists for inner products and quadratic polynomials. ImageFE selects several convolution kernels to obfuscate an image. Note that convolution can be considered as a sequence of inner products. Given an original image, ImageFE derives a function key for every convolution kernel. Decrypting with a function key has the same effect as applying the corresponding convolution kernel to the original image. The function key that corresponds to the identity kernel can recover the original image.

Currently, ImageFE uses Decisional Diffie-Hellman ([DDH](https://eprint.iacr.org/2015/017.pdf)) as the underlying FE algorithm.

## Application

Consider the use case of blockchain. The ciphertext can be released to all users. A function key can be protected by public-key encryption schemes widely used by blockchains. The encryption and transmission of function keys to specific users can be triggered automatically by smart contracts. In this way, ImageFE can be naturally integrated with DEFI and NFT.

&nbsp;

# Use ImageFE

## Dependency

<https://go.dev/doc/install>

## Building from source

```sh
cd imgfe
cd cmd
go build -o imgfe.exe
```

## Usages

The *enc* command takes an original image and a specification of convolution kernels. It produces the ciphertext and the function keys.

The *dec* command takes the ciphertext and a function key. It produces an image that is equivalent to the result of applying the corresponding convolution kernel to the original image. Multiple function keys in one mode share the same ciphertext.

__Help messages__
```sh
./imgfe.exe enc -h
./image.exe dec -h
```

__Down resolution mode__

*downsize_partial* yields an image that is scaled down *(kernel_offset+1)* folds.

*downsize_full* yields the original image.

```sh
./imgfe.exe enc -kernel_type="downsize_partial,downsize_full" 

./imgfe.exe dec -image_file="../workspace/cart_partial.png" -key_file="../workspace/func_key_downsize_partial.json"

./imgfe.exe dec -image_file="../workspace/cart_full.png" -key_file="../workspace/func_key_downsize_full.json"
```

__Blur mode__

*gaussian_blur* and *box_blur* yield blurred images. Kernel size equals *(kernel_offset\*2+1)*.

*identity* yields the original image.

```sh
./imgfe.exe enc -kernel_type="identity,gaussian_blur,box_blur" 

./imgfe.exe dec -image_file="../workspace/cart_gaussian_blur.png" -key_file="../workspace/func_key_gaussian_blur.json"

./imgfe.exe dec -image_file="../workspace/cart_identity.png" -key_file="../workspace/func_key_identity.json"
```

&nbsp;

# Miscellaneous

## Contact

The *Dappanomics Lab* is founded by Northwestern faculties, alumni and students. We are currently developing DApps surrounding NFT and DEFI. If you are interested in our project, please contact us at *you.li at u.northwestern.edu* .

## Acknowledgements

[Dr. Tilen Marc](https://www.researchgate.net/profile/Tilen-Marc) for his kind support.

The [FENTEC project](https://github.com/fentec-project).