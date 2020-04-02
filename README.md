# google-jwt
Simple package for authorizing google JWT token

## Installation

```golang
go get github.com/stefanosx/google-jwt
```

## Usage

You need to intialize the package with an `http.HandlerFunc()` and a domain(in case you want to whitelist specific domains only)

```golang
import {
	jwt "github.com/stefanosx/google-jwt"
}

// Pass empty string if you don't want to whitelist a domain
authorizationMiddleware := jwt.Init(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authorized Hello World"))
}), "")

// whitelisting a domain
authorizationMiddleware := jwt.Init(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authorized Hello World"))
}), "stefanosx.com")

mux.Handle("/test", authorizationMiddleware)
```
