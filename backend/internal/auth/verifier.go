package auth

import (
	"context"
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

type Verifier struct {
	jwksURL    string
	skipInDev  bool
	httpClient *http.Client

	mu        sync.RWMutex
	keys      map[string]any
	fetchedAt time.Time
}

func NewVerifier(jwksURL string, skipInDev bool) *Verifier {
	return &Verifier{
		jwksURL:    jwksURL,
		skipInDev:  skipInDev,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		keys:       map[string]any{},
	}
}

func (v *Verifier) Verify(ctx context.Context, tokenString string) (User, error) {
	if v.skipInDev && strings.TrimSpace(tokenString) == "" {
		return User{ID: "dev-user", Email: "dev@example.com", Name: "Dev User"}, nil
	}

	if strings.TrimSpace(tokenString) == "" {
		return User{}, errors.New("missing bearer token")
	}

	keys, err := v.getKeys(ctx)
	if err != nil {
		return User{}, err
	}

	claims := jwt.MapClaims{}
	parsed, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		key, ok := keys[kid]
		if !ok {
			return nil, fmt.Errorf("unknown key id %q", kid)
		}
		return key, nil
	}, jwt.WithValidMethods([]string{"EdDSA", "RS256"}))
	if err != nil || !parsed.Valid {
		return User{}, fmt.Errorf("invalid token: %w", err)
	}

	userID, _ := claims["sub"].(string)
	if userID == "" {
		userID, _ = claims["user_id"].(string)
	}
	if userID == "" {
		return User{}, errors.New("token missing subject")
	}

	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)

	return User{ID: userID, Email: email, Name: name}, nil
}

type jwksDocument struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (v *Verifier) getKeys(ctx context.Context) (map[string]any, error) {
	v.mu.RLock()
	if time.Since(v.fetchedAt) < 10*time.Minute && len(v.keys) > 0 {
		keys := v.keys
		v.mu.RUnlock()
		return keys, nil
	}
	v.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create jwks request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	var doc jwksDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode jwks: %w", err)
	}

	keys := make(map[string]any, len(doc.Keys))
	for _, key := range doc.Keys {
		switch {
		case key.Kty == "OKP" && key.Crv == "Ed25519":
			pub, err := parseEd25519(key.X)
			if err != nil {
				return nil, err
			}
			keys[key.Kid] = pub
		case key.Kty == "RSA":
			pub, err := parseRSA(key.N, key.E)
			if err != nil {
				return nil, err
			}
			keys[key.Kid] = pub
		}
	}

	v.mu.Lock()
	v.keys = keys
	v.fetchedAt = time.Now()
	v.mu.Unlock()

	return keys, nil
}

func parseEd25519(x string) (any, error) {
	raw, err := base64.RawURLEncoding.DecodeString(x)
	if err != nil {
		return nil, fmt.Errorf("decode ed25519 key: %w", err)
	}
	if len(raw) != 32 {
		return nil, fmt.Errorf("invalid ed25519 key length")
	}
	return ed25519.PublicKey(raw), nil
}

func parseRSA(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("decode rsa n: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("decode rsa e: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}
