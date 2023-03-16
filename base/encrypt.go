package base

import (
	"encoding/base64"
	"github.com/forgoer/openssl"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

type Encrypt struct {
	Key string
	Iv  string
}

// EncryptData 加密
// @auth frog
// @date 2023-03-14 15:12:25
func (handle *Encrypt) EncryptData(src string) string {
	byteKey := []byte(handle.Key)
	for i := len(byteKey); i < 8; i++ {
		byteKey = append(byteKey, 0x00)
	}
	byteIv := []byte(handle.Iv)
	for i := len(byteIv); i < 8; i++ {
		byteIv = append(byteIv, 0x00)
	}
	byteRet, err := openssl.DesCBCEncrypt([]byte(src), byteKey, byteIv, openssl.PKCS7_PADDING)
	if err != nil {
		log.Error(`加密失败 %s %s %s %s`, src, handle.Key, handle.Iv, err.Error())
		return ``
	}
	return cast.ToString(base64.StdEncoding.EncodeToString(byteRet))
}

// DecryptData 解密
// @auth frog
// @date 2023-03-14 15:19:31
func (handle *Encrypt) DecryptData(src string) string {
	byteRet, err := openssl.DesCBCDecrypt([]byte(src), []byte(handle.Key), []byte(handle.Iv), openssl.PKCS7_PADDING)
	if err != nil {
		log.Error(`反解失败 %s`, err.Error())
		return ``
	}
	return cast.ToString(byteRet)
}
