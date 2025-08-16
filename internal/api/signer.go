package api

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"golang.org/x/crypto/sha3"
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

func (s *Signer) SignAction(action interface{}, walletAddress string, nonce int64) (string, error) {
	// Create the typed data structure for EIP-712
	typedData := s.createTypedData(action, walletAddress, nonce)

	// Encode the typed data
	domainSeparator, err := s.encodeDomain(typedData.Domain)
	if err != nil {
		return "", err
	}

	// Encode the primary type
	typeHash, err := s.encodeType(typedData.PrimaryType, typedData.Types)
	if err != nil {
		return "", err
	}

	// Encode the message
	messageBytes, err := json.Marshal(typedData.Message)
	if err != nil {
		return "", err
	}

	messageHash := s.keccak256(messageBytes)

	// Create the final hash to sign
	finalHash := s.keccak256([]byte("\x19\x01"), domainSeparator, s.keccak256(typeHash, messageHash))

	// Sign the hash
	signature, err := crypto.Sign(finalHash, s.privateKey)
	if err != nil {
		return "", err
	}

	// Adjust recovery ID for Ethereum
	if signature[64] < 27 {
		signature[64] += 27
	}

	return "0x" + hex.EncodeToString(signature), nil
}

func (s *Signer) createTypedData(action interface{}, walletAddress string, nonce int64) apitypes.TypedData {
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Agent": []apitypes.Type{
				{Name: "source", Type: "string"},
				{Name: "connectionId", Type: "bytes32"},
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
			ChainId:           math.NewHexOrDecimal256(998), // Hyperliquid chain ID
			VerifyingContract: walletAddress,
		},
		Message: apitypes.TypedDataMessage{
			"action":  s.jsonStringify(action),
			"nonce":   fmt.Sprintf("%d", nonce),
			"chainId": "998",
		},
	}
}

func (s *Signer) encodeDomain(domain apitypes.TypedDataDomain) ([]byte, error) {
	domainType := []apitypes.Type{
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
		{Name: "verifyingContract", Type: "address"},
	}

	typeHash, err := s.encodeType("EIP712Domain", map[string][]apitypes.Type{
		"EIP712Domain": domainType,
	})
	if err != nil {
		return nil, err
	}

	nameHash := s.keccak256([]byte(domain.Name))
	versionHash := s.keccak256([]byte(domain.Version))
	chainId := (*big.Int)(domain.ChainId)
	address := common.HexToAddress(domain.VerifyingContract)

	return s.keccak256(
		typeHash,
		nameHash,
		versionHash,
		math.U256Bytes(chainId),
		address.Bytes(),
	), nil
}

func (s *Signer) encodeType(primaryType string, types map[string][]apitypes.Type) ([]byte, error) {
	typeString := primaryType + "("

	typeFields := types[primaryType]
	for i, field := range typeFields {
		if i > 0 {
			typeString += ","
		}
		typeString += field.Type + " " + field.Name
	}
	typeString += ")"

	return s.keccak256([]byte(typeString)), nil
}

func (s *Signer) keccak256(data ...[]byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	for _, d := range data {
		hash.Write(d)
	}
	return hash.Sum(nil)
}

func (s *Signer) jsonStringify(v interface{}) string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
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
