package keystone

import (
	"errors"
	"testing"
	"time"
)

func testEncryptor(t *testing.T) *Encryptor {
	t.Helper()
	key := []byte("0123456789abcdef0123456789abcdef") // 32-byte AES-256 key
	enc, err := NewEncryptor(key)
	if err != nil {
		t.Fatal(err)
	}
	return enc
}

func TestNewEncryptor_ValidKeys(t *testing.T) {
	for _, size := range []int{16, 24, 32} {
		key := make([]byte, size)
		enc, err := NewEncryptor(key)
		if err != nil {
			t.Errorf("key size %d: unexpected error: %v", size, err)
		}
		if enc == nil {
			t.Errorf("key size %d: expected non-nil encryptor", size)
		}
	}
}

func TestNewEncryptor_InvalidKey(t *testing.T) {
	_, err := NewEncryptor([]byte("short"))
	if !errors.Is(err, ErrInvalidKeyLength) {
		t.Errorf("expected ErrInvalidKeyLength, got %v", err)
	}
}

func TestNewEncryptor_KeyIsCopied(t *testing.T) {
	key := []byte("0123456789abcdef")
	enc, _ := NewEncryptor(key)
	key[0] = 0xFF
	if enc.key[0] == 0xFF {
		t.Error("encryptor should hold a copy of the key")
	}
}

func TestEncrypt_RoundTrip(t *testing.T) {
	enc := testEncryptor(t)
	m, err := enc.Encrypt("secret-value", "sec***")
	if err != nil {
		t.Fatal(err)
	}

	if m.String() != "sec***" {
		t.Errorf("expected masked text 'sec***', got %q", m.String())
	}
	if len(m.Raw()) == 0 {
		t.Error("expected non-empty raw ciphertext")
	}
	if !m.Time().IsZero() {
		t.Error("expected zero time when no TTL set")
	}

	plaintext, err := enc.Decrypt(m)
	if err != nil {
		t.Fatal(err)
	}
	if plaintext != "secret-value" {
		t.Errorf("expected 'secret-value', got %q", plaintext)
	}
}

func TestEncrypt_UniquePerCall(t *testing.T) {
	enc := testEncryptor(t)
	m1, _ := enc.Encrypt("same-value", "sam***")
	m2, _ := enc.Encrypt("same-value", "sam***")

	if string(m1.Raw()) == string(m2.Raw()) {
		t.Error("expected different ciphertexts for each encrypt (random nonce)")
	}
}

func TestEncryptWithTTL_RoundTrip(t *testing.T) {
	enc := testEncryptor(t)
	ttl := time.Now().Add(time.Hour).UTC().Truncate(time.Millisecond)
	m, err := enc.EncryptWithTTL("secret", "sec***", ttl)
	if err != nil {
		t.Fatal(err)
	}

	if m.Time().IsZero() {
		t.Fatal("expected time to be set")
	}

	plaintext, err := enc.Decrypt(m)
	if err != nil {
		t.Fatal(err)
	}
	if plaintext != "secret" {
		t.Errorf("expected 'secret', got %q", plaintext)
	}
}

func TestEncryptWithTTL_RequiredForDecryption(t *testing.T) {
	enc := testEncryptor(t)
	ttl := time.Now().Add(time.Hour).UTC().Truncate(time.Millisecond)
	m, _ := enc.EncryptWithTTL("secret", "sec***", ttl)

	// Remove the TTL — decryption should fail because AAD won't match
	m.SetTime(time.Time{})
	_, err := enc.Decrypt(m)
	if !errors.Is(err, ErrDecryptionFailed) {
		t.Errorf("expected ErrDecryptionFailed without correct TTL, got %v", err)
	}
}

func TestEncryptWithTTL_WrongTTLFails(t *testing.T) {
	enc := testEncryptor(t)
	ttl := time.Now().Add(time.Hour).UTC().Truncate(time.Millisecond)
	m, _ := enc.EncryptWithTTL("secret", "sec***", ttl)

	// Set a different TTL
	m.SetTime(ttl.Add(time.Second))
	_, err := enc.Decrypt(m)
	if !errors.Is(err, ErrDecryptionFailed) {
		t.Errorf("expected ErrDecryptionFailed with wrong TTL, got %v", err)
	}
}

func TestDecrypt_Expired(t *testing.T) {
	enc := testEncryptor(t)
	ttl := time.Now().Add(-time.Hour) // already expired
	m, _ := enc.EncryptWithTTL("secret", "sec***", ttl)

	_, err := enc.Decrypt(m)
	if !errors.Is(err, ErrDataExpired) {
		t.Errorf("expected ErrDataExpired, got %v", err)
	}

	// Masked value still accessible
	if m.String() != "sec***" {
		t.Errorf("expected masked 'sec***', got %q", m.String())
	}
}

func TestDecrypt_EmptyRaw(t *testing.T) {
	enc := testEncryptor(t)
	m := Mixed{}
	m.SetString("masked")

	plaintext, err := enc.Decrypt(m)
	if err != nil {
		t.Errorf("expected no error for empty raw, got %v", err)
	}
	if plaintext != "" {
		t.Errorf("expected empty plaintext, got %q", plaintext)
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	enc := testEncryptor(t)
	m, _ := enc.Encrypt("secret", "sec***")

	wrongKey, _ := NewEncryptor([]byte("different-key---different-key---"))
	_, err := wrongKey.Decrypt(m)
	if err != ErrDecryptionFailed {
		t.Errorf("expected ErrDecryptionFailed with wrong key, got %v", err)
	}
}

func TestDecrypt_ShortCiphertext(t *testing.T) {
	enc := testEncryptor(t)
	m := Mixed{}
	// Valid version byte but payload too short for nonce
	m.SetRaw([]byte{encryptionV1, 0x01, 0x02})

	_, err := enc.Decrypt(m)
	if !errors.Is(err, ErrCiphertextShort) {
		t.Errorf("expected ErrCiphertextShort, got %v", err)
	}
}

func TestDecrypt_EmptyRawBytes(t *testing.T) {
	enc := testEncryptor(t)
	m := Mixed{}
	m.SetRaw([]byte{})

	plaintext, err := enc.Decrypt(m)
	if err != nil {
		t.Errorf("expected no error for empty raw, got %v", err)
	}
	if plaintext != "" {
		t.Errorf("expected empty plaintext, got %q", plaintext)
	}
}

func TestDecrypt_UnknownVersion(t *testing.T) {
	enc := testEncryptor(t)
	m, _ := enc.Encrypt("secret", "sec***")

	// Corrupt the version byte to an unknown version
	raw := m.Raw()
	raw[0] = 0xFF
	m.SetRaw(raw)

	_, err := enc.Decrypt(m)
	if !errors.Is(err, ErrUnknownCipherFormat) {
		t.Errorf("expected ErrUnknownCipherFormat, got %v", err)
	}
}

func TestEncrypt_VersionBytePresent(t *testing.T) {
	enc := testEncryptor(t)
	m, _ := enc.Encrypt("test", "t***")

	raw := m.Raw()
	if len(raw) == 0 {
		t.Fatal("expected non-empty raw")
	}
	if raw[0] != encryptionV1 {
		t.Errorf("expected version byte %d, got %d", encryptionV1, raw[0])
	}
}

func TestEncrypt_MarshalUnmarshalRoundTrip(t *testing.T) {
	enc := testEncryptor(t)
	ttl := time.Now().Add(time.Hour).UTC().Truncate(time.Millisecond)
	m, _ := enc.EncryptWithTTL("secret", "sec***", ttl)

	// Marshal to proto
	val, err := m.MarshalValue()
	if err != nil {
		t.Fatal(err)
	}

	// Unmarshal from proto
	var restored Mixed
	if err := restored.UnmarshalValue(val); err != nil {
		t.Fatal(err)
	}

	// Decrypt the restored value
	plaintext, err := enc.Decrypt(restored)
	if err != nil {
		t.Fatal(err)
	}
	if plaintext != "secret" {
		t.Errorf("expected 'secret', got %q", plaintext)
	}
	if restored.String() != "sec***" {
		t.Errorf("expected masked 'sec***', got %q", restored.String())
	}
}
