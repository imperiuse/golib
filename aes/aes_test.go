package aes

//nolint
import (
	"crypto/md5"
	"crypto/rand"
	"io"
	"testing"
)

const key = "secret_key"

// go test -covermode=count -coverprofile=coverage.cov && go tool cover -html=coverage.cov

func TestNew(t *testing.T) {
	a := New(key)
	if a == nil {
		t.Errorf("return nil pointer")
	}
}

func Test_PKCS5Padding_PKCS5UnPadding(t *testing.T) {
	var testCases = []struct{ len, block int }{
		{17, 16},
		{33, 32},
		{65, 64},
		{129, 128},
	}

	for i, v := range testCases {
		for j := 0; j < v.len; j++ {
			temp := make([]byte, j)
			if _, err := io.ReadFull(rand.Reader, temp); err != nil {
				t.Errorf("ReadFull rand err:%v", err)
			}
			p := PKCS5Padding(temp, v.block)
			if len(p)%v.block != 0 {
				t.Errorf("bad padding in %d:%d", i, j)
			}

			u, err := PKCS5UnPadding(p)
			if j != 0 && err != nil {
				t.Errorf("PKCS5UnPadding error in %d:%d, err:%v", i, j, err)
			}

			if len(u) != len(temp) {
				t.Errorf("bad unpadding in %d:%d", i, j)
			}
		}
	}
}

//nolint
func Test_EncryptCBD_DecryptCBC(t *testing.T) {
	key := md5.Sum([]byte(key))

	for i := 0; i < 1024; i++ {
		var data []byte
		if i > 0 {
			data = make([]byte, i)
		}
		if _, err := io.ReadFull(rand.Reader, data); err != nil {
			t.Errorf("ReadFull rand err:%v", err)
		}

		encrypted, err := EncryptCBC(data, key[:])
		if i != 0 && err != nil {
			t.Errorf("EncryptCBC error in %d case, err:%v", i, err)
		}

		decrypted, err := DecryptCBC(encrypted, key[:])
		if i != 0 && err != nil {
			t.Errorf("DecryptCBC error in %d case, err:%v", i, err)
			t.Error(data, encrypted, decrypted)
		}

		if string(decrypted) != string(data) {
			t.Errorf("Not expected decrypted data in %d case", i)
			t.Error(encrypted, decrypted)
		}
	}
}

func TestAES_EncodeUrl_DecodeUrl(t *testing.T) {
	aes := New(key)

	for i := 0; i < 1024; i++ {
		var data []byte
		if i > 0 {
			data = make([]byte, i)
		}
		if _, err := io.ReadFull(rand.Reader, data); err != nil {
			t.Errorf("ReadFull rand err:%v", err)
		}

		encrypted, err := aes.EncodeUrl(string(data))
		if i != 0 && err != nil {
			t.Errorf("EncodeUrl error in %d case, err:%v", i, err)
		}

		decrypted, err := aes.DecodeUrl(encrypted)
		if i != 0 && err != nil {
			t.Errorf("DecodeUrl error in %d case, err:%v", i, err)
			t.Error(data, encrypted, decrypted)
		}

		if decrypted != string(data) {
			t.Errorf("Not expected decrypted data in %d case", i)
			t.Error(encrypted, decrypted)
		}
	}
}
