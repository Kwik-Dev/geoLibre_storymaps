package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AccessTokenTTL is how long an issued access token stays valid (≈15 minutes).
const AccessTokenTTL = 15 * time.Minute

// JWT signs and verifies HS256 access tokens. Both the Bearer access token and
// the httpOnly refresh cookie carry the same compact token format, so a single
// verifier is used for both paths (the refresh cookie is simply a token stored
// in an httpOnly cookie rather than in the client's JS-accessible storage).
type JWT struct {
	secret []byte
}

// NewJWT builds a JWT signer/verifier from the raw JWT_SECRET.
func NewJWT(secret string) *JWT {
	return &JWT{secret: []byte(secret)}
}

// Sign issues a signed HS256 access token for the given user. The claims are
// exactly {sub: user_id, role, iat, exp} with a ≈15-minute expiry.
func (j *JWT) Sign(user User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  strconv.FormatInt(user.ID, 10),
		"role": user.Role,
		"iat":  now.Unix(),
		"exp":  now.Add(AccessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.secret)
	if err != nil {
		return "", err
	}
	return signed, nil
}

// Parse validates an HS256 token (signature, signing method, expiry) and
// returns its claims. It returns an error for any invalid token.
func (j *JWT) Parse(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}

// UserFromClaims extracts the User described by token claims. It reads the sub
// (user id) and role claims. The user's github_login / admin_email are not part
// of the token and are left empty; authorisation only needs the id + role.
func UserFromClaims(claims jwt.MapClaims) (*User, error) {
	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("missing sub claim")
	}
	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil || id <= 0 {
		return nil, fmt.Errorf("invalid sub claim")
	}
	role, _ := claims["role"].(string)
	if role == "" {
		role = "user"
	}
	return &User{ID: id, Role: role}, nil
}
