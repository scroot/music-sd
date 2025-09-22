package common

import (
	"crypto/aes"
	"encoding/hex"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

func AddHeader(req *http.Request) {
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Charset", "UTF-8,*;q=0.5")
	req.Header.Add("Accept-Encoding", "")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:60.0) Gecko/20100101 Firefox/60.0")
	req.Header.Add("referer", "http://google.com/")
}

// 加密postForm
func EncryptForm(requestBytes []byte) (encryptedString string) {
	//获取key的二进制字符串作为key
	key, _ := hex.DecodeString("7246674226682325323F5E6544673A51")
	encryptedBytes := AESEncrypt(requestBytes, key)
	//获取加密后的字符串
	encryptedString = hex.EncodeToString(encryptedBytes)
	//解密
	//https://tools.lami.la/jiami/aes
	//decodeBytes, _ := hex.DecodeString(encryptedString)
	//decrypted := AESDecrypt(decodeBytes, key)
	//fmt.Printf("%s", decrypted)
	//fmt.Println(encryptedString)
	return encryptedString

}

func AESEncrypt(src []byte, key []byte) (encrypted []byte) {
	cipher, _ := aes.NewCipher(generateKey(key))
	length := (len(src) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)
	copy(plain, src)
	pad := byte(len(plain) - len(src))
	for i := len(src); i < len(plain); i++ {
		plain[i] = pad
	}
	encrypted = make([]byte, len(plain))

	for bs, be := 0, cipher.BlockSize(); bs <= len(src); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Encrypt(encrypted[bs:be], plain[bs:be])
	}

	return encrypted
}

func AESDecrypt(encrypted []byte, key []byte) (decrypted []byte) {
	cipher, _ := aes.NewCipher(generateKey(key))
	decrypted = make([]byte, len(encrypted))
	for bs, be := 0, cipher.BlockSize(); bs < len(encrypted); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Decrypt(decrypted[bs:be], encrypted[bs:be])
	}

	trim := 0
	if len(decrypted) > 0 {
		trim = len(decrypted) - int(decrypted[len(decrypted)-1])
	}

	return decrypted[:trim]
}

func generateKey(key []byte) (genKey []byte) {
	genKey = make([]byte, 16)
	copy(genKey, key)
	for i := 16; i < len(key); {
		for j := 0; j < 16 && i < len(key); j, i = j+1, i+1 {
			genKey[j] ^= key[i]
		}
	}
	return genKey
}

// Head the url
func GetContentLen(url string) int {
	response, err := http.Head(url)
	if err != nil {
		log.Panic(err)
	}
	//fmt.Println(response.ContentLength, response.StatusCode)
	if response.StatusCode == http.StatusOK {
		len := response.Header.Get("Content-Length")
		if len != "" {
			i, err := strconv.Atoi(len)
			if err != nil {
				log.Panic(err)
			}
			return i
		}
		return 0
	}
	return 0
}

// 返回随机数
func Random(min, max int) int {
	return rand.Intn(max-min) + min
}
