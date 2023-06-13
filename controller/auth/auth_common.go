package auth


import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func VerifySign(message, messageSign, publicKey string) (bool, error) {

	// Decode hex-encoded serialized public key.
	pubKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return false, err
	}
	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return false, err
	}

	// Decode hex-encoded serialized signature.
	sigBytes, err := hex.DecodeString(messageSign)
	if err != nil {
		return false, err
	}
	signature, err := ecdsa.ParseSignature(sigBytes)
	if err != nil {
		return false, err
	}

	// Verify the signature for the message using the public key.
	messageHash := chainhash.DoubleHashB([]byte(message))
	verified := signature.Verify(messageHash, pubKey)
	return verified, nil
}

func SignMessage(message, privateKey string) (string, error) {
	// Decode a hex-encoded private key.
	pkBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", err
	}
	privKey, _ := btcec.PrivKeyFromBytes(pkBytes)

	// Sign a message using the private key.
	messageHash := chainhash.DoubleHashB([]byte(message))
	signature := ecdsa.Sign(privKey, messageHash)

	// Serialize and display the signature.
	//fmt.Printf("Serialized Signature: %x\n", signature.Serialize())
	return hex.EncodeToString(signature.Serialize()), nil
}