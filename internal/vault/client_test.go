package vault_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/efortin/kubectl-auth-vault/internal/jwt"
	"github.com/efortin/kubectl-auth-vault/internal/vault"
)

func writeVaultResponse(w http.ResponseWriter, resp vault.OIDCTokenResponse) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createTestJWT(exp int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := jwt.Payload{Exp: exp, Iat: time.Now().Unix(), Sub: "test"}
	payloadBytes, _ := json.Marshal(payload)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))
	return strings.Join([]string{header, payloadB64, signature}, ".")
}

var _ = Describe("Vault Client", func() {
	Describe("NewClient", func() {
		It("should create a client with valid address", func() {
			client, err := vault.NewClient("http://localhost:8200")
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
		})

		It("should return error for invalid address", func() {
			_, err := vault.NewClient("://invalid-url")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetOIDCToken", func() {
		var server *httptest.Server

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("with a valid response", func() {
			It("should return the token and expiration", func() {
				exp := time.Now().Add(time.Hour).Unix()
				testToken := createTestJWT(exp)

				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/v1/identity/oidc/token/test_role"))
					writeVaultResponse(w, vault.OIDCTokenResponse{
						Data: vault.OIDCTokenData{Token: testToken},
					})
				}))

				client, err := vault.NewClient(server.URL)
				Expect(err).NotTo(HaveOccurred())

				token, gotExp, err := client.GetOIDCToken(context.Background(), "identity/oidc/token/test_role")
				Expect(err).NotTo(HaveOccurred())
				Expect(token).To(Equal(testToken))
				Expect(gotExp).To(Equal(exp))
			})
		})

		Context("with no data in response", func() {
			// Using map to simulate empty response without data field
			It("should return an error", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]interface{}{})
				}))

				client, err := vault.NewClient(server.URL)
				Expect(err).NotTo(HaveOccurred())

				_, _, err = client.GetOIDCToken(context.Background(), "identity/oidc/token/test_role")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with no token field", func() {
			// Using map to simulate response without token field in data
			It("should return an error", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					resp := map[string]interface{}{
						"data": map[string]interface{}{
							"other_field": "value",
						},
					}
					_ = json.NewEncoder(w).Encode(resp)
				}))

				client, err := vault.NewClient(server.URL)
				Expect(err).NotTo(HaveOccurred())

				_, _, err = client.GetOIDCToken(context.Background(), "identity/oidc/token/test_role")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with token not being a string", func() {
			// Using map[string]interface{} to simulate malformed Vault response
			It("should return an error", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					resp := map[string]interface{}{
						"data": map[string]interface{}{
							"token": 12345,
						},
					}
					_ = json.NewEncoder(w).Encode(resp)
				}))

				client, err := vault.NewClient(server.URL)
				Expect(err).NotTo(HaveOccurred())

				_, _, err = client.GetOIDCToken(context.Background(), "identity/oidc/token/test_role")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with invalid JWT (no exp)", func() {
			It("should return token with default expiration", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					writeVaultResponse(w, vault.OIDCTokenResponse{
						Data: vault.OIDCTokenData{Token: "invalid-jwt-without-exp"},
					})
				}))

				client, err := vault.NewClient(server.URL)
				Expect(err).NotTo(HaveOccurred())

				token, exp, err := client.GetOIDCToken(context.Background(), "identity/oidc/token/test_role")
				Expect(err).NotTo(HaveOccurred())
				Expect(token).To(Equal("invalid-jwt-without-exp"))

				now := time.Now().Unix()
				Expect(exp).To(BeNumerically(">=", now))
				Expect(exp).To(BeNumerically("<=", now+3700))
			})
		})

		Context("with server error", func() {
			It("should return an error", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))

				client, err := vault.NewClient(server.URL)
				Expect(err).NotTo(HaveOccurred())

				_, _, err = client.GetOIDCToken(context.Background(), "identity/oidc/token/test_role")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
