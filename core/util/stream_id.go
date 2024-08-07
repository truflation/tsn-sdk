package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateStreamId is the hash fn to generate a stream Id from a string
// it already prepends the "st" prefix to the hash, and returns the first 30 characters, to fit kwil's limit
func GenerateStreamId(s string) StreamId {
	hasher := sha256.New()
	hasher.Write([]byte(s))
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	streamIdStr := "st" + hashString[:30]

	streamId, _ := NewStreamId(streamIdStr)
	return *streamId
}

type StreamId struct {
	Id string
}

func NewStreamId(s string) (*StreamId, error) {
	id := StreamId{
		Id: s,
	}

	if err := id.Validate(); err != nil {
		return nil, err
	}

	return &id, nil
}

func (s *StreamId) Validate() error {
	// verify if the string is a valid stream Id
	if len(s.Id) != 32 || s.Id[:2] != "st" {
		return fmt.Errorf("invalid stream Id '%s'", s)
	}
	return nil
}

func (s *StreamId) String() string {
	return s.Id
}

func (s *StreamId) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", s.Id)), nil
}

func (s *StreamId) UnmarshalJSON(b []byte) error {
	// remove quotes
	s.Id = string(b[1 : len(b)-1])
	return nil
}

type StreamIdSlice []StreamId

func (s StreamIdSlice) Strings() []string {
	strs := make([]string, len(s))
	for i, streamId := range s {
		strs[i] = streamId.String()
	}
	return strs
}

func (s StreamIdSlice) Len() int           { return len(s) }
func (s StreamIdSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s StreamIdSlice) Less(i, j int) bool { return s[i].String() < s[j].String() }
