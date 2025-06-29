package zeno

import "github.com/bytedance/sonic"

// EncoderFunc defines the signature for encoding any Go value into JSON, XML, etc.
// It returns a byte slice that can be directly written to the response.
type EncoderFunc func(v any) ([]byte, error)

// DecoderFunc defines the signature for decoding request body bytes into a Go struct.
type DecoderFunc func(data []byte, v any) error

// default encoder for json
func jsonEncoder(v any) ([]byte, error) {
	return sonic.Marshal(v)
}

// default decoder for jso
func jsonDecoder(data []byte, v any) error {
	return sonic.Unmarshal(data, v)
}

// default encoder for pretty json
func jsonPrettyEncoder(v any) ([]byte, error) {
	return sonic.MarshalIndent(v, " ", " ")
}
