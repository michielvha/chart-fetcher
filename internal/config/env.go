package config

import "os"

// OverrideFromEnv checks whether envVar is set in the environment and, if so,
// writes its value into the string pointed to by ptr.
func OverrideFromEnv(ptr *string, envVar string) {
	if envVal := os.Getenv(envVar); envVal != "" {
		*ptr = envVal
	}
}
