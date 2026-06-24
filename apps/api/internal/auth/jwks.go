package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type jwksKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type jwksResponse struct {
	Keys []jwksKey `json:"keys"`
}

type keySet struct {
	url         string
	httpClient  *http.Client
	minInterval time.Duration

	mu        sync.RWMutex
	keys      map[string]*ecdsa.PublicKey
	lastFetch time.Time
}

func newKeySet(url string) *keySet {
	return &keySet{
		url:         url,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		minInterval: time.Minute,
		keys:        map[string]*ecdsa.PublicKey{},
	}
}

func (k *keySet) keyByID(kid string) (*ecdsa.PublicKey, error) {
	if key := k.cached(kid); key != nil {
		return key, nil
	}
	if err := k.refresh(); err != nil {
		return nil, err
	}
	if key := k.cached(kid); key != nil {
		return key, nil
	}
	return nil, errors.New("signing key not found")
}

func (k *keySet) cached(kid string) *ecdsa.PublicKey {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.keys[kid]
}

func (k *keySet) refresh() error {
	k.mu.Lock()
	if !k.lastFetch.IsZero() && time.Since(k.lastFetch) < k.minInterval {
		k.mu.Unlock()
		return nil
	}
	k.mu.Unlock()

	resp, err := k.httpClient.Get(k.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("jwks endpoint returned non-200 status")
	}

	var parsed jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return err
	}

	keys := make(map[string]*ecdsa.PublicKey, len(parsed.Keys))
	for _, key := range parsed.Keys {
		if key.Kty != "EC" || key.Crv != "P-256" {
			continue
		}
		pub, err := parseECPublicKey(key.X, key.Y)
		if err != nil {
			continue
		}
		keys[key.Kid] = pub
	}

	k.mu.Lock()
	k.keys = keys
	k.lastFetch = time.Now()
	k.mu.Unlock()
	return nil
}

func parseECPublicKey(xB64, yB64 string) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(xB64)
	if err != nil {
		return nil, err
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(yB64)
	if err != nil {
		return nil, err
	}
	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}
