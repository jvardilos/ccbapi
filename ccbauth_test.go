package ccbapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsAuthorizedValid(t *testing.T) {

	testToken := &Token{
		ExpiresIn: time.Now().Add(1 * time.Hour).Unix(),
	}

	assert.True(t, isAuthorized(testToken), "Token should be valid")
}

func TestIsAuthorizedExpired(t *testing.T) {

	testToken := &Token{
		ExpiresIn: time.Now().Add(10).Unix(),
	}

	assert.False(t, isAuthorized(testToken), "Token should be expired")

}
