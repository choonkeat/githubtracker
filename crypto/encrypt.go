package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// DecryptWithSecretEnv uses `secret` uuid to decrypt plain+nonce
func DecryptWithSecretEnv(secret, ciphertext string, noncetext string) (plaintext string, err error) {
	key, err := uuid.Parse(secret)
	if err != nil {
		return "", errors.Wrapf(err, "uuid parse key")
	}
	nonce, err := uuid.Parse(noncetext)
	if err != nil {
		return "", errors.Wrapf(err, "uuid parse nonce")
	}

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", errors.Wrapf(err, "aes new cipher")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.Wrapf(err, "cipher new gcm")
	}

	cipherbytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", errors.Wrapf(err, "base64 decode")
	}

	plainbytes, err := aesgcm.Open(nil, nonce[:aesgcm.NonceSize()], cipherbytes, nil)
	if err != nil {
		return "", errors.Wrapf(err, "aesgcm open")
	}

	return string(plainbytes), nil
}

// EncryptWithSecretENV uses `secret` uuid to encrypt `plaintext`
func EncryptWithSecretENV(secret, plaintext string) (ciphertext string, noncetext string, err error) {
	key, err := uuid.Parse(secret)
	if err != nil {
		return "", "", errors.Wrapf(err, "uuid parse key")
	}
	nonce := uuid.New()

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", "", errors.Wrapf(err, "aes new cipher")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", errors.Wrapf(err, "cipher new gcm")
	}

	cipherbytes := aesgcm.Seal(nil, nonce[:aesgcm.NonceSize()], []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(cipherbytes), nonce.String(), nil
}
