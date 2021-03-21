package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPassword(t *testing.T) {
	hash, err := HashPassword("secret")
	require.NoError(t, err)
	require.True(t, ComparePassword(hash, "secret"))
}

func TestToken(t *testing.T) {
	const accountID = 1000
	secret := []byte("secret")

	token, err := EncodeToken(secret, accountID)
	require.NoError(t, err)

	expectedAccountID, err := DecodeToken(secret, token)
	require.NoError(t, err)
	require.EqualValues(t, expectedAccountID, accountID)
}
