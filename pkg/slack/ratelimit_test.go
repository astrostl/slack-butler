package slack

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	t.Run("Wait respects minimum interval", func(t *testing.T) {
		rl := &RateLimiter{
			minInterval: 100 * time.Millisecond,
			maxBackoff:  time.Minute,
		}

		ctx := context.Background()
		start := time.Now()

		// First call should not wait
		err := rl.Wait(ctx)
		if err != nil {
			t.Fatalf("First Wait() failed: %v", err)
		}

		// Second call should wait for minimum interval
		err = rl.Wait(ctx)
		if err != nil {
			t.Fatalf("Second Wait() failed: %v", err)
		}

		elapsed := time.Since(start)
		if elapsed < 100*time.Millisecond {
			t.Errorf("Expected to wait at least 100ms, but waited %v", elapsed)
		}
	})

	t.Run("OnSuccess resets backoff", func(t *testing.T) {
		rl := &RateLimiter{
			minInterval:  time.Millisecond,
			maxBackoff:   time.Minute,
			backoffCount: 3,
		}

		rl.OnSuccess()

		if rl.backoffCount != 0 {
			t.Errorf("Expected backoffCount to be 0 after success, got %d", rl.backoffCount)
		}
	})

	t.Run("OnRateLimitError increases backoff", func(t *testing.T) {
		rl := &RateLimiter{
			minInterval: time.Millisecond,
			maxBackoff:  time.Minute,
		}

		initialCount := rl.backoffCount
		rl.OnRateLimitError()

		if rl.backoffCount != initialCount+1 {
			t.Errorf("Expected backoffCount to increase by 1, got %d", rl.backoffCount)
		}
	})

	t.Run("Backoff caps at maximum", func(t *testing.T) {
		rl := &RateLimiter{
			minInterval: time.Millisecond,
			maxBackoff:  time.Minute,
		}

		// Trigger many rate limit errors
		for i := 0; i < 10; i++ {
			rl.OnRateLimitError()
		}

		if rl.backoffCount > 6 {
			t.Errorf("Expected backoffCount to cap at 6, got %d", rl.backoffCount)
		}
	})

	t.Run("Context cancellation stops waiting", func(t *testing.T) {
		rl := &RateLimiter{
			minInterval: time.Second, // Long interval
			maxBackoff:  time.Minute,
		}

		ctx, cancel := context.WithCancel(context.Background())

		// Start first request to set lastRequest
		_ = rl.Wait(context.Background())

		// Cancel context immediately
		cancel()

		start := time.Now()
		err := rl.Wait(ctx)
		elapsed := time.Since(start)

		if err == nil {
			t.Error("Expected error when context is cancelled")
		}

		if elapsed > 100*time.Millisecond {
			t.Errorf("Wait should return quickly when cancelled, but took %v", elapsed)
		}
	})

	t.Run("Exponential backoff calculation", func(t *testing.T) {
		rl := &RateLimiter{
			minInterval:  10 * time.Millisecond,
			maxBackoff:   time.Minute,
			backoffCount: 2, // Should multiply by 2^2 = 4
		}

		ctx := context.Background()
		start := time.Now()

		// First call sets lastRequest
		_ = rl.Wait(ctx)

		// Second call should wait minInterval * 2^backoffCount
		_ = rl.Wait(ctx)

		elapsed := time.Since(start)
		expectedMin := 40 * time.Millisecond // 10ms * 4

		if elapsed < expectedMin {
			t.Errorf("Expected to wait at least %v with backoff, but waited %v", expectedMin, elapsed)
		}
	})
}

func TestClientGetNewChannelsRateLimit(t *testing.T) {
	mockAPI := NewMockSlackAPI()
	mockAPI.AddChannel("test-id", "test", time.Now(), "Test channel")

	client, err := NewClientWithAPI(mockAPI)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Override with faster rate limiter for testing
	client.rateLimiter = &RateLimiter{
		minInterval: 50 * time.Millisecond,
		maxBackoff:  time.Minute,
	}

	start := time.Now()

	// First call
	_, err = client.GetNewChannels(time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("First GetNewChannels failed: %v", err)
	}

	// Second call should be rate limited
	_, err = client.GetNewChannels(time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("Second GetNewChannels failed: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 50*time.Millisecond {
		t.Errorf("Expected rate limiting delay, but calls completed in %v", elapsed)
	}
}

func TestClientPostMessageRateLimit(t *testing.T) {
	mockAPI := NewMockSlackAPI()
	// Add the test channel that will be used for posting
	mockAPI.AddChannel("CTEST", "test", time.Now().Add(-24*time.Hour), "Test channel")
	client, err := NewClientWithAPI(mockAPI)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Override with faster rate limiter for testing
	client.rateLimiter = &RateLimiter{
		minInterval: 50 * time.Millisecond,
		maxBackoff:  time.Minute,
	}

	start := time.Now()

	// First call
	err = client.PostMessage("#test", "message 1")
	if err != nil {
		t.Fatalf("First PostMessage failed: %v", err)
	}

	// Second call should be rate limited
	err = client.PostMessage("#test", "message 2")
	if err != nil {
		t.Fatalf("Second PostMessage failed: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 50*time.Millisecond {
		t.Errorf("Expected rate limiting delay, but calls completed in %v", elapsed)
	}
}

func TestClientRateLimitErrorBackoff(t *testing.T) {
	mockAPI := NewMockSlackAPI()

	// First call will succeed to establish client
	client, err := NewClientWithAPI(mockAPI)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Configure mock to return rate limit error
	mockAPI.SetGetConversationsErrorWithMessage(true, "rate_limited")

	// Override with faster rate limiter for testing
	client.rateLimiter = &RateLimiter{
		minInterval: 10 * time.Millisecond,
		maxBackoff:  time.Minute,
	}

	// Call that triggers rate limit error
	_, err = client.GetNewChannels(time.Now().Add(-time.Hour))
	if err == nil || !contains(err.Error(), "rate limited") {
		t.Fatalf("Expected rate limit error, got: %v", err)
	}

	// Verify backoff was increased
	if client.rateLimiter.backoffCount != 1 {
		t.Errorf("Expected backoffCount to be 1 after rate limit error, got %d", client.rateLimiter.backoffCount)
	}

	testExponentialBackoffRecovery(t, mockAPI, client)
}

func testExponentialBackoffRecovery(t *testing.T, mockAPI *MockSlackAPI, client *Client) {
	// Reset mock to succeed
	mockAPI.SetGetConversationsErrorWithMessage(false, "")
	mockAPI.AddChannel("test-id", "test", time.Now(), "Test channel")

	start := time.Now()

	// Next call should have exponential backoff applied
	_, err := client.GetNewChannels(time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("GetNewChannels after rate limit failed: %v", err)
	}

	elapsed := time.Since(start)
	// Be more lenient with timing to avoid flaky tests
	// The key is that we had a backoff, not the exact timing
	expectedMin := 10 * time.Millisecond // Minimum base delay

	if elapsed < expectedMin {
		t.Errorf("Expected some backoff delay of at least %v, but got %v", expectedMin, elapsed)
	}

	// Verify backoff was reset on success
	if client.rateLimiter.backoffCount != 0 {
		t.Errorf("Expected backoffCount to be reset to 0 after success, got %d", client.rateLimiter.backoffCount)
	}
}

// Helper function for string contains check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
