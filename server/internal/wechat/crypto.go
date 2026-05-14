package wechat

// WeChat 开放平台 / 公众平台消息加解密。
// 参考官方文档「消息加解密技术方案」：
//   https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/Message_Encryption/Message_encryption_and_decryption.html

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// EncryptedEnvelope 是微信推送到回调 URL 的 XML 消息最外层
type EncryptedEnvelope struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string   `xml:"ToUserName"`
	Encrypt    string   `xml:"Encrypt"`
	AppID      string   `xml:"AppId"`
}

// SignatureValid 校验消息签名：sha1(sorted(token, timestamp, nonce, encrypt))
func SignatureValid(token, timestamp, nonce, encrypt, signature string) bool {
	arr := []string{token, timestamp, nonce, encrypt}
	sort.Strings(arr)
	h := sha1.Sum([]byte(strings.Join(arr, "")))
	return hex.EncodeToString(h[:]) == signature
}

// DecodeAESKey 把 43 位 EncodingAESKey 还原为 32 字节 AES 密钥
func DecodeAESKey(encodingAESKey string) ([]byte, error) {
	if len(encodingAESKey) != 43 {
		return nil, errors.New("EncodingAESKey 长度必须为 43")
	}
	key, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, errors.New("EncodingAESKey 解码长度不对")
	}
	return key, nil
}

// Decrypt 将 base64(encrypt) 解密成明文 payload
// 规范：random(16) + msg_len(4BE) + msg + appid
func Decrypt(encryptedB64 string, aesKey []byte) (payload []byte, appID string, err error) {
	cipherText, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return nil, "", err
	}
	if len(cipherText) < aes.BlockSize || len(cipherText)%aes.BlockSize != 0 {
		return nil, "", errors.New("AES ciphertext 长度异常")
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, "", err
	}
	iv := aesKey[:aes.BlockSize]
	mode := cipher.NewCBCDecrypter(block, iv)
	plain := make([]byte, len(cipherText))
	mode.CryptBlocks(plain, cipherText)
	plain = pkcs7Unpad(plain)
	if len(plain) < 20 {
		return nil, "", errors.New("解密明文长度不足")
	}
	msgLen := binary.BigEndian.Uint32(plain[16:20])
	if int(msgLen) > len(plain)-20 {
		return nil, "", errors.New("msg_len 超长")
	}
	payload = plain[20 : 20+msgLen]
	appID = string(plain[20+msgLen:])
	return payload, appID, nil
}

// Encrypt 将明文封装成 base64(aes_encrypt(random16 + msg_len + msg + appid))
func Encrypt(plain []byte, appID string, aesKey []byte, nonce16 []byte) (string, error) {
	if len(nonce16) != 16 {
		return "", errors.New("nonce 必须 16 字节")
	}
	var buf bytes.Buffer
	buf.Write(nonce16)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(plain)))
	buf.Write(plain)
	buf.WriteString(appID)
	src := pkcs7Pad(buf.Bytes(), aes.BlockSize)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	iv := aesKey[:aes.BlockSize]
	mode := cipher.NewCBCEncrypter(block, iv)
	dst := make([]byte, len(src))
	mode.CryptBlocks(dst, src)
	return base64.StdEncoding.EncodeToString(dst), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	if padLen == 0 {
		padLen = blockSize
	}
	return append(data, bytes.Repeat([]byte{byte(padLen)}, padLen)...)
}

func pkcs7Unpad(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	padLen := int(data[len(data)-1])
	if padLen < 1 || padLen > aes.BlockSize {
		return data
	}
	return data[:len(data)-padLen]
}

// BuildReply 用于给微信消息加密回包
func BuildReply(plain []byte, appID string, token string, aesKey []byte, timestamp, nonce string, random16 []byte) ([]byte, error) {
	enc, err := Encrypt(plain, appID, aesKey, random16)
	if err != nil {
		return nil, err
	}
	arr := []string{token, timestamp, nonce, enc}
	sort.Strings(arr)
	h := sha1.Sum([]byte(strings.Join(arr, "")))
	sig := hex.EncodeToString(h[:])

	reply := fmt.Sprintf(`<xml>
<Encrypt><![CDATA[%s]]></Encrypt>
<MsgSignature><![CDATA[%s]]></MsgSignature>
<TimeStamp>%s</TimeStamp>
<Nonce><![CDATA[%s]]></Nonce>
</xml>`, enc, sig, timestamp, nonce)
	return []byte(reply), nil
}
