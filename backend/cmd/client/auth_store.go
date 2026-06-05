package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

const appName = "music-test"

type authData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func configDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appName), nil
}

func getOrCreateKey() ([]byte, error) {
	cdir, err := configDir()
	if err != nil {
		return nil, err
	}
	keyFile := filepath.Join(cdir, "auth.key")
	if data, err := os.ReadFile(keyFile); err == nil {
		key, err := base64.StdEncoding.DecodeString(string(data))
		if err == nil && len(key) == 32 {
			return key, nil
		}
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cdir, 0700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(keyFile, []byte(base64.StdEncoding.EncodeToString(key)), 0600); err != nil {
		return nil, err
	}
	return key, nil
}

func encrypt(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(cipherB64 string, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aead.Open(nil, nonce, ciphertext, nil)
}

func saveCredentials(email, password string) {
	key, err := getOrCreateKey()
	if err != nil {
		return
	}
	data, _ := json.Marshal(authData{Email: email, Password: password})
	encrypted, err := encrypt(data, key)
	if err != nil {
		return
	}
	cdir, _ := configDir()
	credFile := filepath.Join(cdir, "credentials")
	_ = os.MkdirAll(cdir, 0700)
	_ = os.WriteFile(credFile, []byte(encrypted), 0600)
}

func loadCredentials() (email, password string) {
	key, err := getOrCreateKey()
	if err != nil {
		return "", ""
	}
	cdir, _ := configDir()
	credFile := filepath.Join(cdir, "credentials")
	data, err := os.ReadFile(credFile)
	if err != nil {
		return "", ""
	}
	decrypted, err := decrypt(string(data), key)
	if err != nil {
		return "", ""
	}
	var auth authData
	if err := json.Unmarshal(decrypted, &auth); err != nil {
		return "", ""
	}
	return auth.Email, auth.Password
}

func clearCredentials() {
	cdir, err := configDir()
	if err != nil {
		return
	}
	credFile := filepath.Join(cdir, "credentials")
	_ = os.Remove(credFile)
}
