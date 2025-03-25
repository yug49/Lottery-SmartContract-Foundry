package tronkey

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/smartcontractkit/chainlink/v2/core/services/keystore/internal"
	"github.com/smartcontractkit/chainlink/v2/core/utils"
)

const keyTypeIdentifier = "Tron"

// FromEncryptedJSON gets key from json and password
func FromEncryptedJSON(keyJSON []byte, password string) (Key, error) {
	return internal.FromEncryptedJSON(
		keyTypeIdentifier,
		keyJSON,
		password,
		adulteratedPassword,
		func(_ internal.EncryptedKeyExport, rawPrivKey internal.Raw) (Key, error) {
			return KeyFor(rawPrivKey), nil
		},
	)
}

// ToEncryptedJSON returns encrypted JSON representing key
func (key Key) ToEncryptedJSON(password string, scryptParams utils.ScryptParams) (export []byte, err error) {
	return internal.ToEncryptedJSON(
		keyTypeIdentifier,
		key,
		password,
		scryptParams,
		adulteratedPassword,
		func(id string, key Key, cryptoJSON keystore.CryptoJSON) internal.EncryptedKeyExport {
			return internal.EncryptedKeyExport{
				KeyType:   id,
				PublicKey: key.PublicKeyStr(),
				Crypto:    cryptoJSON,
			}
		},
	)
}

func adulteratedPassword(password string) string {
	return "tronkey" + password
}
