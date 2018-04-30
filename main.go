package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

// EnvPort, EnvRemoteHost, EnvUserID, EnvPrivateKey environment variable constants.
var (
	EnvPort       = "PORT"
	EnvRemoteHost = "REMOTE_HOST"

	EnvUserID     = "K_USER_ID"
	EnvPrivateKey = "K_PRIVATE_KEY"
)

// HeaderTimestamp, HeaderUserID, HeaderSignature header constants.
var (
	HeaderTimestamp = "Third-Party-Timestamp"
	HeaderUserID    = "Third-Party-User-Id"
	HeaderSignature = "Third-Party-Signature"
)

func main() {
	if err := checkEnv(); err != nil {
		log.Fatalf("Missing parameters: %s", err.Error())
	}
	port := os.Getenv(EnvPort)
	addr := fmt.Sprintf(":%s", port)
	http.Handle("/", proxyHander{})
	log.Printf("Listening on `%s`", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error listening/serving: %s\n", err.Error())
	}
}

func checkEnv() error {
	if os.Getenv(EnvRemoteHost) == "" {
		return errors.New(fmt.Sprintf("Host parameter %s not present", EnvRemoteHost))
	}
	if os.Getenv(EnvPort) == "" {
		return errors.New(fmt.Sprintf("Port parameter %s not present", EnvPort))
	}
	if os.Getenv(EnvUserID) == "" {
		return errors.New(fmt.Sprintf("User ID parameter %s not present", EnvUserID))
	}
	if os.Getenv(EnvPrivateKey) == "" {
		return errors.New(fmt.Sprintf("Private Key parameter %s not present", EnvPrivateKey))
	}

	return nil
}

type proxyHander struct{}

func (ph proxyHander) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Println(req.Method, req.URL.Path, req.URL.RawQuery)
	// Create remote request.
	remoteReq, reqErr := http.NewRequest(
		req.Method,
		fmt.Sprintf("%s%s?%s", os.Getenv(EnvRemoteHost), req.URL.Path, req.URL.RawQuery),
		req.Body,
	)
	if reqErr != nil {
		log.Printf("Error creating remote request: %s\n", reqErr.Error())
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Copy headers from incoming request.
	remoteReq.Header = req.Header
	// Sign request.
	signRequest(remoteReq)

	// Send request off.
	resp, doErr := http.DefaultClient.Do(remoteReq)
	if doErr != nil {
		log.Printf("Error talking to remote host: %s", doErr.Error())
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Copy data back to pending local request.
	rw.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(rw, resp.Body); err != nil {
		log.Printf("Error copying data from remote: %s", err.Error())
	}

	return
}

func signRequest(req *http.Request) {
	if req.Header.Get(HeaderTimestamp) == "" {
		ts := time.Now().UnixNano()
		req.Header.Add(HeaderTimestamp, fmt.Sprint(ts))
		log.Printf("Signing %s: %s", HeaderTimestamp, req.Header.Get(HeaderTimestamp))
	}

	if req.Header.Get(HeaderUserID) == "" {
		req.Header.Add(HeaderUserID, os.Getenv(EnvUserID))
		log.Printf("Signing %s: %s", HeaderUserID, req.Header.Get(HeaderUserID))
	}

	if req.Header.Get(HeaderSignature) == "" {
		envelope := fmt.Sprintf(
			"%s-%s",
			req.Header.Get(HeaderUserID),
			req.Header.Get(HeaderTimestamp),
		)

		hash := hmac.New(sha512.New, []byte(os.Getenv(EnvPrivateKey)))
		_, _ = hash.Write([]byte(envelope))
		req.Header.Add(HeaderSignature, base64.URLEncoding.EncodeToString(hash.Sum(nil)))
		log.Printf("Signing %s: %s", HeaderSignature, req.Header.Get(HeaderSignature))
	}
}
