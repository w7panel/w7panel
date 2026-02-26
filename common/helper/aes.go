package helper

//https://blog.csdn.net/weixin_46274168/article/details/119881396#AES_236
import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// 全零填充
func zero_padding(text []byte, block_size int) []byte {

	// 计算最后一个区块的长度
	last_block := len(text) % block_size

	// 计算填充长度
	padding_len := block_size - last_block

	// 全零填充
	padding := bytes.Repeat([]byte{0}, padding_len)

	result := append(text, padding...)

	return result
}

// 去除填充
func un_padding(encode_text []byte) []byte {

	// 去除尾部的0
	un_pad := bytes.TrimRightFunc(encode_text, func(r rune) bool {
		return r == rune(0)
	})

	return un_pad
}

// AES加密
func AES_encrypt(text, key []byte) ([]byte, error) {

	// 创建密码, 根据密码加密
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 定义大小 (16byte=128bit)
	block_size := block.BlockSize()

	// 定义偏移量
	iv := key[:block_size]

	// 填充
	text_padded := zero_padding(text, block_size)

	// 创建加密算法
	block_mode := cipher.NewCBCEncrypter(block, iv)

	// 创建空间
	encrypt := make([]byte, len(text_padded))

	// 加密
	block_mode.CryptBlocks(encrypt, text_padded)

	return encrypt, nil

}

// AES解密
func AES_decrypt(text, key []byte) ([]byte, error) {

	// 创建密码, 根据密码解密
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 定义大小 (16byte=128bit)
	block_size := block.BlockSize()

	// 定义偏移量
	iv := key[:block_size]

	// 创建加密算法
	block_mode := cipher.NewCBCDecrypter(block, iv)

	// 创建空间
	decrypt := make([]byte, len(text))

	// 解密
	block_mode.CryptBlocks(decrypt, text)

	// 去除填充
	result := un_padding(decrypt)

	return result, nil

}

// func main() {

// 	// 定义待加密文本
// 	str := "Hello World"

// 	// 定义密码(16, 24, 32)
// 	key := []byte("我是小白呀!")
// 	fmt.Println("原文本:", str, "密码:", string(key))

// 	// 加密
// 	encrypted, err := AES_encrypt([]byte(str), key)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("加密结果:", base64.StdEncoding.EncodeToString(encrypted))

// 	// 解密
// 	decrypted, err := AES_decrypt(encrypted, key)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("解密结果:", string(decrypted))

// }
