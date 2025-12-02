package jwt_test

import (
	"encoding/base64"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/efortin/kubectl-auth-vault/internal/jwt"
)

func createTestJWT(payload jwt.Payload) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payloadBytes, _ := json.Marshal(payload)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))
	return header + "." + payloadB64 + "." + signature
}

var _ = Describe("JWT", func() {
	Describe("ExtractExp", func() {
		Context("with a valid JWT containing exp", func() {
			It("should return the exp value", func() {
				token := createTestJWT(jwt.Payload{Exp: 1234567890, Iat: 1234567800, Sub: "test"})
				exp, err := jwt.ExtractExp(token)
				Expect(err).NotTo(HaveOccurred())
				Expect(exp).To(Equal(int64(1234567890)))
			})
		})

		Context("with a JWT without exp", func() {
			It("should return an error", func() {
				token := createTestJWT(jwt.Payload{Iat: 1234567800, Sub: "test"})
				_, err := jwt.ExtractExp(token)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no exp claim"))
			})
		})

		Context("with an invalid JWT format", func() {
			It("should return an error for too few parts", func() {
				_, err := jwt.ExtractExp("invalid.token")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid JWT format"))
			})

			It("should return an error for no dots", func() {
				_, err := jwt.ExtractExp("invalidtoken")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with invalid base64 payload", func() {
			It("should return an error", func() {
				_, err := jwt.ExtractExp("header.!!!invalid!!!.signature")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with invalid JSON payload", func() {
			It("should return an error", func() {
				invalidPayload := base64.RawURLEncoding.EncodeToString([]byte("not json"))
				_, err := jwt.ExtractExp("header." + invalidPayload + ".signature")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("DecodePayload", func() {
		Context("with a valid JWT", func() {
			It("should decode all fields correctly", func() {
				token := createTestJWT(jwt.Payload{
					Exp: 1234567890,
					Iat: 1234567800,
					Iss: "test-issuer",
					Sub: "test-subject",
				})
				payload, err := jwt.DecodePayload(token)
				Expect(err).NotTo(HaveOccurred())
				Expect(payload.Exp).To(Equal(int64(1234567890)))
				Expect(payload.Iat).To(Equal(int64(1234567800)))
				Expect(payload.Iss).To(Equal("test-issuer"))
				Expect(payload.Sub).To(Equal("test-subject"))
			})
		})

		Context("with standard base64 encoding", func() {
			It("should handle standard base64 encoded payload", func() {
				payload := jwt.Payload{Exp: 1234567890}
				payloadBytes, _ := json.Marshal(payload)
				payloadB64 := base64.StdEncoding.EncodeToString(payloadBytes)
				header := base64.StdEncoding.EncodeToString([]byte(`{"alg":"RS256"}`))
				signature := base64.StdEncoding.EncodeToString([]byte("sig"))
				token := header + "." + payloadB64 + "." + signature

				got, err := jwt.DecodePayload(token)
				Expect(err).NotTo(HaveOccurred())
				Expect(got.Exp).To(Equal(int64(1234567890)))
			})
		})

		Context("with invalid format", func() {
			It("should return an error", func() {
				_, err := jwt.DecodePayload("not.a.valid.jwt.token")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
