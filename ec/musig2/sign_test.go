// Copyright 2013-2022 The btcsuite developers

package musig2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"realy.lol/ec"
	"realy.lol/ec/secp256k1"
	"realy.lol/hex"
)

const (
	signVerifyTestVectorFileName = "sign_verify_vectors.json"
	sigCombineTestVectorFileName = "sig_agg_vectors.json"
)

type signVerifyValidCase struct {
	Indices       []no `json:"key_indices"`
	NonceIndices  []no `json:"nonce_indices"`
	AggNonceIndex no   `json:"aggnonce_index"`
	MsgIndex      no   `json:"msg_index"`
	SignerIndex   no   `json:"signer_index"`
	Expected      st   `json:"expected"`
}

type signErrorCase struct {
	Indices       []no `json:"key_indices"`
	AggNonceIndex no   `json:"aggnonce_index"`
	MsgIndex      no   `json:"msg_index"`
	SecNonceIndex no   `json:"secnonce_index"`
	Comment       st   `json:"comment"`
}

type verifyFailCase struct {
	Sig          st   `json:"sig"`
	Indices      []no `json:"key_indices"`
	NonceIndices []no `json:"nonce_indices"`
	MsgIndex     no   `json:"msg_index"`
	SignerIndex  no   `json:"signer_index"`
	Comment      st   `json:"comment"`
}

type verifyErrorCase struct {
	Sig          st   `json:"sig"`
	Indices      []no `json:"key_indices"`
	NonceIndices []no `json:"nonce_indices"`
	MsgIndex     no   `json:"msg_index"`
	SignerIndex  no   `json:"signer_index"`
	Comment      st   `json:"comment"`
}

type signVerifyTestVectors struct {
	SecKey           st                    `json:"sk"`
	PubKeys          []st                  `json:"pubkeys"`
	PrivNonces       []st                  `json:"secnonces"`
	PubNonces        []st                  `json:"pnonces"`
	AggNonces        []st                  `json:"aggnonces"`
	Msgs             []st                  `json:"msgs"`
	ValidCases       []signVerifyValidCase `json:"valid_test_cases"`
	SignErrorCases   []signErrorCase       `json:"sign_error_test_cases"`
	VerifyFailCases  []verifyFailCase      `json:"verify_fail_test_cases"`
	VerifyErrorCases []verifyErrorCase     `json:"verify_error_test_cases"`
}

// TestMusig2SignVerify tests that we pass the musig2 verification tests.
func TestMusig2SignVerify(t *testing.T) {
	t.Parallel()
	testVectorPath := path.Join(
		testVectorBaseDir, signVerifyTestVectorFileName,
	)
	testVectorBytes, err := os.ReadFile(testVectorPath)
	require.NoError(t, err)
	var testCases signVerifyTestVectors
	require.NoError(t, json.Unmarshal(testVectorBytes, &testCases))
	privKey, _ := btcec.SecKeyFromBytes(mustParseHex(testCases.SecKey))
	for i, testCase := range testCases.ValidCases {
		testCase := testCase
		testName := fmt.Sprintf("valid_case_%v", i)
		t.Run(testName, func(t *testing.T) {
			pubKeys, err := keysFromIndices(
				t, testCase.Indices, testCases.PubKeys,
			)
			require.NoError(t, err)
			pubNonces := pubNoncesFromIndices(
				t, testCase.NonceIndices, testCases.PubNonces,
			)
			combinedNonce, err := AggregateNonces(pubNonces)
			require.NoError(t, err)
			var msg [32]byte
			copy(msg[:], mustParseHex(testCases.Msgs[testCase.MsgIndex]))
			var secNonce [SecNonceSize]byte
			copy(secNonce[:], mustParseHex(testCases.PrivNonces[0]))
			partialSig, err := Sign(
				secNonce, privKey, combinedNonce, pubKeys,
				msg,
			)
			var partialSigBytes [32]byte
			partialSig.S.PutBytesUnchecked(partialSigBytes[:])
			require.Equal(
				t, hex.Enc(partialSigBytes[:]),
				hex.Enc(mustParseHex(testCase.Expected)),
			)
		})
	}
	for _, testCase := range testCases.SignErrorCases {
		testCase := testCase
		testName := fmt.Sprintf("invalid_case_%v",
			strings.ToLower(testCase.Comment))
		t.Run(testName, func(t *testing.T) {
			pubKeys, err := keysFromIndices(
				t, testCase.Indices, testCases.PubKeys,
			)
			if err != nil {
				require.ErrorIs(t, err, secp256k1.ErrPubKeyNotOnCurve)
				return
			}
			var aggNonce [PubNonceSize]byte
			copy(
				aggNonce[:],
				mustParseHex(
					testCases.AggNonces[testCase.AggNonceIndex],
				),
			)
			var msg [32]byte
			copy(msg[:], mustParseHex(testCases.Msgs[testCase.MsgIndex]))
			var secNonce [SecNonceSize]byte
			copy(
				secNonce[:],
				mustParseHex(
					testCases.PrivNonces[testCase.SecNonceIndex],
				),
			)
			_, err = Sign(
				secNonce, privKey, aggNonce, pubKeys,
				msg,
			)
			require.Error(t, err)
		})
	}
	for _, testCase := range testCases.VerifyFailCases {
		testCase := testCase
		testName := fmt.Sprintf("verify_fail_%v",
			strings.ToLower(testCase.Comment))
		t.Run(testName, func(t *testing.T) {
			pubKeys, err := keysFromIndices(
				t, testCase.Indices, testCases.PubKeys,
			)
			require.NoError(t, err)
			pubNonces := pubNoncesFromIndices(
				t, testCase.NonceIndices, testCases.PubNonces,
			)
			combinedNonce, err := AggregateNonces(pubNonces)
			require.NoError(t, err)
			var msg [32]byte
			copy(
				msg[:],
				mustParseHex(testCases.Msgs[testCase.MsgIndex]),
			)
			var secNonce [SecNonceSize]byte
			copy(secNonce[:], mustParseHex(testCases.PrivNonces[0]))
			signerNonce := secNonceToPubNonce(secNonce)
			var partialSig PartialSignature
			err = partialSig.Decode(
				bytes.NewReader(mustParseHex(testCase.Sig)),
			)
			if err != nil && strings.Contains(testCase.Comment, "group size") {
				require.ErrorIs(t, err, ErrPartialSigInvalid)
			}
			err = verifyPartialSig(
				&partialSig, signerNonce, combinedNonce,
				pubKeys, privKey.PubKey().SerializeCompressed(),
				msg,
			)
			require.Error(t, err)
		})
	}

	for _, testCase := range testCases.VerifyErrorCases {
		testCase := testCase
		testName := fmt.Sprintf("verify_error_%v",
			strings.ToLower(testCase.Comment))
		t.Run(testName, func(t *testing.T) {
			switch testCase.Comment {
			case "Invalid pubnonce":
				pubNonces := pubNoncesFromIndices(
					t, testCase.NonceIndices, testCases.PubNonces,
				)
				_, err := AggregateNonces(pubNonces)
				require.ErrorIs(t, err, secp256k1.ErrPubKeyNotOnCurve)

			case "Invalid pubkey":
				_, err := keysFromIndices(
					t, testCase.Indices, testCases.PubKeys,
				)
				require.ErrorIs(t, err, secp256k1.ErrPubKeyNotOnCurve)

			default:
				t.Fatalf("unhandled case: %v", testCase.Comment)
			}
		})
	}

}

type sigCombineValidCase struct {
	AggNonce     st   `json:"aggnonce"`
	NonceIndices []no `json:"nonce_indices"`
	Indices      []no `json:"key_indices"`
	TweakIndices []no `json:"tweak_indices"`
	IsXOnly      []bo `json:"is_xonly"`
	PSigIndices  []no `json:"psig_indices"`
	Expected     st   `json:"expected"`
}

type sigCombineTestVectors struct {
	PubKeys    []st                  `json:"pubkeys"`
	PubNonces  []st                  `json:"pnonces"`
	Tweaks     []st                  `json:"tweaks"`
	Psigs      []st                  `json:"psigs"`
	Msg        st                    `json:"msg"`
	ValidCases []sigCombineValidCase `json:"valid_test_cases"`
}

func pSigsFromIndicies(t *testing.T, sigs []st,
	indices []no) []*PartialSignature {
	pSigs := make([]*PartialSignature, len(indices))
	for i, idx := range indices {
		var pSig PartialSignature
		err := pSig.Decode(bytes.NewReader(mustParseHex(sigs[idx])))
		require.NoError(t, err)
		pSigs[i] = &pSig
	}
	return pSigs
}

// TestMusig2SignCombine tests that we pass the musig2 sig combination tests.
func TestMusig2SignCombine(t *testing.T) {
	t.Parallel()
	testVectorPath := path.Join(
		testVectorBaseDir, sigCombineTestVectorFileName,
	)
	testVectorBytes, err := os.ReadFile(testVectorPath)
	require.NoError(t, err)
	var testCases sigCombineTestVectors
	require.NoError(t, json.Unmarshal(testVectorBytes, &testCases))
	var msg [32]byte
	copy(msg[:], mustParseHex(testCases.Msg))
	for i, testCase := range testCases.ValidCases {
		testCase := testCase
		testName := fmt.Sprintf("valid_case_%v", i)
		t.Run(testName, func(t *testing.T) {
			pubKeys, err := keysFromIndices(
				t, testCase.Indices, testCases.PubKeys,
			)
			require.NoError(t, err)
			pubNonces := pubNoncesFromIndices(
				t, testCase.NonceIndices, testCases.PubNonces,
			)
			partialSigs := pSigsFromIndicies(
				t, testCases.Psigs, testCase.PSigIndices,
			)
			var (
				combineOpts []CombineOption
				keyOpts     []KeyAggOption
			)
			if len(testCase.TweakIndices) > 0 {
				tweaks := tweaksFromIndices(
					t, testCase.TweakIndices,
					testCases.Tweaks, testCase.IsXOnly,
				)
				combineOpts = append(combineOpts, WithTweakedCombine(
					msg, pubKeys, tweaks, false,
				))
				keyOpts = append(keyOpts, WithKeyTweaks(tweaks...))
			}
			combinedKey, _, _, err := AggregateKeys(
				pubKeys, false, keyOpts...,
			)
			require.NoError(t, err)
			combinedNonce, err := AggregateNonces(pubNonces)
			require.NoError(t, err)
			finalNonceJ, _, err := computeSigningNonce(
				combinedNonce, combinedKey.FinalKey, msg,
			)
			finalNonceJ.ToAffine()
			finalNonce := btcec.NewPublicKey(
				&finalNonceJ.X, &finalNonceJ.Y,
			)
			combinedSig := CombineSigs(
				finalNonce, partialSigs, combineOpts...,
			)
			require.Equal(t,
				strings.ToLower(testCase.Expected),
				hex.Enc(combinedSig.Serialize()),
			)
		})
	}
}
