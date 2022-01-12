package util

import (
	"fmt"
	"os"
)

func EnvOrError(envVar string) (string, error) {
	val, defined := os.LookupEnv(envVar)
	if !defined {
		return "", fmt.Errorf("env var %s is not defined", envVar)
	}

	return val, nil
}

func EnvOrDefault(envVar, def string) string {
	val, err := EnvOrError(envVar)
	if err != nil {
		return def
	}

	return val
}

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
