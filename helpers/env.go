// helpers/env.go
package helpers

import "os"

// OverrideFromEnv checks if the given environment variable is set,
// and if it is, sets the provided pointer to that value.
func OverrideFromEnv(ptr *string, envVar string) {
    if envVal := os.Getenv(envVar); envVal != "" {
        *ptr = envVal
    }
}
