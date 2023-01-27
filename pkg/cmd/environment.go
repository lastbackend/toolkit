package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func getEnv(envVar string) (string, bool) {
	envVar = strings.TrimSpace(envVar)
	envVar = envName(envVar)
	return syscall.Getenv(envVar)
}

func getEnvAsInt(name string) (int, bool) {
	valueStr, ok := getEnv(name)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value, ok
	}
	return 0, false
}

func getEnvAsInt32(name string) (int32, bool) {
	valueStr, ok := getEnv(name)
	if value, err := strconv.ParseInt(valueStr, 10, 32); err == nil {
		return int32(value), ok
	}
	return 0, false
}

func getEnvAsBool(name string) (bool, bool) {
	valStr, ok := getEnv(name)
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val, ok
	}
	return false, false
}

func getEnvAsDuration(name string) (time.Duration, bool) {
	valStr, ok := getEnv(name)
	if val, err := strconv.ParseInt(valStr, 0, 64); err == nil {
		return time.Duration(val) * time.Millisecond, ok
	}
	return time.Duration(0) * time.Millisecond, false
}

func getEnvAsSlice(name string, sep string) ([]string, bool) {
	valStr, ok := getEnv(name)
	if valStr == "" {
		return make([]string, 0), false
	}
	val := strings.Split(valStr, sep)
	return val, ok
}

func envName(name string) string {
	return fmt.Sprintf("%s_%s", EnvPrefix, strings.Replace(strings.ToUpper(name), "-", "_", -1))
}
