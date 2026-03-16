package keystone

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"time"
)

const (
	// encryptionV1 is AES-GCM with a 12-byte random nonce
	encryptionV1 byte = 0x01
)

var (
	ErrInvalidKeyLength    = errors.New("encryption key must be 16, 24, or 32 bytes")
	ErrDecryptionFailed    = errors.New("decryption failed")
	ErrDataExpired         = errors.New("encrypted data has expired")
	ErrCiphertextShort     = errors.New("ciphertext too short")
	ErrUnknownCipherFormat = errors.New("unknown cipher format version")
)

// Encryptor holds an AES encryption key for encrypting and decrypting values stored in Mixed fields
type Encryptor struct {
	key []byte
}

// NewEncryptor creates a new Encryptor with the given AES key.
// Key must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256.
func NewEncryptor(key []byte) (*Encryptor, error) {
	switch len(key) {
	case 16, 24, 32:
		k := make([]byte, len(key))
		copy(k, key)
		return &Encryptor{key: k}, nil
	default:
		return nil, ErrInvalidKeyLength
	}
}

func (enc *Encryptor) gcm() (cipher.AEAD, error) {
	block, err := aes.NewCipher(enc.key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// Encrypt encrypts plaintext and returns a Mixed value.
// Text = masked, Raw = nonce + ciphertext.
func (enc *Encryptor) Encrypt(plaintext, masked string) (Mixed, error) {
	raw, err := enc.seal(plaintext, nil)
	if err != nil {
		return Mixed{}, err
	}
	m := Mixed{}
	m.SetString(masked)
	m.SetRaw(raw)
	return m, nil
}

// EncryptWithTTL encrypts plaintext with a TTL expiry and returns a Mixed value.
// The TTL is stored in the Time field and included as authenticated data,
// so the correct TTL is required for decryption.
func (enc *Encryptor) EncryptWithTTL(plaintext, masked string, ttl time.Time) (Mixed, error) {
	aad := ttlAAD(ttl)
	raw, err := enc.seal(plaintext, aad)
	if err != nil {
		return Mixed{}, err
	}
	m := Mixed{}
	m.SetString(masked)
	m.SetRaw(raw)
	m.SetTime(ttl.UTC())
	return m, nil
}

// Decrypt decrypts a Mixed value and returns the plaintext string.
// If the Mixed has a Time value set (TTL), it is validated against expiry
// and included as authenticated data for decryption.
func (enc *Encryptor) Decrypt(m Mixed) (string, error) {
	raw := m.Raw()
	if len(raw) == 0 {
		return "", nil
	}

	ttl := m.Time()
	if !ttl.IsZero() && time.Now().After(ttl) {
		return "", ErrDataExpired
	}

	var aad []byte
	if !ttl.IsZero() {
		aad = ttlAAD(ttl)
	}

	return enc.open(raw, aad)
}

// seal encrypts plaintext and returns versioned ciphertext: [version:1][nonce:12][ciphertext+tag:N]
func (enc *Encryptor) seal(plaintext string, aad []byte) ([]byte, error) {
	gcm, err := enc.gcm()
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// prepend version byte, then nonce, then sealed ciphertext
	out := make([]byte, 1, 1+gcm.NonceSize()+len(plaintext)+gcm.Overhead())
	out[0] = encryptionV1
	out = append(out, nonce...)
	out = gcm.Seal(out, nonce, []byte(plaintext), aad)
	return out, nil
}

// open decrypts versioned ciphertext: [version:1][nonce:12][ciphertext+tag:N]
func (enc *Encryptor) open(raw, aad []byte) (string, error) {
	if len(raw) < 1 {
		return "", ErrCiphertextShort
	}

	version := raw[0]
	payload := raw[1:]

	switch version {
	case encryptionV1:
		return enc.openV1(payload, aad)
	default:
		return "", ErrUnknownCipherFormat
	}
}

func (enc *Encryptor) openV1(payload, aad []byte) (string, error) {
	gcm, err := enc.gcm()
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(payload) < nonceSize {
		return "", ErrCiphertextShort
	}

	nonce, ciphertext := payload[:nonceSize], payload[nonceSize:]

	plainBytes, err := gcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plainBytes), nil
}

func ttlAAD(ttl time.Time) []byte {
	aad := make([]byte, 8)
	binary.BigEndian.PutUint64(aad, uint64(ttl.UTC().UnixMilli()))
	return aad
}
