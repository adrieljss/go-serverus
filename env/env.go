package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// this is env cache is stored, to not do os.Getenv() everytime
// and to do strict typing
var (
	// in .env file, descriptions are in the file
	CAppName            string
	CAppRootUrl         string
	CFrontendRootUrl    string
	CServerAddress      string
	CProductionMode     bool
	CPostgresURI        string
	CJwtSignature       []byte
	CSMTPHost           string
	CSMTPPort           uint16
	CSMTPFrom           string
	CSMTPPass           string
	CGoogleClientID     string
	CGoogleClientSecret string

	// not in .env

	// rate limit options
	RateLimitBucketSize                = 2               // for bucket size and frequency, read the ratelimit documentation in golang standard lib
	RateLimitFrequency                 = 5               // ^^^
	RateLimitTTLMapObliteratorInterval = time.Hour * 2   // every X minutes the map will delete expired keys (to avoid memleak)
	RateLimitInstanceTTL               = time.Minute * 5 // ratelimit TTL, if a user is not active in the last Y minutes, its ratelimit instance will be deleted

	// Email Confirmation options
	EmailConfirmationObliteratorInterval = time.Hour * 2
	EmailConfirmationTTL                 = time.Minute * 10 // user only have up to X minutes to verify their emails
	EmailResetPassObliteratorInterval    = time.Hour * 2
	EmailResetPassTTL                    = time.Minute * 10 // user only have up to X minutes to reset pass
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
