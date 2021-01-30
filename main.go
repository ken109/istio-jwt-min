package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", Hello)
	mux.HandleFunc("/token", GetToken)
	mux.HandleFunc("/jwks.json", GetJWK)
	srv := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: time.Second * 5}

	log.Printf("Listening on %v", ":8080")
	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func Hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello World"))
}

func GetToken(w http.ResponseWriter, r *http.Request) {
	tokenString, err := createSignedTokenString()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Signed token string:\n%v\n", tokenString)

	w.WriteHeader(http.StatusOK)
	w.Write(tokenString)
}

func createSignedTokenString() ([]byte, error) {
	privateKey, err := readRsaPrivateKey("private.key")
	if err != nil {
		return nil, fmt.Errorf("error reading private key file: %v\n", err)
	}

	t := jwt.New()

	t.Set(jwt.IssuerKey, "test@example.com")
	t.Set(jwt.SubjectKey, "test@example.com")
	t.Set(jwt.ExpirationKey, time.Now().Add(time.Second*20).Unix())
	t.Set(jwt.IssuedAtKey, time.Now().Unix())

	signed, err := jwt.Sign(t, jwa.RS256, privateKey)

	return signed, nil
}

func GetJWK(w http.ResponseWriter, r *http.Request) {
	privateKey, err := readRsaPrivateKey("private.key")
	if err != nil {
		return
	}

	key, err := jwk.New(privateKey.Public())
	if err != nil {
		fmt.Printf("failed to create symmetric key: %s\n", err)
		return
	}

	if _, ok := key.(jwk.RSAPublicKey); !ok {
		fmt.Printf("expected jwk.SymmetricKey, got %T\n", key)
		return
	}

	buf, err := json.MarshalIndent(map[string]interface{}{"keys": []interface{}{key}}, "", "  ")
	if err != nil {
		fmt.Printf("failed to marshal key into JSON: %s\n", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}

func readRsaPrivateKey(pemFile string) (*rsa.PrivateKey, error) {
	bytes, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("invalid private key data")
	}

	var key *rsa.PrivateKey
	if block.Type == "RSA PRIVATE KEY" {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	} else if block.Type == "PRIVATE KEY" {
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		key, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not RSA private key")
		}
	} else {
		return nil, fmt.Errorf("invalid private key type : %s", block.Type)
	}

	key.Precompute()

	if err := key.Validate(); err != nil {
		return nil, err
	}

	return key, nil
}
