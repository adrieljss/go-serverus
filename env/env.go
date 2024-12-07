package env

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// this is env cache, to not do os.Getenv() everytime
// and to do strict typing
var (
	// in .env file, descriptions are in the file
	CAppName            string
	CAppRootUrl         string
	CFrontendRootUrl    string
	CServerAddress      string
	CProductionMode     bool
	CRequestTimeout     uint16
	CEnableRedisCaching bool
	CRedisCacheDuration int16
	CRedisURI           string
	CPostgresURI        string
	CJwtSignature       []byte
	CSMTPHost           string
	CSMTPPort           uint16
	CSMTPFrom           string
	CSMTPPass           string
	CGoogleClientID     string
	CGoogleClientSecret string
)

// not needed in .env
var (
	JwtTtl = time.Hour * 24 * 7 // jwt token valid for 7 days

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
		logrus.Fatal(fmt.Sprintf("env var %s not found", key))
	}
	return v
}

func LoadString(key string) string {
	return getEnv(key)
}

// Loads an unsigned integer (positive only).
// Parsed using:
//
//	strconv.ParseUint(val, 10, 16)
func LoadUint16(key string) uint16 {
	vStr := getEnv(key)
	v, err := strconv.ParseUint(vStr, 10, 16)
	if err != nil {
		logrus.Fatal(fmt.Sprintf("env var %s not valid: %s", key, err))
	}
	return uint16(v)
}

func LoadInt(key string) int16 {
	vStr := getEnv(key)
	v, err := strconv.ParseInt(vStr, 10, 16)
	if err != nil {
		logrus.Fatal(fmt.Sprintf("env var %s not valid: %s", key, err))
	}
	return int16(v)
}

// Accepts 0, 1, true, false.
// Parsed using:
//
//	strconv.ParseBool(val)
func LoadBool(key string) bool {
	vStr := getEnv(key)
	v, err := strconv.ParseBool(vStr)
	if err != nil {
		logrus.Fatal(fmt.Sprintf("env var %s not valid: %s", key, err))
	}
	return v
}

func LoadByteSlice(key string) []byte {
	vStr := getEnv(key)
	return []byte(vStr)
}
