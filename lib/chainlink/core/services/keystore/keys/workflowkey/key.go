package workflowkey

import (
	"crypto/ed25519"
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"

	"github.com/smartcontractkit/chainlink/v2/core/services/keystore/internal"
)

func KeyFor(raw internal.Raw) Key {
	privateKey := [32]byte(raw.Bytes())
	return Key{
		privateKey: &privateKey,
		publicKey:  curve25519PubKeyFromPrivateKey(privateKey),
	}
}

type Key struct {
	privateKey *[curve25519.PointSize]byte
	publicKey  *[curve25519.PointSize]byte
}

func New() (Key, error) {
	publicKey, privateKey, err := box.GenerateKey(cryptorand.Reader)
	if err != nil {
		return Key{}, err
	}

	return Key{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (k Key) PublicKey() [curve25519.PointSize]byte {
	if k.publicKey == nil {
		return [curve25519.PointSize]byte{}
	}

	return *k.publicKey
}

func (k Key) PublicKeyString() string {
	if k.publicKey == nil {
		return ""
	}

	return hex.EncodeToString(k.publicKey[:])
}

func (k Key) ID() string {
	return k.PublicKeyString()
}

func (k Key) Raw() internal.Raw {
	raw := make([]byte, curve25519.PointSize)
	copy(raw, k.privateKey[:])
	return internal.NewRaw(raw)
}

func (k Key) String() string {
	return fmt.Sprintf("%sKey{PrivateKey: <redacted>, PublicKey: %s}", keyTypeIdentifier, *k.publicKey)
}

func (k Key) GoString() string {
	return k.String()
}

// Encrypt encrypts a message using the public key
func (k Key) Encrypt(plaintext []byte) ([]byte, error) {
	publicKey := k.PublicKey()
	if publicKey == [curve25519.PointSize]byte{} {
		return nil, errors.New("public key is empty")
	}

	encrypted, err := box.SealAnonymous(nil, plaintext, &publicKey, cryptorand.Reader)
	if err != nil {
		return nil, err
	}

	return encrypted, nil
}

// Decrypt decrypts a message that was encrypted using the private key
func (k Key) Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	publicKey := k.PublicKey()
	if publicKey == [curve25519.PointSize]byte{} {
		return nil, errors.New("public key is empty")
	}

	decrypted, success := box.OpenAnonymous(nil, ciphertext, &publicKey, k.privateKey)
	if !success {
		return nil, errors.New("decryption failed")
	}

	return decrypted, nil
}

func curve25519PubKeyFromPrivateKey(privateKey [curve25519.PointSize]byte) *[curve25519.PointSize]byte {
	var publicKey [curve25519.PointSize]byte

	// Derive the public key
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	return &publicKey
}

func MustNewXXXTestingOnly(k *big.Int) Key {
	seed := make([]byte, ed25519.SeedSize)
	copy(seed, k.Bytes())
	privKey := ed25519.NewKeyFromSeed(seed)

	var privateKey [32]byte
	copy(privateKey[:], privKey.Seed())
	return Key{
		privateKey: &privateKey,
		publicKey:  curve25519PubKeyFromPrivateKey(privateKey),
	}
}
