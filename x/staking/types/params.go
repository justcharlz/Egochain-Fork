package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Staking params default values
const (
	// DefaultUnbondingTime reflects 21 days in seconds
	DefaultUnbondingTime = time.Duration(21 * 24 * 60 * 60) * time.Second

	// DefaultMaxValidators is the default maximum number of validators
	DefaultMaxValidators uint32 = 100

	// DefaultMaxEntries - default maximum entries
	DefaultMaxEntries uint32 = 7

	// DefaultHistoricalEntries - default amount of historical entries
	DefaultHistoricalEntries uint32 = 10000
)

// Parameter keys
var (
	KeyUnbondingTime     = []byte("UnbondingTime")
	KeyMaxValidators     = []byte("MaxValidators")
	KeyMaxEntries        = []byte("MaxEntries")
	KeyHistoricalEntries = []byte("HistoricalEntries")
	KeyBondDenom         = []byte("BondDenom")
	KeyMaxCommissionRate = []byte("MaxCommissionRate")
	KeyMaxCommissionChangeRate = []byte("MaxCommissionChangeRate")
)

// Default parameter values
var (
	DefaultBondDenom               = "aevmos"
	DefaultMaxCommissionRate       = sdk.NewDecWithPrec(20, 2)       // 20%
	DefaultMaxCommissionChangeRate = sdk.NewDecWithPrec(1, 2)        // 1% per day
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// Params defines the parameters for the staking module.
type Params struct {
	UnbondingTime         time.Duration `json:"unbonding_time" yaml:"unbonding_time"`
	MaxValidators         uint32        `json:"max_validators" yaml:"max_validators"`
	MaxEntries           uint32        `json:"max_entries" yaml:"max_entries"`
	HistoricalEntries    uint32        `json:"historical_entries" yaml:"historical_entries"`
	BondDenom            string        `json:"bond_denom" yaml:"bond_denom"`
	MaxCommissionRate    sdk.Dec       `json:"max_commission_rate" yaml:"max_commission_rate"`
	MaxCommissionChangeRate sdk.Dec    `json:"max_commission_change_rate" yaml:"max_commission_change_rate"`
}

// NewParams creates a new Params instance
func NewParams(
	unbondingTime time.Duration,
	maxValidators uint32,
	maxEntries uint32,
	historicalEntries uint32,
	bondDenom string,
	maxCommissionRate sdk.Dec,
	maxCommissionChangeRate sdk.Dec,
) Params {
	return Params{
		UnbondingTime:         unbondingTime,
		MaxValidators:         maxValidators,
		MaxEntries:           maxEntries,
		HistoricalEntries:    historicalEntries,
		BondDenom:            bondDenom,
		MaxCommissionRate:    maxCommissionRate,
		MaxCommissionChangeRate: maxCommissionChangeRate,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultUnbondingTime,
		DefaultMaxValidators,
		DefaultMaxEntries,
		DefaultHistoricalEntries,
		DefaultBondDenom,
		DefaultMaxCommissionRate,
		DefaultMaxCommissionChangeRate,
	)
}

// String implements the stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyUnbondingTime, &p.UnbondingTime, validateUnbondingTime),
		paramtypes.NewParamSetPair(KeyMaxValidators, &p.MaxValidators, validateMaxValidators),
		paramtypes.NewParamSetPair(KeyMaxEntries, &p.MaxEntries, validateMaxEntries),
		paramtypes.NewParamSetPair(KeyHistoricalEntries, &p.HistoricalEntries, validateHistoricalEntries),
		paramtypes.NewParamSetPair(KeyBondDenom, &p.BondDenom, validateBondDenom),
		paramtypes.NewParamSetPair(KeyMaxCommissionRate, &p.MaxCommissionRate, validateMaxCommissionRate),
		paramtypes.NewParamSetPair(KeyMaxCommissionChangeRate, &p.MaxCommissionChangeRate, validateMaxCommissionChangeRate),
	}
}

// Validate performs basic validation on staking parameters.
func (p Params) Validate() error {
	if err := validateUnbondingTime(p.UnbondingTime); err != nil {
		return err
	}

	if err := validateMaxValidators(p.MaxValidators); err != nil {
		return err
	}

	if err := validateMaxEntries(p.MaxEntries); err != nil {
		return err
	}

	if err := validateHistoricalEntries(p.HistoricalEntries); err != nil {
		return err
	}

	if err := validateBondDenom(p.BondDenom); err != nil {
		return err
	}

	if err := validateMaxCommissionRate(p.MaxCommissionRate); err != nil {
		return err
	}

	if err := validateMaxCommissionChangeRate(p.MaxCommissionChangeRate); err != nil {
		return err
	}

	if p.MaxCommissionChangeRate.GT(p.MaxCommissionRate) {
		return fmt.Errorf("max commission change rate cannot be greater than max commission rate")
	}

	return nil
}

func validateUnbondingTime(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("unbonding time must be positive: %d", v)
	}

	return nil
}

func validateMaxValidators(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max validators must be positive: %d", v)
	}

	return nil
}

func validateMaxEntries(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max entries must be positive: %d", v)
	}

	return nil
}

func validateHistoricalEntries(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateBondDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("bond denom cannot be blank")
	}

	return nil
}

func validateMaxCommissionRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("max commission rate cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("max commission rate cannot be greater than 100%%: %s", v)
	}

	return nil
}

func validateMaxCommissionChangeRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("max commission change rate cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("max commission change rate cannot be greater than 100%%: %s", v)
	}

	return nil
}