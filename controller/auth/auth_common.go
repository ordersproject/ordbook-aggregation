package auth


import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// Text used to signify that a signed message follows and to prevent
// inadvertently signing a transaction.
const messageSignatureHeader = "Bitcoin Signed Message:\n"

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



func VerifyTextSign(message, messageSign, publicKey string) (bool, error) {
	sigBytes, err := base64.StdEncoding.DecodeString(messageSign)
	if err != nil {
		return false, err
	}

	var buf bytes.Buffer
	wire.WriteVarString(&buf, 0, messageSignatureHeader)
	wire.WriteVarString(&buf, 0, message)
	expectedMessageHash := chainhash.DoubleHashB(buf.Bytes())
	pk, _, err := ecdsa.RecoverCompact(sigBytes,
		expectedMessageHash)
	if err != nil {
		return false, err
	}

	fmt.Println(hex.EncodeToString(pk.SerializeCompressed()))
	return hex.EncodeToString(pk.SerializeCompressed()) == publicKey, nil
}

func SignTextMessage(message, privateKey string) (string, error) {
	// Decode a hex-encoded private key.
	pkBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", err
	}
	privKey, _ := btcec.PrivKeyFromBytes(pkBytes)

	// Sign a message using the private key.
	var buf bytes.Buffer
	wire.WriteVarString(&buf, 0, messageSignatureHeader)
	wire.WriteVarString(&buf, 0, message)
	messageHash := chainhash.DoubleHashB(buf.Bytes())

	sig, err := ecdsa.SignCompact(privKey,
		messageHash, true)
	if err != nil {
		return "", err
	}

	// Serialize and display the signature.
	//fmt.Printf("Serialized Signature: %x\n", signature.Serialize())
	return base64.StdEncoding.EncodeToString(sig), nil
}