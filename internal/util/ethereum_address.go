package util

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type EthereumAddress struct {
	correctlyCreated bool
	address          string
}

func NewEthereumAddressFromString(address string) (EthereumAddress, error) {
	ethereumAddress := EthereumAddress{
		correctlyCreated: true,
		address:          address,
	}

	if err := ethereumAddress.Validate(); err != nil {
		return EthereumAddress{}, err
	}

	return ethereumAddress, nil
}

func NewEthereumAddressFromBytes(address []byte) (EthereumAddress, error) {
	return NewEthereumAddressFromString(hex.EncodeToString(address))
}

// Unsafe_NewEthereumAddressFromString the difference is that it panics on errors
func Unsafe_NewEthereumAddressFromString(address string) EthereumAddress {
	e, err := NewEthereumAddressFromString(address)
	if err != nil {
		panic(err)
	}
	return e
}

func (e *EthereumAddress) Validate() error {
	if e.address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	// A common error here is including 0x. We won't allow it.
	if strings.HasPrefix(e.address, "0x") {
		return fmt.Errorf("please, do not include 0x in the address: %s", e.address)
	}

	regexStr := "^[a-fA-F0-9]{40}$"
	if !regexp.MustCompile(regexStr).MatchString(e.address) {
		return fmt.Errorf("address does not match regex %s: %s", regexStr, e.address)
	}

	return nil
}

func (e *EthereumAddress) CheckCorrectlyCreated() {
	if !e.correctlyCreated {
		panic("please create an EthereumAddress with NewEthereumAddress")
	}
}

// Address returns the address as a string
func (e *EthereumAddress) Address() string {
	e.CheckCorrectlyCreated()
	return e.address
}

// implement JSON marshall and unmarshall as simple string

// MarshalJSON implements the json.Marshaler interface
func (e *EthereumAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.address)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (e *EthereumAddress) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &e.address); err != nil {
		return err
	}

	// verify when decoding
	if err := e.Validate(); err != nil {
		return err
	}

	e.correctlyCreated = true
	return nil
}
