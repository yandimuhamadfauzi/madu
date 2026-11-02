package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

const formTokenCookie = "ft"

// newFormToken membuat token acak untuk form sebelum login (double-submit cookie).
// Penyerang di situs lain tidak bisa membaca atau menyetel cookie ini, jadi tidak bisa memalsukan form.
func newFormToken(c *gin.Context, secure bool) string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	t := base64.RawURLEncoding.EncodeToString(b)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(formTokenCookie, t, 1800, "/", "", secure, true)
	return t
}

func checkFormToken(c *gin.Context) bool {
	ck, err := c.Cookie(formTokenCookie)
	if err != nil || ck == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(ck), []byte(c.PostForm("ft"))) == 1
}
