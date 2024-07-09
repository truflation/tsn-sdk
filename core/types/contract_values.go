package types

import "fmt"

type StreamType string

const (
	StreamTypeComposed  StreamType = "composed"
	StreamTypePrimitive StreamType = "primitive"
)

type MetadataKey string

const (
	ReadonlyKey           MetadataKey = "readonly_key"
	StreamOwner           MetadataKey = "stream_owner"
	TypeKey               MetadataKey = "type"
	ComposeVisibilityKey  MetadataKey = "compose_visibility"
	ReadVisibilityKey     MetadataKey = "read_visibility"
	AllowReadWalletKey    MetadataKey = "allow_read_wallet"
	AllowComposeStreamKey MetadataKey = "allow_compose_stream"
)

func (s MetadataKey) GetType() MetadataType {
	switch s {
	case ReadonlyKey:
		return MetadataTypeString
	case StreamOwner:
		return MetadataTypeRef
	case TypeKey:
		return MetadataTypeString
	case ComposeVisibilityKey:
		return MetadataTypeInt
	case ReadVisibilityKey:
		return MetadataTypeInt
	case AllowReadWalletKey:
		return MetadataTypeRef
	case AllowComposeStreamKey:
		return MetadataTypeRef
	default:
		return MetadataTypeString
	}
}

func (s MetadataKey) String() string {
	return string(s)
}

type MetadataType string

const (
	MetadataTypeInt    MetadataType = "int"
	MetadataTypeBool   MetadataType = "bool"
	MetadataTypeFloat  MetadataType = "float"
	MetadataTypeString MetadataType = "string"
	MetadataTypeRef    MetadataType = "ref"
)

func (s MetadataType) StringFromValue(valueObj MetadataValue) (string, error) {
	value := valueObj.value

	switch s {
	case MetadataTypeInt:
		return fmt.Sprintf("%d", value.(int)), nil
	case MetadataTypeBool:
		return fmt.Sprintf("%t", value.(bool)), nil
	case MetadataTypeFloat:
		return fmt.Sprintf("%f", value.(float64)), nil
	case MetadataTypeString:
		return value.(string), nil
	case MetadataTypeRef:
		return value.(string), nil
	default:
		return "", fmt.Errorf("unknown metadata type: %s", s)
	}
}

type MetadataValue struct {
	// do not export this, to prevent direct access to the value
	value any
}

func NewMetadataValue[T string | int | bool | float64 | MetadataValue](value T) MetadataValue {
	return MetadataValue{value: value}
}
