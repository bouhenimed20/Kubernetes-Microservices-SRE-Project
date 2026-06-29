package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResp struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
}

var (
	jwtSecret = getEnv("JWT_SECRET", "dev-secret-change-me")
	port      = getEnv("PORT", "8080")

	reqTotal     uint64
	reqErrors    uint64
	loginSuccess uint64
	loginFail    uint64
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", withMetrics(healthHandler))
	mux.HandleFunc("/health/live", withMetrics(healthLiveHandler))
	mux.HandleFunc("/health/ready", withMetrics(healthReadyHandler))
	mux.HandleFunc("/login", withMetrics(loginHandler))
	mux.HandleFunc("/verify", withMetrics(verifyHandler))
	mux.HandleFunc("/metrics", metricsHandler)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("auth-service listening on :%s", port)
	log.Fatal(srv.ListenAndServe())
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "auth-service"})
}

func healthLiveHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "live"})
}

func healthReadyHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		atomic.AddUint64(&reqErrors, 1)
		return
	}

	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		atomic.AddUint64(&reqErrors, 1)
		return
	}

	if req.Username != "admin" || req.Password != "admin" {
		atomic.AddUint64(&loginFail, 1)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	atomic.AddUint64(&loginSuccess, 1)
	exp := time.Now().Add(30 * time.Minute).Unix()

	token, err := signJWT(req.Username, exp, jwtSecret)
	if err != nil {
		atomic.AddUint64(&reqErrors, 1)
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, loginResp{Token: token, ExpiresAt: exp})
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		atomic.AddUint64(&reqErrors, 1)
		http.Error(w, "missing bearer token", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(auth, "Bearer ")

	claims, err := verifyJWT(token, jwtSecret)
	if err != nil {
		atomic.AddUint64(&reqErrors, 1)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"valid": true,
		"sub":   claims.Sub,
		"exp":   claims.Exp,
	})
}

type jwtClaims struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
}

func signJWT(sub string, exp int64, secret string) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims := jwtClaims{Sub: sub, Exp: exp, Iat: time.Now().Unix()}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	unsigned := header + "." + payload
	sig := hmacSHA256(unsigned, secret)
	return unsigned + "." + sig, nil
}

func verifyJWT(token string, secret string) (jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return jwtClaims{}, fmt.Errorf("bad token format")
	}

	unsigned := parts[0] + "." + parts[1]
	expected := hmacSHA256(unsigned, secret)
	if !hmac.Equal([]byte(parts[2]), []byte(expected)) {
		return jwtClaims{}, fmt.Errorf("bad signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return jwtClaims{}, fmt.Errorf("bad payload")
	}

	var c jwtClaims
	if err := json.Unmarshal(payloadBytes, &c); err != nil {
		return jwtClaims{}, fmt.Errorf("bad claims")
	}
	if time.Now().Unix() > c.Exp {
		return jwtClaims{}, fmt.Errorf("expired")
	}
	return c, nil
}

func hmacSHA256(data, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func withMetrics(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&reqTotal, 1)
		next(w, r)
	}
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	fmt.Fprintf(w, "# HELP http_requests_total Total HTTP requests.\n")
	fmt.Fprintf(w, "# TYPE http_requests_total counter\n")
	fmt.Fprintf(w, "http_requests_total %d\n", atomic.LoadUint64(&reqTotal))

	fmt.Fprintf(w, "# HELP http_errors_total Total HTTP errors.\n")
	fmt.Fprintf(w, "# TYPE http_errors_total counter\n")
	fmt.Fprintf(w, "http_errors_total %d\n", atomic.LoadUint64(&reqErrors))

	fmt.Fprintf(w, "# HELP auth_login_success_total Successful logins.\n")
	fmt.Fprintf(w, "# TYPE auth_login_success_total counter\n")
	fmt.Fprintf(w, "auth_login_success_total %d\n", atomic.LoadUint64(&loginSuccess))

	fmt.Fprintf(w, "# HELP auth_login_fail_total Failed logins.\n")
	fmt.Fprintf(w, "# TYPE auth_login_fail_total counter\n")
	fmt.Fprintf(w, "auth_login_fail_total %d\n", atomic.LoadUint64(&loginFail))
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
