package contracts

import _ "embed"

//go:embed primitive_stream.kf
var PrivateContractContent []byte

//go:embed composed_stream_template.kf
var ComposedContractContent []byte
