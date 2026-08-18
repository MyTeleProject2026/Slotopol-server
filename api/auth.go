package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/MyTeleProject2026/Slotopol-server/config"
	"github.com/MyTeleProject2026/Slotopol-server/util"
)

const (
	// "iss" field for this tokens.
	jwtIssuer = "slotopol"

	// Pointer to User object stored at gin context
	// after successful authorization.
	userKey = "user"

	realmBasic  = `Basic realm="slotopol", charset="UTF-8"`
	realmBearer = `JWT realm="slotopol", charset="UTF-8"`
)

const (
	sqlnewprops = `INSERT INTO props (cid,uid) SELECT cid,? FROM club`
)

var (
	ErrNoJwtID  = errors.New("jwt-token does not have user id")
	ErrBadJwtID = errors.New("jwt-token id does not refer to registered user")
	ErrNoAuth   = errors.New("authorization is required")
	ErrNoScheme = errors.New("authorization does not have expected scheme")
	ErrNoSecret = errors.New("expected password or SHA256 hash on it and current time as a nonce")
	ErrSmallKey = errors.New("password too small")
	ErrNoCred   = errors.New("user with given credentials does not registered")
	ErrActivate = errors.New("activation required for this account")
	ErrOldCode  = errors.New("verification code expired")
	ErrBadCode  = errors.New("verification code does not pass")
	ErrNotPass  = errors.New("password is incorrect")
	ErrSigTime  = errors.New("signing time can not been recognized (time in RFC3339 expected)")
	ErrSigOut   = errors.New("nonce is expired")
	ErrBadHash  = errors.New("hash cannot be decoded in hexadecimal")
)

// Cfg is the global configuration reference
var Cfg = config.Cfg

// Claims of JWT-tokens. Contains additional profile identifier.
type Claims struct {
	jwt.RegisteredClaims
	UID uint64 `json:"uid,omitempty"`
}

func (c *Claims) Validate() error {
	if c.UID == 0 {
		return ErrNoJwtID
	}
	return nil
}

type AuthGetter func
