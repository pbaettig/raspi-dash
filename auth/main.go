package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type BasicAuthUsers map[string]string

func (b *BasicAuthUsers) PopulateFromEnv() {
	prefix := "RASPI_DASH_USER_"
	envUsers := make(map[string]string)
	for _, e := range os.Environ() {

		if strings.HasPrefix(e, prefix) {
			kv := strings.SplitN(e, "=", 2)
			user := strings.ToLower(strings.TrimPrefix(kv[0], prefix))
			envUsers[user] = kv[1]
		}
	}
	*b = envUsers
}

func (b BasicAuthUsers) IsAuthorized(authorization string) bool {
	authValue := strings.TrimRight(strings.TrimPrefix(authorization, "Basic "), "=")

	for k, v := range b {
		valid := base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", k, v)))

		if authValue == valid {
			return true
		}
	}
	return false
}

func (b BasicAuthUsers) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.Header().Add("WWW-Authenticate", "Basic realm=\"Documents access\", charset=\"UTF-8\"")
			http.Error(w, "need authorization", http.StatusUnauthorized)
			return
		}

		if !b.IsAuthorized(auth) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
