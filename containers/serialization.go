package containers

import "encoding/json"

// JSONSerializer combines Go's encoding/json integration with GoDS's explicit
// method name. ToJSON and MarshalJSON must produce equivalent representations.
// Implementations define element order and how empty containers are encoded.
type JSONSerializer interface {
	json.Marshaler
	ToJSON() ([]byte, error)
}

// JSONDeserializer replaces contents with decoded values. FromJSON and
// UnmarshalJSON must have the same behavior. Decode into temporary storage and
// commit only on success, so invalid input does not partially mutate a container.
// Implementations document null handling and duplicate-key policy.
type JSONDeserializer interface {
	json.Unmarshaler
	FromJSON([]byte) error
}
