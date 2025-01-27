package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	ParamStoreKeyCommunityTax        = []byte("CommunitytaxParam")
	ParamStoreKeyBaseProposerReward  = []byte("BaseProposerReward")
	ParamStoreKeyBonusProposerReward = []byte("BonusProposerReward")
	ParamStoreKeyWithdrawAddrEnabled = []byte("WithdrawAddrEnabled")
)

// Parameter keys for distribution module
var (
	KeyCommunityTax        = []*paramtypes.ParamDescriptor{{KeyTable: ParamStoreKeyCommunityTax, ValueType: sdk.Dec{}, DefaultValue: DefaultCommunityTax}}
	KeyBaseProposerReward  = []*paramtypes.ParamDescriptor{{KeyTable: ParamStoreKeyBaseProposerReward, ValueType: sdk.Dec{}, DefaultValue: DefaultBaseProposerReward}}
	KeyBonusProposerReward = []*paramtypes.ParamDescriptor{{KeyTable: ParamStoreKeyBonusProposerReward, ValueType: sdk.Dec{}, DefaultValue: DefaultBonusProposerReward}}
	KeyWithdrawAddrEnabled = []*paramtypes.ParamDescriptor{{KeyTable: ParamStoreKeyWithdrawAddrEnabled, ValueType: false, DefaultValue: DefaultWithdrawAddrEnabled}}
)

// Distribution parameters
var (
	DefaultCommunityTax        = sdk.NewDecWithPrec(2, 2)  // 2% to community pool
	DefaultBaseProposerReward  = sdk.NewDecWithPrec(3, 2)  // 3% to block proposer
	DefaultBonusProposerReward = sdk.NewDecWithPrec(5, 2)  // 5% to block proposer bonus
	DefaultWithdrawAddrEnabled = true
)

// Params defines the parameters for the distribution module.
type Params struct {
	CommunityTax        sdk.Dec `json:"community_tax" yaml:"community_tax"`
	BaseProposerReward  sdk.Dec `json:"base_proposer_reward" yaml:"base_proposer_reward"`
	BonusProposerReward sdk.Dec `json:"bonus_proposer_reward" yaml:"bonus_proposer_reward"`
	WithdrawAddrEnabled bool    `json:"withdraw_addr_enabled" yaml:"withdraw_addr_enabled"`
}

// ParamTable for distribution module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns default distribution parameters
func DefaultParams() Params {
	return Params{
		CommunityTax:        DefaultCommunityTax,
		BaseProposerReward:  DefaultBaseProposerReward,
		BonusProposerReward: DefaultBonusProposerReward,
		WithdrawAddrEnabled: DefaultWithdrawAddrEnabled,
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyCommunityTax, &p.CommunityTax, validateCommunityTax),
		paramtypes.NewParamSetPair(ParamStoreKeyBaseProposerReward, &p.BaseProposerReward, validateBaseProposerReward),
		paramtypes.NewParamSetPair(ParamStoreKeyBonusProposerReward, &p.BonusProposerReward, validateBonusProposerReward),
		paramtypes.NewParamSetPair(ParamStoreKeyWithdrawAddrEnabled, &p.WithdrawAddrEnabled, validateWithdrawAddrEnabled),
	}
}

// Validate performs basic validation on distribution parameters.
func (p Params) Validate() error {
	if err := validateCommunityTax(p.CommunityTax); err != nil {
		return err
	}
	if err := validateBaseProposerReward(p.BaseProposerReward); err != nil {
		return err
	}
	if err := validateBonusProposerReward(p.BonusProposerReward); err != nil {
		return err
	}
	if err := validateWithdrawAddrEnabled(p.WithdrawAddrEnabled); err != nil {
		return err
	}

	return nil
}

func validateCommunityTax(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("community tax must be positive: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("community tax too large: %s", v)
	}

	return nil
}

func validateBaseProposerReward(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("base proposer reward must be positive: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("base proposer reward too large: %s", v)
	}

	return nil
}

func validateBonusProposerReward(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("bonus proposer reward must be positive: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("bonus proposer reward too large: %s", v)
	}

	return nil
}

func validateWithdrawAddrEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func (p Params) String() string {
	out := strings.TrimSpace(fmt.Sprintf(`Distribution Params:
  Community Tax:        %s
  Base Proposer Reward: %s
  Bonus Proposer Reward:%s
  Withdraw Addr Enabled:%v`,
		p.CommunityTax,
		p.BaseProposerReward,
		p.BonusProposerReward,
		p.WithdrawAddrEnabled,
	))
	return out
}