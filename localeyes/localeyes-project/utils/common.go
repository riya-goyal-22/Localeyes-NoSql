package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"os"
	"strings"
)

func GenerateRandomId() string {
	id := uuid.New()
	idString := base64.RawStdEncoding.EncodeToString(id[:])[:8]
	if strings.Contains(idString, "/") {
		idString = strings.Replace(idString, "/", "A", -1)
		return idString
	}
	return idString
}

func DecryptAES(cipherTextBase64 string) (string, error) {
	// Decode the base64 string into raw ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(cipherTextBase64)
	if err != nil {
		return "", err
	}

	// AES block size (AES uses 128-bit blocks, i.e., 16 bytes)
	block, err := aes.NewCipher([]byte(os.Getenv("DecryptionSecret")))
	if err != nil {
		return "", err
	}

	// The IV is the first AES.BlockSize (16 bytes) of the ciphertext
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// Create AES cipher in CBC mode
	mode := cipher.NewCBCDecrypter(block, iv)

	// Decrypt the ciphertext
	plainText := make([]byte, len(ciphertext))
	mode.CryptBlocks(plainText, ciphertext)

	// Remove padding (PKCS7 padding scheme)
	plainText = unpad(plainText)

	// Return the decrypted plaintext as a string
	return string(plainText), nil
}

func unpad(data []byte) []byte {
	// Print original data
	fmt.Println("Original Data:", string(data))

	// Get the padding length (the last byte)
	padding := int(data[len(data)-1])

	// Print padding value
	fmt.Println("Padding:", padding)

	// Check if the padding is valid
	if padding < 1 || padding > len(data) {
		// Invalid padding
		fmt.Println("Invalid padding!")
		return nil
	}

	// Return the unpadded data (slice the last 'padding' bytes)
	return data[:len(data)-padding]
}
