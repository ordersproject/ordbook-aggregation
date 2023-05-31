package create_key

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"log"
)

func CreateTaprootKey(netParams *chaincfg.Params) (string, string, error){
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		return "", "", err
	}
	privateKeyHex := hex.EncodeToString(privateKey.Serialize())
	log.Printf("new priviate key %s \n", privateKeyHex)

	publicKey := hex.EncodeToString(privateKey.PubKey().SerializeCompressed())
	log.Printf("new public key %s \n", publicKey)
	taprootAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootKeyNoScript(privateKey.PubKey())), netParams)
	if err != nil {
		return "", "", err
	}
	log.Printf("new taproot address %s \n", taprootAddress.EncodeAddress())
	return privateKeyHex, taprootAddress.EncodeAddress(), nil
}
