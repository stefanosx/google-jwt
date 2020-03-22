package googlejwt

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	cache "github.com/patrickmn/go-cache"
)

type AuthorizationHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type AuthorizationMiddleware struct {
	WrappedHandler AuthorizationHandler
	Cache          *cache.Cache
	Domain         string
}

func (h AuthorizationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := authorized(r.Header.Get("Authorization"), h.Cache, h.Domain)
	if !t {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	h.WrappedHandler.ServeHTTP(w, r)
}

func Init(handler AuthorizationHandler, domain string) AuthorizationMiddleware {
	ch := cache.New(60*time.Minute, 120*time.Minute)
	return AuthorizationMiddleware{
		WrappedHandler: handler,
		Cache:          ch,
		Domain:         domain,
	}
}

func authorized(authHeader string, c *cache.Cache, domain string) bool {
	tokenString := strings.Replace(authHeader, "Bearer ", "", -1)

	token, err := jws.ParseString(tokenString)
	if err != nil {
		log.Printf("Failed to parse with error: %s", err)
		return false
	}
	key, _ := token.Signatures()[0].ProtectedHeaders().Get("kid")
	keyID := fmt.Sprintf("%v", key)

	var dat map[string]interface{}
	if err := json.Unmarshal(token.Payload(), &dat); err != nil {
		log.Printf("Failed to parse Json, with error: %s", err)
		return false
	}
	if domain != "" && dat["hd"] != domain {
		return false
	}

	keys, err := findKeys(c, keyID)
	if err != nil {
		log.Printf("Failed to lookup key: %s", err)
		return false
	}

	payload, err := jws.VerifyWithJWK([]byte(tokenString), keys[0])
	if err != nil {
		log.Printf("Failed to verify message: %s", err)
		return false
	}
	return payload != nil
}

func findKeys(c *cache.Cache, keyID string) ([]jwk.Key, error) {
	url := "https://www.googleapis.com/oauth2/v3/certs"
	var keys []jwk.Key
	loop := true
	count := 0
	for loop {
		fetchedKeys, found := c.Get("fetchedKeys")
		var set *jwk.Set
		var err error
		if found {
			set = fetchedKeys.(*jwk.Set)
		} else {
			set, err = jwk.Fetch(url)
			if err != nil {
				log.Printf("Failed to parse JWK: %s", err)
				return keys, err
			}
			c.Set("fetchedKeys", set, cache.NoExpiration)
		}

		keys = set.LookupKeyID(keyID)
		if len(keys) > 0 {
			return keys, nil
		} else if count > 1 {
			return nil, errors.New("Can not find the key")
		}
		c.Delete("fetchedKeys")
		count++
	}
	return nil, errors.New("Can not find the key")
}
