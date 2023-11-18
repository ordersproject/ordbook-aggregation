package create_key

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

func CreateTaprootKey(netParams *chaincfg.Params) (string, string, error) {
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		return "", "", err
	}
	privateKeyHex := hex.EncodeToString(privateKey.Serialize())
	//log.Printf("new priviate key %s \n", privateKeyHex)

	publicKey := hex.EncodeToString(privateKey.PubKey().SerializeCompressed())
	_ = publicKey
	//log.Printf("new public key %s \n", publicKey)
	taprootAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootKeyNoScript(privateKey.PubKey())), netParams)
	if err != nil {
		return "", "", err
	}
	//log.Printf("new taproot address %s \n", taprootAddress.EncodeAddress())
	return privateKeyHex, taprootAddress.EncodeAddress(), nil
}

func CreateSegwitKey(netParams *chaincfg.Params) (string, string, error) {
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		return "", "", err
	}
	privateKeyHex := hex.EncodeToString(privateKey.Serialize())
	//log.Printf("new priviate key %s \n", privateKeyHex)

	publicKey := hex.EncodeToString(privateKey.PubKey().SerializeCompressed())
	//log.Printf("new public key %s \n", publicKey)
	_ = publicKey
	nativeSegwitAddress, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(privateKey.PubKey().SerializeCompressed()), netParams)
	if err != nil {
		return "", "", err
	}
	//log.Printf("new native segwit address %s \n", nativeSegwitAddress.EncodeAddress())
	return privateKeyHex, nativeSegwitAddress.EncodeAddress(), nil
}

//func CreateBlackHoleAddress(netParams *chaincfg.Params) (string, string, error) {
//	//pubKeyBytes, err := hex.DecodeString("03782f1f1736fbd1048a3b29ac9e7f5ab8c64f0c87d6a0bd671c0d6d67a3181da2")
//	pubKeyBytes, err := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000000")
//	if err != nil {
//		fmt.Println(err)
//		return "", "", err
//	}
//	pubKey, err := secp256k1.ParsePubKey(pubKeyBytes)
//	if err != nil {
//		fmt.Println(err)
//		return "", "", err
//	}
//	nativeSegwitAddress, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(pubKey.SerializeCompressed()), netParams)
//	if err != nil {
//		return "", "", err
//	}
//	return "", nativeSegwitAddress.EncodeAddress(), nil
//}
