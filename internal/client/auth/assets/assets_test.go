package assets_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/cli/internal/client/auth/assets"
)

func TestWriteStatusPage(t *testing.T) {
	rec := httptest.NewRecorder()
	err := assets.WriteStatusPage(rec)
	require.NoError(t, err)

	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	if assert.NotEmpty(t, rec.Body.String()) {
		assert.NotContains(t, rec.Body.String(), "<body class=\"error\">")
	}
}

func TestWriteStatusPageWithError(t *testing.T) {
	rec := httptest.NewRecorder()
	err := assets.WriteStatusPageWithError(rec, errors.New("some error"))
	require.NoError(t, err)

	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	if assert.NotEmpty(t, rec.Body.String()) {
		assert.Contains(t, rec.Body.String(), "<body class=\"error\">")
		assert.Contains(t, rec.Body.String(), "some error")
	}
}
