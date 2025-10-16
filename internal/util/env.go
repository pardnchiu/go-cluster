package util

import (
	"os"
	"runtime"
	"strconv"
)

type EnvValue string

func GetEnv(key string) EnvValue {
	value := os.Getenv(key)
	return EnvValue(value)
}

func (e EnvValue) Int(value int) int {
	if string(e) == "" {
		return value
	}
	intValue, err := strconv.Atoi(string(e))
	if err != nil {
		return value
	}
	return intValue
}

func (e EnvValue) String(value string) string {
	if string(e) == "" {
		return value
	}
	return string(e)
}

func GetFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}
