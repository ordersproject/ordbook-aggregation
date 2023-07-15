package order_brc20_service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

type addressToKey struct {
	key        *btcec.PrivateKey
	compressed bool
}

func mkGetKey(keys map[string]addressToKey) txscript.KeyDB {
	if keys == nil {
		return txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey,
			bool, error) {
			return nil, false, errors.New("nope")
		})
	}
	return txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey,
		bool, error) {
		a2k, ok := keys[addr.EncodeAddress()]
		if !ok {
			return nil, false, errors.New("nope")
		}
		return a2k.key, a2k.compressed, nil
	})
}

func mkGetScript(scripts map[string][]byte) txscript.ScriptDB {
	if scripts == nil {
		return txscript.ScriptClosure(func(addr btcutil.Address) ([]byte, error) {
			return nil, errors.New("nope")
		})
	}
	return txscript.ScriptClosure(func(addr btcutil.Address) ([]byte, error) {
		script, ok := scripts[addr.EncodeAddress()]
		if !ok {
			return nil, errors.New("nope")
		}
		return script, nil
	})
}

func createMultiSigAddress(net *chaincfg.Params, pubKey ...string) (string, error) {
	var (
		pubKeys = make([]*btcutil.AddressPubKey, 0)
	)
	for _, v := range pubKey {
		pubByte, err := hex.DecodeString(v)
		if err != nil {
			return "", err
		}
		pub, err := btcutil.NewAddressPubKey(pubByte, net)
		if err != nil {
			return "", err
		}
		pubKeys = append(pubKeys, pub)
	}

	requiredSigs := 2

	// 构建多签脚本
	script, err := txscript.MultiSigScript(pubKeys, requiredSigs)
	if err != nil {
		fmt.Println("Failed to create multi-sig script:", err)
		return "", err
	}

	// 从脚本中获取多签地址
	address, err := btcutil.NewAddressScriptHash(script, net)
	if err != nil {
		fmt.Println("Failed to create address:", err)
		return "", err
	}

	fmt.Println("Multi-Sig Address:", address)
	return address.EncodeAddress(), nil
}

func signMultiSigScript(net *chaincfg.Params, tx *wire.MsgTx, i int, pkScript []byte, hashType txscript.SigHashType, priKey string, preSigScript []byte) ([]byte, error) {
	privateKeyBytes, err := hex.DecodeString(priKey)
	if err != nil {
		return nil, err
	}
	privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

	publicKey := privateKey.PubKey().SerializeCompressed()
	address, err := btcutil.NewAddressPubKey(publicKey, net)
	if err != nil {
		return nil, err
	}

	scriptAddr, err := btcutil.NewAddressScriptHash(pkScript, net)
	if err != nil {
		return nil, err
	}

	// Sign with the other key and merge
	sigScript, err := txscript.SignTxOutput(&chaincfg.TestNet3Params,
		tx, i, pkScript, hashType,
		mkGetKey(map[string]addressToKey{
			address.EncodeAddress(): {privateKey, true},
		}), mkGetScript(map[string][]byte{
			scriptAddr.EncodeAddress(): pkScript,
		}), preSigScript)
	return sigScript, err
}
