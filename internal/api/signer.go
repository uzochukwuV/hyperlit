package api

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type Signer struct {
	privateKey *ecdsa.PrivateKey
}

func NewSigner(privateKeyHex string) (*Signer, error) {
	if privateKeyHex == "" {
		return nil, fmt.Errorf("private key is required")
	}

	// Remove 0x prefix if present
	if privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}

	return &Signer{privateKey: privateKey}, nil
}

func (s *Signer) SignAction(action interface{}, walletAddress string, nonce int64) (map[string]interface{}, error) {
	// Create the EIP-712 typed data structure for Hyperliquid
	actionBytes, err := json.Marshal(action)
	if err != nil {
		return nil, err
	}

	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"HyperliquidTransaction": []apitypes.Type{
				{Name: "action", Type: "string"},
				{Name: "nonce", Type: "uint64"},
				{Name: "chainId", Type: "uint256"},
			},
		},
		PrimaryType: "HyperliquidTransaction",
		Domain: apitypes.TypedDataDomain{
			Name:              "HyperliquidSignTransaction",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(42161), // Arbitrum Chain ID
			VerifyingContract: walletAddress,
		},
		Message: apitypes.TypedDataMessage{
			"action":  string(actionBytes),
			"nonce":   strconv.FormatInt(nonce, 10),
			"chainId": "42161",
		},
	}

	// Hash and sign the typed data
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, err
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, err
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	hash := crypto.Keccak256Hash(rawData)

	signature, err := crypto.Sign(hash.Bytes(), s.privateKey)
	if err != nil {
		return nil, err
	}

	// Adjust v value for Ethereum
	if signature[64] < 27 {
		signature[64] += 27
	}

	return map[string]interface{}{
		"r": hex.EncodeToString(signature[0:32]),
		"s": hex.EncodeToString(signature[32:64]),
		"v": int(signature[64]),
	}, nil
}


func (s *Signer) GetAddress() string {
	publicKey := s.privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return ""
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address.Hex()
}

// GenerateNonce generates a unique nonce for Hyperliquid
func (s *Signer) GenerateNonce() int64 {
	return time.Now().UnixMilli()
}
