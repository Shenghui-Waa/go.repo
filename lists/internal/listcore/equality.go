package listcore

import (
	"fmt"
	"reflect"

	"github.com/Shenghui-Waa/go.repo/base"
)

// Validation applies only to default value lookup and object equality/hashing,
// never to storage, positional operations, sorting or JSON decoding.
// Scalars use Go equality. Objects must implement both contracts.
func validate(value any) error {
	if value == nil {
		return nil
	}
	_, equal := value.(base.Equaler)
	_, hash := value.(base.Hasher)
	if equal && hash {
		return nil
	}
	if equal || hash {
		return fmt.Errorf("lists: %T must implement both Equal and HashCode", value)
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		return nil
	default:
		return fmt.Errorf("lists: %T requires Equal and HashCode", value)
	}
}

func requireValues[T any](values []T) {
	for _, value := range values {
		if err := validate(value); err != nil {
			panic(err)
		}
	}
}
