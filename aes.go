package main

import (
	"bytes"
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
	if !a.isAligned(key) {
		key = a.pad(key)
	}
	cip, err := aes.NewCipher(key)
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
	if !a.isAligned(ciphertext) {
		return nil, errors.New("blocksize must be multiple of decoded message length")
	}

	iv := ciphertext[:aes.BlockSize]
	msg := ciphertext[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(a.cip, iv)
	cfb.XORKeyStream(msg, msg)

	plaintext, err := a.unpad(msg)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (a *Aes) isAligned(b []byte) bool {
	return len(b)%aes.BlockSize == 0
}

func (a *Aes) pad(src []byte) []byte {
	n := aes.BlockSize - len(src)%aes.BlockSize
	fill := []byte{byte(n)}
	return append(src, bytes.Repeat(fill, n)...)
}

func (a *Aes) unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}
