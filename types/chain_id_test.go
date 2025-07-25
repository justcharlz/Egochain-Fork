package types

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseChainID(t *testing.T) {
	testCases := []struct {
		name     string
		chainID  string
		expError bool
		expInt   *big.Int
	}{
		{
			"valid chain-id, single digit", "evmos_1-1", false, big.NewInt(1),
		},
		{
			"valid chain-id, multiple digits", "aragonchain_256-1", false, big.NewInt(256),
		},
		{
			"invalid chain-id, double dash", "aragonchain-1-1", true, nil,
		},
		{
			"invalid chain-id, double underscore", "aragonchain_1_1", true, nil,
		},
		{
			"invalid chain-id, dash only", "-", true, nil,
		},
		{
			"invalid chain-id, undefined identifier and EIP155", "-1", true, nil,
		},
		{
			"invalid chain-id, undefined identifier", "_1-1", true, nil,
		},
		{
			"invalid chain-id, uppercases", "EVMOS_1-1", true, nil,
		},
		{
			"invalid chain-id, mixed cases", "Evmos_1-1", true, nil,
		},
		{
			"invalid chain-id, special chars", "$&*#!_1-1", true, nil,
		},
		{
			"invalid eip155 chain-id, cannot start with 0", "dhives_001-1", true, nil,
		},
		{
			"invalid eip155 chain-id, cannot invalid base", "dhives_0x212-1", true, nil,
		},
		{
			"invalid eip155 chain-id, non-integer", "dhives_dhives_5438-1", true, nil,
		},
		{
			"invalid epoch, undefined", "dhives_-", true, nil,
		},
		{
			"blank chain ID", " ", true, nil,
		},
		{
			"empty chain ID", "", true, nil,
		},
		{
			"empty content for chain id, eip155 and epoch numbers", "_-", true, nil,
		},
		{
			"long chain-id", "evmos_" + strings.Repeat("1", 45) + "-1", true, nil,
		},
	}

	for _, tc := range testCases {
		chainIDEpoch, err := ParseChainID(tc.chainID)
		if tc.expError {
			require.Error(t, err, tc.name)
			require.Nil(t, chainIDEpoch)

			require.False(t, IsValidChainID(tc.chainID), tc.name)
		} else {
			require.NoError(t, err, tc.name)
			require.Equal(t, tc.expInt, chainIDEpoch, tc.name)
			require.True(t, IsValidChainID(tc.chainID))
		}
	}
}
