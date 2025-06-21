package config

import "time"

// DatabaseURL is the env variable name for the database url
const DatabaseURL = "DATABASE_URL"

// AuthServerPort is the env variable name for the port to use for the auth server
const AuthServerPort = "AUTH_SERVER_PORT"

// SessionCookieName is the env variable name used to set the cookie for sessions
const SessionKey = "SESSION_KEY"

// SessionCookieName is the env variable name used to set the cookie for sessions
const SessionCookieName = "GOAUTH_SERVICE_SESSION_COOKIE"

const CorsAllowedOrigins = "CORS_ALLOWED_ORIGINS"

// SessionExpiration is the time in seconds when a token will expire
const SessionExpiration = 3600 * 24 * 7

// MinEntropyBits is the minimum number of bits of entropy required for a password.
const MinEntropyBits = 64

// MaxLoginAttempts is the maximum number of times user can attempt to enter the correct password
// before their account is temporarily locked
const MaxLoginAttempts = 5

// AccountLockoutLength is the time in minutes that an account will be locked
const AccountLockoutLength = 1 * time.Minute

// AccountUnlockPeriod is how often in minutes the UnlockExpiredLocks job will
// check for expired locked accounts to unlock
const AccountUnlockPeriod = 5 * time.Minute
