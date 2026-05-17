package hash

import (
	contracts "fiber-starter/app/Providers/Hash/Contracts"
	"fiber-starter/app/Support/appctx"
)

// Facade returns the hashing service from the application container.
func Facade() contracts.Hasher {
	rt := appctx.App()
	if rt == nil {
		return nil
	}

	type hashProvider interface {
		HashService() contracts.Hasher
	}

	if hp, ok := rt.(hashProvider); ok {
		if service := hp.HashService(); service != nil {
			return service
		}
	}

	return nil
}

// Make creates a hash for the given value.
func Make(value string) (string, error) {
	if f := Facade(); f != nil {
		return f.Make(value)
	}
	return "", nil
}

// Check verifies the given value against a hash.
func Check(value, hashedValue string) bool {
	if f := Facade(); f != nil {
		return f.Check(value, hashedValue)
	}
	return false
}

// NeedsRehash checks if the given hash has been hashed using the default driver's options.
func NeedsRehash(hashedValue string) bool {
	if f := Facade(); f != nil {
		return f.NeedsRehash(hashedValue)
	}
	return false
}

// Info returns information about the given hashed value.
func Info(hashedValue string) map[string]interface{} {
	if f := Facade(); f != nil {
		return f.Info(hashedValue)
	}
	return map[string]interface{}{}
}

// Driver returns a specific hashing driver.
func Driver(name string) contracts.Hasher {
	if f := Facade(); f != nil {
		if m, ok := f.(*Manager); ok {
			d, _ := m.Driver(name)
			return d
		}
	}
	return nil
}
