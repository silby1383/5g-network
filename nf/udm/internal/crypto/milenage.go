package crypto

import (
	"crypto/aes"
	"encoding/hex"
	"fmt"
)

// MILENAGE implements the 3GPP MILENAGE algorithm set for 5G-AKA authentication
// Reference: 3GPP TS 35.205, TS 35.206, TS 35.207, TS 35.208

// AuthenticationVector contains the 5G authentication vector
type AuthenticationVector struct {
	RAND []byte // Random challenge (128 bits)
	AUTN []byte // Authentication token (128 bits)
	XRES []byte // Expected response (64-128 bits)
	CK   []byte // Cipher key (128 bits)
	IK   []byte // Integrity key (128 bits)
	AK   []byte // Anonymity key (48 bits)
}

// MILENAGE represents the MILENAGE algorithm implementation
type MILENAGE struct {
	// OP/OPc: Operator variant algorithm configuration field (128 bits)
	// K: Subscriber key (128 or 256 bits)
}

// ComputeOPc computes OPc from K and OP
// OPc = E[K](OP) XOR OP
func ComputeOPc(k, op []byte) ([]byte, error) {
	if len(k) != 16 {
		return nil, fmt.Errorf("K must be 128 bits (16 bytes), got %d bytes", len(k))
	}
	if len(op) != 16 {
		return nil, fmt.Errorf("OP must be 128 bits (16 bytes), got %d bytes", len(op))
	}

	block, err := aes.NewCipher(k)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	opc := make([]byte, 16)
	block.Encrypt(opc, op)
	
	// XOR with OP
	for i := 0; i < 16; i++ {
		opc[i] ^= op[i]
	}

	return opc, nil
}

// f1 computes MAC-A/MAC-S (network authentication function)
// MAC = f1(K, RAND, SQN, AMF)
func f1(k, opc, rand, sqn, amf []byte) ([]byte, error) {
	temp := make([]byte, 16)
	
	// Concatenate SQN || AMF || SQN || AMF
	for i := 0; i < 6; i++ {
		temp[i] = sqn[i]
	}
	for i := 0; i < 2; i++ {
		temp[i+6] = amf[i]
	}
	for i := 0; i < 6; i++ {
		temp[i+8] = sqn[i]
	}
	for i := 0; i < 2; i++ {
		temp[i+14] = amf[i]
	}

	// XOR with OPc
	for i := 0; i < 16; i++ {
		temp[i] ^= opc[i]
	}

	// Encrypt with K
	block, err := aes.NewCipher(k)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	in := make([]byte, 16)
	for i := 0; i < 16; i++ {
		in[i] = rand[i] ^ opc[i]
	}

	block.Encrypt(temp, in)
	
	// Rotate and XOR
	for i := 0; i < 16; i++ {
		temp[i] ^= opc[i]
	}

	mac := make([]byte, 8)
	copy(mac, temp[:8])
	
	return mac, nil
}

// f2345 computes RES, CK, IK, and AK
func f2345(k, opc, rand []byte) (res, ck, ik, ak []byte, err error) {
	block, err := aes.NewCipher(k)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Compute temporary value
	temp := make([]byte, 16)
	for i := 0; i < 16; i++ {
		temp[i] = rand[i] ^ opc[i]
	}

	out := make([]byte, 16)
	block.Encrypt(out, temp)

	// f2 - RES (response)
	res = make([]byte, 8)
	for i := 0; i < 16; i++ {
		out[i] ^= opc[i]
	}
	copy(res, out[:8])

	// f3 - CK (cipher key)
	ck = make([]byte, 16)
	temp2 := make([]byte, 16)
	for i := 0; i < 16; i++ {
		temp2[i] = rand[i] ^ opc[i]
	}
	// Rotate temp2
	temp2[15] ^= 1
	block.Encrypt(ck, temp2)
	for i := 0; i < 16; i++ {
		ck[i] ^= opc[i]
	}

	// f4 - IK (integrity key)
	ik = make([]byte, 16)
	temp3 := make([]byte, 16)
	for i := 0; i < 16; i++ {
		temp3[i] = rand[i] ^ opc[i]
	}
	// Rotate temp3
	temp3[15] ^= 2
	block.Encrypt(ik, temp3)
	for i := 0; i < 16; i++ {
		ik[i] ^= opc[i]
	}

	// f5 - AK (anonymity key)
	ak = make([]byte, 6)
	temp4 := make([]byte, 16)
	for i := 0; i < 16; i++ {
		temp4[i] = rand[i] ^ opc[i]
	}
	// Rotate temp4
	temp4[15] ^= 4
	akOut := make([]byte, 16)
	block.Encrypt(akOut, temp4)
	for i := 0; i < 16; i++ {
		akOut[i] ^= opc[i]
	}
	copy(ak, akOut[:6])

	return res, ck, ik, ak, nil
}

// GenerateAuthVector generates a 5G authentication vector using MILENAGE
func GenerateAuthVector(k, opc, rand, sqn, amf []byte) (*AuthenticationVector, error) {
	// Validate inputs
	if len(k) != 16 {
		return nil, fmt.Errorf("K must be 16 bytes, got %d", len(k))
	}
	if len(opc) != 16 {
		return nil, fmt.Errorf("OPc must be 16 bytes, got %d", len(opc))
	}
	if len(rand) != 16 {
		return nil, fmt.Errorf("RAND must be 16 bytes, got %d", len(rand))
	}
	if len(sqn) != 6 {
		return nil, fmt.Errorf("SQN must be 6 bytes, got %d", len(sqn))
	}
	if len(amf) != 2 {
		return nil, fmt.Errorf("AMF must be 2 bytes, got %d", len(amf))
	}

	// Compute MAC (f1)
	mac, err := f1(k, opc, rand, sqn, amf)
	if err != nil {
		return nil, fmt.Errorf("failed to compute MAC: %w", err)
	}

	// Compute RES, CK, IK, AK (f2, f3, f4, f5)
	res, ck, ik, ak, err := f2345(k, opc, rand)
	if err != nil {
		return nil, fmt.Errorf("failed to compute RES/CK/IK/AK: %w", err)
	}

	// Compute AUTN = (SQN ⊕ AK) || AMF || MAC
	autn := make([]byte, 16)
	for i := 0; i < 6; i++ {
		autn[i] = sqn[i] ^ ak[i]
	}
	copy(autn[6:8], amf)
	copy(autn[8:16], mac)

	return &AuthenticationVector{
		RAND: rand,
		AUTN: autn,
		XRES: res,
		CK:   ck,
		IK:   ik,
		AK:   ak,
	}, nil
}

// HexToBytes converts hex string to bytes
func HexToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// BytesToHex converts bytes to hex string
func BytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}
