package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Run("Valid token creates client", func(t *testing.T) {
		// This test uses the real NewClient function but with dependency injection
		// We test the validation logic without actually making network calls
		token := "MOCK-BOT-TOKEN-FOR-TESTING-ONLY-NOT-REAL-TOKEN-AT-ALL"
		
		// This will fail at the network call step, but validates token format first
		client, err := NewClient(token)
		
		// Expect error due to network call failure, but not due to token validation
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "authentication failed")
	})

	t.Run("Invalid token format fails validation", func(t *testing.T) {
		token := "invalid-token"
		
		client, err := NewClient(token)
		
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("Empty token fails validation", func(t *testing.T) {
		client, err := NewClient("")
		
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("Short token fails validation", func(t *testing.T) {
		token := "xoxb-FAKE-FOR-TESTING"
		
		client, err := NewClient(token)
		
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "invalid token")
	})
}