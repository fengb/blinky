package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

type Aes struct {
	cip cipher.Block
}

func NewAes(key []byte) (*Aes, error) {
	a := Aes{}
	cip, err := aes.NewCipher(a.pad(key))
	if err != nil {
		return nil, err
	}

	a.cip = cip

	return &a, nil
}

func (a *Aes) Encrypt(plaintext []byte) ([]byte, error) {
	plaintext = a.pad(plaintext)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	_, err := rand.Read(iv)
	if err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBEncrypter(a.cip, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

func (a *Aes) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("blocksize must be multiple of decoded message length")
	}

	iv := ciphertext[:aes.BlockSize]
	msg := ciphertext[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(a.cip, iv)
	cfb.XORKeyStream(msg, msg)

	return msg, nil
}

func (a *Aes) pad(src []byte) []byte {
	n := aes.BlockSize - len(src)%aes.BlockSize
	return append(src, make([]byte, n)...)
}
