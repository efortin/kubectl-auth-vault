package cache_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/efortin/kubectl-auth-vault/internal/cache"
)

var _ = Describe("Cache", func() {
	var (
		tmpDir    string
		cacheFile string
		c         *cache.Cache
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "cache-test")
		Expect(err).NotTo(HaveOccurred())
		cacheFile = filepath.Join(tmpDir, "test-cache.json")
		c = cache.New(cacheFile)
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("New", func() {
		It("should create a cache with the given file path", func() {
			Expect(c.FilePath()).To(Equal(cacheFile))
		})
	})

	Describe("DefaultCacheFile", func() {
		It("should generate correct filename for simple path", func() {
			path := cache.DefaultCacheFile("identity/oidc/token/my_role")
			Expect(filepath.Base(path)).To(Equal("vault_identity_oidc_token_my_role_token.json"))
		})

		It("should generate correct filename for path with special chars", func() {
			path := cache.DefaultCacheFile("auth/oidc/role")
			Expect(filepath.Base(path)).To(Equal("vault_auth_oidc_role_token.json"))
		})
	})

	Describe("Save and Load", func() {
		Context("with a valid token", func() {
			It("should save and load the token", func() {
				token := "test-token-12345"
				exp := time.Now().Add(time.Hour).Unix()

				err := c.Save(token, exp)
				Expect(err).NotTo(HaveOccurred())

				loadedToken, ok := c.Load()
				Expect(ok).To(BeTrue())
				Expect(loadedToken).To(Equal(token))
			})
		})

		Context("with an expired token", func() {
			It("should not load the token", func() {
				token := "expired-token"
				exp := time.Now().Add(-time.Hour).Unix()

				err := c.Save(token, exp)
				Expect(err).NotTo(HaveOccurred())

				_, ok := c.Load()
				Expect(ok).To(BeFalse())
			})
		})

		Context("when creating nested directories", func() {
			It("should create the directory structure", func() {
				nestedFile := filepath.Join(tmpDir, "subdir", "nested", "cache.json")
				nestedCache := cache.New(nestedFile)

				err := nestedCache.Save("token", time.Now().Add(time.Hour).Unix())
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(nestedFile)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Load", func() {
		Context("with a non-existent file", func() {
			It("should return false", func() {
				nonExistent := cache.New("/nonexistent/path/cache.json")
				_, ok := nonExistent.Load()
				Expect(ok).To(BeFalse())
			})
		})

		Context("with invalid JSON", func() {
			It("should return false", func() {
				err := os.WriteFile(cacheFile, []byte("not valid json"), 0600)
				Expect(err).NotTo(HaveOccurred())

				_, ok := c.Load()
				Expect(ok).To(BeFalse())
			})
		})
	})

	Describe("Clear", func() {
		It("should remove the cache file", func() {
			err := c.Save("token", time.Now().Add(time.Hour).Unix())
			Expect(err).NotTo(HaveOccurred())

			err = c.Clear()
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(cacheFile)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		Context("with a non-existent file", func() {
			It("should return an error", func() {
				nonExistent := cache.New("/nonexistent/path/cache.json")
				err := nonExistent.Clear()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
