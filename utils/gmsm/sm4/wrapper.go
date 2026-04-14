package sm4

import (
	"bytes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
)

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		return []byte{}
	}
	unpadding := int(origData[length-1])
	if length < unpadding {
		return []byte{}
	}
	return origData[:(length - unpadding)]
}

func deriveKey(pwd []byte) []byte {
	h := md5.Sum(pwd)
	return h[:]
}

func Sm4EncryptToHex(pwd []byte, data []byte) (string, error) {
	key := deriveKey(pwd)
	block, err := NewCipher(key)
	if err != nil {
		return "", err
	}
	data = PKCS7Padding(data, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key) // using key as IV as well
	crypted := make([]byte, len(data))
	blockMode.CryptBlocks(crypted, data)
	return hex.EncodeToString(crypted), nil
}

func Sm4DecryptFromHex(pwd []byte, hexStr string) ([]byte, error) {
	crypted, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	key := deriveKey(pwd)
	block, err := NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(crypted) == 0 || len(crypted)%block.BlockSize() != 0 {
		return nil, nil
	}

	blockMode := cipher.NewCBCDecrypter(block, key)
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS7UnPadding(origData)
	return origData, nil
}
