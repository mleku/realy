// Copyright 2013-2022 The btcsuite developers

package musig2

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"realy.lol/hex"
)

type nonceGenTestCase struct {
	Rand     st  `json:"rand_"`
	Sk       st  `json:"sk"`
	AggPk    st  `json:"aggpk"`
	Msg      *st `json:"msg"`
	ExtraIn  st  `json:"extra_in"`
	Pk       st  `json:"pk"`
	Expected st  `json:"expected"`
}

type nonceGenTestCases struct {
	TestCases []nonceGenTestCase `json:"test_cases"`
}

const (
	nonceGenTestVectorsFileName = "nonce_gen_vectors.json"
	nonceAggTestVectorsFileName = "nonce_agg_vectors.json"
)

// TestMusig2NonceGenTestVectors tests the nonce generation function with the
// testvectors defined in the Musig2 BIP.
func TestMusig2NonceGenTestVectors(t *testing.T) {
	t.Parallel()
	testVectorPath := path.Join(
		testVectorBaseDir, nonceGenTestVectorsFileName,
	)
	testVectorBytes, err := os.ReadFile(testVectorPath)
	require.NoError(t, err)
	var testCases nonceGenTestCases
	require.NoError(t, json.Unmarshal(testVectorBytes, &testCases))
	for i, testCase := range testCases.TestCases {
		testCase := testCase
		customOpts := nonceGenOpts{
			randReader:  &memsetRandReader{i: 0},
			secretKey:   mustParseHex(testCase.Sk),
			combinedKey: mustParseHex(testCase.AggPk),
			auxInput:    mustParseHex(testCase.ExtraIn),
			publicKey:   mustParseHex(testCase.Pk),
		}
		if testCase.Msg != nil {
			customOpts.msg = mustParseHex(*testCase.Msg)
		}
		t.Run(fmt.Sprintf("test_case=%v", i), func(t *testing.T) {
			nonce, err := GenNonces(withCustomOptions(customOpts))
			if err != nil {
				t.Fatalf("err gen nonce aux bytes %v", err)
			}
			expectedBytes, _ := hex.Dec(testCase.Expected)
			if !equals(nonce.SecNonce[:], expectedBytes) {
				t.Fatalf("nonces don't match: expected %x, got %x",
					expectedBytes, nonce.SecNonce[:])
			}
		})
	}
}

type nonceAggError struct {
	Type    st `json:"type"`
	Signer  no `json:"signer"`
	Contrib st `json:"contrib"`
}

type nonceAggValidCase struct {
	Indices  []no `json:"pnonce_indices"`
	Expected st   `json:"expected"`
	Comment  st   `json:"comment"`
}

type nonceAggInvalidCase struct {
	Indices     []no          `json:"pnonce_indices"`
	Error       nonceAggError `json:"error"`
	Comment     st            `json:"comment"`
	ExpectedErr st            `json:"btcec_err"`
}

type nonceAggTestCases struct {
	Nonces       []st                  `json:"pnonces"`
	ValidCases   []nonceAggValidCase   `json:"valid_test_cases"`
	InvalidCases []nonceAggInvalidCase `json:"error_test_cases"`
}

// TestMusig2AggregateNoncesTestVectors tests that the musig2 implementation
// passes the nonce aggregration test vectors for musig2 1.0.
func TestMusig2AggregateNoncesTestVectors(t *testing.T) {
	t.Parallel()
	testVectorPath := path.Join(
		testVectorBaseDir, nonceAggTestVectorsFileName,
	)
	testVectorBytes, err := os.ReadFile(testVectorPath)
	require.NoError(t, err)
	var testCases nonceAggTestCases
	require.NoError(t, json.Unmarshal(testVectorBytes, &testCases))
	nonces := make([][PubNonceSize]byte, len(testCases.Nonces))
	for i := range testCases.Nonces {
		var nonce [PubNonceSize]byte
		copy(nonce[:], mustParseHex(testCases.Nonces[i]))
		nonces[i] = nonce
	}
	for i, testCase := range testCases.ValidCases {
		testCase := testCase
		var testNonces [][PubNonceSize]byte
		for _, idx := range testCase.Indices {
			testNonces = append(testNonces, nonces[idx])
		}
		t.Run(fmt.Sprintf("valid_case=%v", i), func(t *testing.T) {
			aggregatedNonce, err := AggregateNonces(testNonces)
			require.NoError(t, err)
			var expectedNonce [PubNonceSize]byte
			copy(expectedNonce[:], mustParseHex(testCase.Expected))
			require.Equal(t, aggregatedNonce[:], expectedNonce[:])
		})
	}

	for i, testCase := range testCases.InvalidCases {
		var testNonces [][PubNonceSize]byte
		for _, idx := range testCase.Indices {
			testNonces = append(testNonces, nonces[idx])
		}
		t.Run(fmt.Sprintf("invalid_case=%v", i), func(t *testing.T) {
			_, err := AggregateNonces(testNonces)
			require.True(t, err != nil)
			require.Equal(t, testCase.ExpectedErr, err.Error())
		})
	}
}
