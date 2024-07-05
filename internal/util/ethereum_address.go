package util

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type EthereumAddress struct {
	correctlyCreated bool
	address          string
}

func NewEthereumAddress(address string) (EthereumAddress, error) {
	ethereumAddress := EthereumAddress{
		correctlyCreated: true,
		address:          address,
	}

	if err := ethereumAddress.Check(); err != nil {
		return EthereumAddress{}, err
	}

	return ethereumAddress, nil
}

// Unsafe_NewEthereumAddress the difference is that it panics on errors
func Unsafe_NewEthereumAddress(address string) EthereumAddress {
	e, err := NewEthereumAddress(address)
	if err != nil {
		panic(err)
	}
	return e
}

func (e EthereumAddress) Check() error {
	if e.address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	if !regexp.MustCompile("^0x[a-fA-F0-9]{40}$").MatchString(e.address) {
		return fmt.Errorf("address is not valid")
	}

	return nil
}

func (e EthereumAddress) Address() string {
	if !e.correctlyCreated {
		panic("please create an EthereumAddress with NewEthereumAddress")
	}
	return e.address
}

// implement JSON marshall and unmarshall as simple string

// MarshalJSON implements the json.Marshaler interface
func (e EthereumAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.address)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (e *EthereumAddress) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &e.address); err != nil {
		return err
	}

	// verify when decoding
	if err := e.Check(); err != nil {
		return err
	}

	e.correctlyCreated = true
	return nil
}
