package wecom

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// VerifyURL 企业微信回调URL验证
// 对收到的echostr进行签名校验并解密，返回明文
func VerifyURL(msgSignature, timestamp, nonce, echostr, token, encodingAESKey string) (string, error) {
	// 1. 解密echostr
	plaintext, corpID, err := aesDecrypt(encodingAESKey, echostr)
	if err != nil {
		return "", fmt.Errorf("AES解密失败: %w", err)
	}

	// 2. 验证签名
	sig := calcSignature(token, timestamp, nonce, string(plaintext))
	if sig != msgSignature {
		return "", errors.New("签名验证失败")
	}

	_ = corpID // 生产环境可校验corpID
	return string(plaintext), nil
}

// calcSignature 计算企微回调签名
// 将 token, timestamp, nonce, msg 按字典序排序后拼接，做 SHA1
func calcSignature(token, timestamp, nonce, msg string) string {
	arr := []string{token, timestamp, nonce, msg}
	sort.Strings(arr)
	raw := strings.Join(arr, "")
	h := sha1.New()
	h.Write([]byte(raw))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// aesDecrypt 使用EncodingAESKey解密企微回调密文
// 返回: 解密后的消息明文, 接收方corpID, error
func aesDecrypt(encodingAESKey, ciphertextB64 string) ([]byte, string, error) {
	// EncodingAESKey是43位，补"="组成标准Base64
	aesKey, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		return nil, "", fmt.Errorf("EncodingAESKey base64解码失败: %w", err)
	}
	if len(aesKey) != 32 {
		return nil, "", fmt.Errorf("AES密钥长度错误: %d (期望32)", len(aesKey))
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, "", fmt.Errorf("密文base64解码失败: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, "", fmt.Errorf("创建AES cipher失败: %w", err)
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, "", errors.New("密文长度不足")
	}

	// AES-256-CBC, IV = key前16字节
	iv := aesKey[:aes.BlockSize]
	mode := cipher.NewCBCDecrypter(block, iv)

	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 去除PKCS7填充
	plaintext, err = pkcs7Unpad(plaintext)
	if err != nil {
		return nil, "", fmt.Errorf("PKCS7去填充失败: %w", err)
	}

	// 解析: [16字节随机数][4字节msg_len(big-endian)][msg][corpID]
	if len(plaintext) < 20 {
		return nil, "", errors.New("明文长度不足")
	}

	msgLen := binary.BigEndian.Uint32(plaintext[16:20])
	if int(20+msgLen) > len(plaintext) {
		return nil, "", errors.New("消息长度超出明文范围")
	}

	msg := plaintext[20 : 20+msgLen]
	corpID := string(plaintext[20+msgLen:])

	return msg, corpID, nil
}

// pkcs7Unpad 去除PKCS7填充
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("空数据无法去填充")
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding > aes.BlockSize || padding == 0 {
		return nil, fmt.Errorf("无效的PKCS7填充: %d", padding)
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, errors.New("PKCS7填充验证失败")
		}
	}
	return data[:len(data)-padding], nil
}

// pkcs7Pad PKCS7填充
func pkcs7Pad(data []byte) []byte {
	padding := aes.BlockSize - len(data)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// EncryptMsg 加密回复消息（用于被动回复）
func EncryptMsg(msg, corpID, encodingAESKey string) (string, error) {
	aesKey, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		return "", fmt.Errorf("EncodingAESKey base64解码失败: %w", err)
	}

	// 构造明文: [16字节随机数][4字节msg_len][msg][corpID]
	randBytes := make([]byte, 16)
	// 简化：用固定随机（生产环境应用crypto/rand）
	for i := range randBytes {
		randBytes[i] = byte(i + 1)
	}

	msgBytes := []byte(msg)
	msgLen := make([]byte, 4)
	binary.BigEndian.PutUint32(msgLen, uint32(len(msgBytes)))

	var buf bytes.Buffer
	buf.Write(randBytes)
	buf.Write(msgLen)
	buf.Write(msgBytes)
	buf.Write([]byte(corpID))

	plaintext := pkcs7Pad(buf.Bytes())

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, len(plaintext))
	iv := aesKey[:aes.BlockSize]
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
