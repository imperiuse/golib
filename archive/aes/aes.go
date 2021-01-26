package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

type AES struct {
	key []byte
}

//nolint
func New(key string) *AES {
	md5secret := md5.Sum([]byte(key))
	return &AES{md5secret[:]}
}

func (a *AES) EncodeUrl(s string) (string, error) {
	data, err := EncryptCBC([]byte(s), a.key)
	if err != nil {
		return "", errors.WithMessage(err, "DecryptCBC error")
	}
	return base64.URLEncoding.EncodeToString(data), nil
}

func (a *AES) DecodeUrl(s string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", errors.WithMessage(err, "base64.URLEncoding.DecodeString error")
	}

	data, err := DecryptCBC(decoded, a.key)
	if err != nil {
		return "", errors.WithMessage(err, "DecryptCBC error")
	}
	return string(data), err
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(encrypt []byte) ([]byte, error) {
	padding := encrypt[len(encrypt)-1]
	ipadding := int(padding)
	if ipadding < 0 || ipadding >= len(encrypt) {
		return nil, errors.New("incorrect padding")
	}
	return encrypt[:len(encrypt)-int(padding)], nil
}

// AES-256 CBC MOD
func EncryptCBC(data []byte, key []byte) (ciphertext []byte, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("panic recovered in EncryptCBC: %v", p)
		}
	}()

	// Load your secret key from a safe place and reuse it across multiple
	// NewCipher calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	//key, _ := hex.DecodeString("6368616e676520746869732070617373")
	//plaintext := []byte("exampleplaintext")

	// CBC mode works on blocks so plaintexts may need to be padded to the
	// next whole block. For an example of such padding, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. Here we'll
	// assume that the plaintext is already of the correct length.
	data = PKCS5Padding(data, aes.BlockSize)

	if len(data)%aes.BlockSize != 0 {
		return nil, errors.New("Bad padding!")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.WithMessage(err, "NewCipher")
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext = make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, errors.WithMessage(err, "ReadFull rand")
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], data)

	return ciphertext, nil
}

func DecryptCBC(ciphertext []byte, key []byte) (data []byte, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("panic recovered in DecryptCBC: %v", p)
		}
	}()

	// Load your secret key from a safe place and reuse it across multiple
	// NewCipher calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	//key, _ := hex.DecodeString("6368616e676520746869732070617373")
	//ciphertext, _ := hex.DecodeString("73c86d43a9d700a253a96c85b0f6b03ac9792e0e757f869cca306bd3cba1c62b")

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.New("bad padding!")
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// CBC mode always works in whole blocks.
	if len(ciphertext)%aes.BlockSize != 0 {
		//fmt.Println(len(ciphertext))
		//fmt.Println(len(ciphertext) % aes.BlockSize)
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(ciphertext, ciphertext)

	// If the original plaintext lengths are not a multiple of the block
	// size, padding would have to be added when encrypting, which would be
	// removed at this point. For an example, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. However, it's
	// critical to note that ciphertexts must be authenticated (i.e. by
	// using crypto/hmac) before being decrypted in order to avoid creating
	// a padding oracle.

	//fmt.Printf("%s\n", ciphertext)
	//Output: exampleplaintext

	// Убираем лишние символы (от padding)
	data, err = PKCS5UnPadding(ciphertext)

	return data, err
}
