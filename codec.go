package zeno

// EncoderFunc defines a function signature used for encoding a Go value into a specific format,
// such as JSON, XML, or other content types. It takes a value of any type and returns the
// encoded byte slice or an error if encoding fails.
//
// Common usage includes setting a default encoder for the response payload.
//
// Example:
//
//	jsonEncoder := func(v any) ([]byte, error) {
//	    return json.Marshal(v)
//	}
type EncoderFunc func(v any) ([]byte, error)

// DecoderFunc defines a function signature used for decoding a byte slice (such as an HTTP request body)
// into a given Go struct. It takes the raw data and a pointer to the target value and returns an error
// if decoding fails.
//
// Common usage includes binding request bodies to structs during JSON, XML, or form parsing.
//
// Example:
//
//	jsonDecoder := func(data []byte, v any) error {
//	    return json.Unmarshal(data, v)
//	}
type DecoderFunc func(data []byte, v any) error

// IndentFunc defines a function signature used for applying indentation settings to a value,
// typically for pretty-printing purposes. It accepts a value, a prefix string, and an indent string,
// and is expected to modify or output a formatted version accordingly.
//
// This is often used in logging, debugging, or pretty-printing JSON/XML for readability.
// Note: This function signature suggests side-effect usage or wrapping purposes.
//
// Example:
//
//	jsonIndent := func(v any, prefix, indent string) {
//	    return json.MarshalIndent(v, prefix, indent)
//	}
type IndentFunc func(v any, prefix, indent string) ([]byte, error)
