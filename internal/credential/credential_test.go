package credential_test

import (
	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/efortin/kubectl-auth-vault/internal/credential"
)

var _ = Describe("Credential", func() {
	Describe("New", func() {
		It("should create an ExecCredential with correct values", func() {
			token := "test-token-12345"
			cred := credential.New(token)

			Expect(cred.APIVersion).To(Equal("client.authentication.k8s.io/v1"))
			Expect(cred.Kind).To(Equal("ExecCredential"))
			Expect(cred.Status.Token).To(Equal(token))
		})
	})

	Describe("Output", func() {
		It("should write valid JSON to the writer", func() {
			token := "test-token-12345"
			var buf bytes.Buffer

			err := credential.Output(&buf, token)
			Expect(err).NotTo(HaveOccurred())

			var cred credential.ExecCredential
			err = json.Unmarshal(buf.Bytes(), &cred)
			Expect(err).NotTo(HaveOccurred())
			Expect(cred.APIVersion).To(Equal("client.authentication.k8s.io/v1"))
			Expect(cred.Kind).To(Equal("ExecCredential"))
			Expect(cred.Status.Token).To(Equal(token))
		})

		It("should handle empty token", func() {
			var buf bytes.Buffer
			err := credential.Output(&buf, "")
			Expect(err).NotTo(HaveOccurred())

			var cred credential.ExecCredential
			err = json.Unmarshal(buf.Bytes(), &cred)
			Expect(err).NotTo(HaveOccurred())
			Expect(cred.Status.Token).To(BeEmpty())
		})
	})

	Describe("JSON", func() {
		It("should return valid JSON bytes", func() {
			cred := credential.New("test-token")
			data, err := cred.JSON()
			Expect(err).NotTo(HaveOccurred())

			var parsed credential.ExecCredential
			err = json.Unmarshal(data, &parsed)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.Status.Token).To(Equal("test-token"))
		})
	})
})
