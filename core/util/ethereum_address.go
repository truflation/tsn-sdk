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
	hex              string
}

func NewEthereumAddressFromString(address string) (EthereumAddress, error) {
	hexAddress := strings.ToLower(address)
	// check if it has 0x prefix, normalize otherwise
	if !strings.HasPrefix(hexAddress, "0x") {
		hexAddress = "0x" + hexAddress
	}

	ethereumAddress := EthereumAddress{
		correctlyCreated: true,
		hex:              hexAddress,
	}

	if err := ethereumAddress.validate(); err != nil {
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

func (e *EthereumAddress) validate() error {
	if e.hex == "" {
		return fmt.Errorf("address cannot be empty")
	}

	regexStr := "^0x[a-fA-F0-9]{40}$"
	if !regexp.MustCompile(regexStr).MatchString(e.hex) {
		return fmt.Errorf("address does not match regex %s: %s", regexStr, e.hex)
	}

	return nil
}

func (e *EthereumAddress) checkCorrectlyCreated() {
	if !e.correctlyCreated {
		panic("please create an EthereumAddress with NewEthereumAddress")
	}
}

// Address returns the address as a hex string, starting with 0x
func (e *EthereumAddress) Address() string {
	e.checkCorrectlyCreated()
	return e.hex
}

// Bytes returns the address as a byte slice
func (e *EthereumAddress) Bytes() []byte {
	e.checkCorrectlyCreated()
	// decode the hex string to bytes (remove the 0x prefix first)
	bytes, err := hex.DecodeString(e.hex[2:])
	if err != nil {
		panic(err)
	}
	return bytes
}

// implement JSON marshall and unmarshall as simple string

// MarshalJSON implements the json.Marshaler interface
func (e *EthereumAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.hex)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (e *EthereumAddress) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &e.hex); err != nil {
		return err
	}

	// verify when decoding
	if err := e.validate(); err != nil {
		return err
	}

	e.correctlyCreated = true
	return nil
}
