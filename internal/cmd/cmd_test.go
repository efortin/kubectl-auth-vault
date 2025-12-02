package cmd_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/efortin/kubectl-auth-vault/internal/cmd"
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

func executeCommand(args ...string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf, err
}

var _ = Describe("Commands", func() {
	Describe("Root Command", func() {
		It("should display help", func() {
			buf, err := executeCommand("--help")
			Expect(err).NotTo(HaveOccurred())
			Expect(buf.Len()).To(BeNumerically(">", 0))
		})
	})

	Describe("Version Command", func() {
		It("should display version information", func() {
			buf, err := executeCommand("version")
			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(ContainSubstring("kubectl-auth_vault"))
		})
	})

	Describe("Config Show Command", func() {
		It("should display current configuration", func() {
			buf, err := executeCommand("config", "show")
			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(ContainSubstring("Current Configuration"))
		})

		Context("with vault address flag", func() {
			It("should display the provided vault address", func() {
				buf, err := executeCommand("config", "show", "--vault-addr", "https://vault.test.com")
				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(ContainSubstring("vault.test.com"))
			})
		})
	})

	Describe("Get Command", func() {
		Context("without VAULT_ADDR", func() {
			BeforeEach(func() {
				_ = os.Unsetenv("VAULT_ADDR")
			})

			It("should return an error", func() {
				_, err := executeCommand("get")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with mock vault server", func() {
			var (
				server    *httptest.Server
				testToken string
				tmpDir    string
			)

			BeforeEach(func() {
				exp := time.Now().Add(time.Hour).Unix()
				testToken = createTestJWT(exp)

				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					writeVaultResponse(w, vault.OIDCTokenResponse{
						Data: vault.OIDCTokenData{Token: testToken},
					})
				}))

				var err error
				tmpDir, err = os.MkdirTemp("", "cmd-test")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				server.Close()
				_ = os.RemoveAll(tmpDir)
			})

			It("should return a valid ExecCredential", func() {
				cacheFile := filepath.Join(tmpDir, "cache.json")
				buf, err := executeCommand(
					"get",
					"--vault-addr", server.URL,
					"--token-path", "identity/oidc/token/test",
					"--cache-file", cacheFile,
				)
				Expect(err).NotTo(HaveOccurred())

				var cred struct {
					Status struct {
						Token string `json:"token"`
					} `json:"status"`
				}
				err = json.Unmarshal(buf.Bytes(), &cred)
				Expect(err).NotTo(HaveOccurred())
				Expect(cred.Status.Token).To(Equal(testToken))
			})

			It("should use cache on second call", func() {
				callCount := 0
				server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					callCount++
					writeVaultResponse(w, vault.OIDCTokenResponse{
						Data: vault.OIDCTokenData{Token: testToken},
					})
				})

				cacheFile := filepath.Join(tmpDir, "cache.json")
				for i := 0; i < 2; i++ {
					_, err := executeCommand(
						"get",
						"--vault-addr", server.URL,
						"--token-path", "identity/oidc/token/test",
						"--cache-file", cacheFile,
					)
					Expect(err).NotTo(HaveOccurred())
				}

				Expect(callCount).To(Equal(1))
			})

			It("should not use cache with --no-cache", func() {
				callCount := 0
				server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					callCount++
					writeVaultResponse(w, vault.OIDCTokenResponse{
						Data: vault.OIDCTokenData{Token: testToken},
					})
				})

				cacheFile := filepath.Join(tmpDir, "cache.json")
				for i := 0; i < 2; i++ {
					_, err := executeCommand(
						"get",
						"--vault-addr", server.URL,
						"--token-path", "identity/oidc/token/test",
						"--cache-file", cacheFile,
						"--no-cache",
					)
					Expect(err).NotTo(HaveOccurred())
				}

				Expect(callCount).To(Equal(2))
			})

			It("should use VAULT_ADDR from environment", func() {
				GinkgoT().Setenv("VAULT_ADDR", server.URL)

				cacheFile := filepath.Join(tmpDir, "cache.json")
				_, err := executeCommand(
					"get",
					"--token-path", "identity/oidc/token/test",
					"--cache-file", cacheFile,
					"--no-cache",
				)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Config Test Command", func() {
		Context("without VAULT_ADDR", func() {
			BeforeEach(func() {
				_ = os.Unsetenv("VAULT_ADDR")
			})

			It("should return an error", func() {
				_, err := executeCommand("config", "test")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with mock vault server", func() {
			var server *httptest.Server

			BeforeEach(func() {
				exp := time.Now().Add(time.Hour).Unix()
				testToken := createTestJWT(exp)

				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					writeVaultResponse(w, vault.OIDCTokenResponse{
						Data: vault.OIDCTokenData{Token: testToken},
					})
				}))
			})

			AfterEach(func() {
				server.Close()
			})

			It("should show success message", func() {
				buf, err := executeCommand(
					"config", "test",
					"--vault-addr", server.URL,
					"--token-path", "identity/oidc/token/test",
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(ContainSubstring("Successfully"))
			})
		})

		Context("with server error", func() {
			var server *httptest.Server

			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			})

			AfterEach(func() {
				server.Close()
			})

			It("should return an error", func() {
				_, err := executeCommand(
					"config", "test",
					"--vault-addr", server.URL,
					"--token-path", "identity/oidc/token/test",
				)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
