package lxruntime

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"io"
	"strings"

	"github.com/dop251/goja"
)

func buildUtils(vm *goja.Runtime) *goja.Object {
	utils := vm.NewObject()

	buffer := vm.NewObject()
	_ = buffer.Set("from", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Undefined()
		}
		input, format := call.Arguments[0], "utf8"
		if len(call.Arguments) > 1 {
			format = call.Arguments[1].String()
		}

		switch strings.ToLower(format) {
		case "base64":
			data, err := base64.StdEncoding.DecodeString(input.String())
			if err != nil {
				panic(vm.ToValue(err.Error()))
			}
			return vm.ToValue(data)
		case "hex":
			data, err := hex.DecodeString(input.String())
			if err != nil {
				panic(vm.ToValue(err.Error()))
			}
			return vm.ToValue(data)
		default:
			return vm.ToValue([]byte(input.String()))
		}
	})
	_ = buffer.Set("bufToString", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		data, err := decodeBufferInput(call.Arguments[0].Export())
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		format := "utf8"
		if len(call.Arguments) > 1 {
			format = call.Arguments[1].String()
		}
		return vm.ToValue(encodeBuffer(data, format))
	})
	_ = utils.Set("buffer", buffer)

	cryptoObj := vm.NewObject()
	_ = cryptoObj.Set("md5", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		sum := md5.Sum([]byte(call.Arguments[0].String()))
		return vm.ToValue(hex.EncodeToString(sum[:]))
	})
	_ = cryptoObj.Set("randomBytes", func(call goja.FunctionCall) goja.Value {
		size := 16
		if len(call.Arguments) > 0 {
			size = int(call.Arguments[0].ToInteger())
		}
		buf := make([]byte, size)
		if _, err := rand.Read(buf); err != nil {
			panic(vm.ToValue(err.Error()))
		}
		return vm.ToValue(buf)
	})
	_ = cryptoObj.Set("aesEncrypt", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 4 {
			panic(vm.ToValue("aesEncrypt requires buffer, mode, key, iv"))
		}
		data, err := decodeBufferInput(call.Arguments[0].Export())
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		mode := call.Arguments[1].String()
		key, err := decodeBufferInput(call.Arguments[2].Export())
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		iv, err := decodeBufferInput(call.Arguments[3].Export())
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}

		switch strings.ToLower(mode) {
		case "aes-128-cbc":
			block, err := aes.NewCipher(key)
			if err != nil {
				panic(vm.ToValue(err.Error()))
			}
			padded := pkcs7Pad(data, block.BlockSize())
			ciphertext := make([]byte, len(padded))
			cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)
			return vm.ToValue(ciphertext)
		case "aes-128-ecb":
			block, err := aes.NewCipher(key)
			if err != nil {
				panic(vm.ToValue(err.Error()))
			}
			padded := pkcs7Pad(data, block.BlockSize())
			ciphertext := make([]byte, len(padded))
			for start := 0; start < len(padded); start += block.BlockSize() {
				block.Encrypt(ciphertext[start:start+block.BlockSize()], padded[start:start+block.BlockSize()])
			}
			return vm.ToValue(ciphertext)
		default:
			return panicJS(vm, "unsupported aes mode")
		}
	})
	_ = cryptoObj.Set("rsaEncrypt", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.ToValue("rsaEncrypt requires buffer and key"))
		}
		data, err := decodeBufferInput(call.Arguments[0].Export())
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		keyText := call.Arguments[1].String()
		block, _ := pem.Decode([]byte(keyText))
		if block == nil {
			panic(vm.ToValue("invalid rsa key"))
		}
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			panic(vm.ToValue("invalid rsa public key"))
		}
		encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, data)
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		return vm.ToValue(encrypted)
	})
	_ = utils.Set("crypto", cryptoObj)

	zlibObj := vm.NewObject()
	_ = zlibObj.Set("inflate", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Undefined()
		}
		data, err := decodeBufferInput(call.Arguments[0].Export())
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		reader, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			reader = flate.NewReader(bytes.NewReader(data))
		}
		defer reader.Close()
		out, err := io.ReadAll(reader)
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		return vm.ToValue(out)
	})
	_ = zlibObj.Set("deflate", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Undefined()
		}
		data, err := decodeBufferInput(call.Arguments[0].Export())
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		var buf bytes.Buffer
		writer := zlib.NewWriter(&buf)
		if _, err := writer.Write(data); err != nil {
			panic(vm.ToValue(err.Error()))
		}
		if err := writer.Close(); err != nil {
			panic(vm.ToValue(err.Error()))
		}
		return vm.ToValue(buf.Bytes())
	})
	_ = utils.Set("zlib", zlibObj)

	return utils
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func panicJS(vm *goja.Runtime, message string) goja.Value {
	panic(vm.ToValue(message))
	return goja.Undefined()
}
