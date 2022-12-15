package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/cli/internal/client"
)

func TestGetAppURL(t *testing.T) {
	s, err := client.GetAppURL("axiom.co")
	require.NoError(t, err)
	assert.Equal(t, "https://app.axiom.co", s)

	s, err = client.GetAppURL("app.axiom.co")
	require.NoError(t, err)
	assert.Equal(t, "https://app.axiom.co", s)

	s, err = client.GetAppURL("dev.axiom.co")
	require.NoError(t, err)
	assert.Equal(t, "https://app.dev.axiom.co", s)
}

func TestGetAPIURL(t *testing.T) {
	s, err := client.GetAPIURL("axiom.co")
	require.NoError(t, err)
	assert.Equal(t, "https://api.axiom.co", s)

	s, err = client.GetAPIURL("api.axiom.co")
	require.NoError(t, err)
	assert.Equal(t, "https://api.axiom.co", s)

	s, err = client.GetAPIURL("dev.axiom.co")
	require.NoError(t, err)
	assert.Equal(t, "https://api.dev.axiom.co", s)
}

func TestGetLoginURL(t *testing.T) {
	s, err := client.GetLoginURL("axiom.co")
	require.NoError(t, err)
	assert.Equal(t, "https://login.axiom.co", s)

	s, err = client.GetLoginURL("login.axiom.co")
	require.NoError(t, err)
	assert.Equal(t, "https://login.axiom.co", s)

	s, err = client.GetLoginURL("dev.axiom.co")
	require.NoError(t, err)
	assert.Equal(t, "https://login.dev.axiom.co", s)
}
