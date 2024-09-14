package env

import (
	"fmt"
	"os"
	"strconv"
)

// this is env cache is stored, to not do os.Getenv() everytime
// and to do strict typing
var (
	CAppName        string
	CAppRootUrl     string
	CServerAddress  string
	CProductionMode bool
	CPostgresURI    string
	CJwtSignature   []byte
	CSMTPHost       string
	CSMTPPort       uint16
	CSMTPFrom       string
	CSMTPPass       string
)

func getEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("env var %s not found", key))
	}
	return v
}

func LoadString(key string) string {
	return getEnv(key)
}

// loads an unsigned integer (positive only), parsed using
//
//	strconv.ParseUint(val, 10, 16)
func LoadUint16(key string) uint16 {
	vStr := getEnv(key)
	v, err := strconv.ParseUint(vStr, 10, 16)
	if err != nil {
		panic(fmt.Sprintf("env var %s not valid: %s", key, err))
	}
	return uint16(v)
}

// accepts 0, 1, true, false or any type of capitalizations, parsed using
//
//	strconv.ParseBool()
func LoadBool(key string) bool {
	vStr := getEnv(key)
	v, err := strconv.ParseBool(vStr)
	if err != nil {
		panic(fmt.Sprintf("env var %s not valid: %s", key, err))
	}
	return v
}

func LoadByteSlice(key string) []byte {
	vStr := getEnv(key)
	return []byte(vStr)
}
