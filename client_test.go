package httpclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestHTTPClient main test ...
func TestHTTPClient(t *testing.T) {

	status := NewStatus()
	mockClient := NewGenericClient("fakeService", status)

	// No-Retry Policy
	policy := NewNoRetryPolicy()
	mockClient.SetRetryPolicy(policy)
	breaker := NewBreaker(BreakerConfig{
		ErrorThreshold:   len(policy.Backoffs()),
		SuccessThreshold: 1,
		Timeout:          10 * time.Second,
	})
	mockClient.SetCircuitBreaker(breaker)

	req, err := http.NewRequest("GET", "localhost:8080", nil)
	if err != nil {
		t.Fatalf("Services Package Test - NoRetryPolicy - Something went wrong with creating a fake request: %+v", err)
	}

	_, err = mockClient.Do(req)
	if err.Error() != "circuit breaker is open" {
		t.Fatalf("Services Package Test - NoRetryPolicy - Something went wrong sending the fake request to the non-existent server: %+v", err)
	}

	if mockClient.GetStatus() {
		t.Fatal("Services Package Test - NoRetryPolicy - Mock Service should have been marked as down.")
	}

	// Single-Retry Policy
	policy2 := NewSingleRetryPolicy(100 * time.Millisecond)
	mockClient.SetRetryPolicy(policy2)
	breaker2 := NewBreaker(BreakerConfig{
		ErrorThreshold:   len(policy2.Backoffs()),
		SuccessThreshold: 1,
		Timeout:          10 * time.Second,
	})
	mockClient.SetCircuitBreaker(breaker2)

	req, err = http.NewRequest("GET", "localhost:8080", nil)
	if err != nil {
		t.Fatalf("Services Package Test - SingleRetryPolicy - Something went wrong with creating a fake request: %+v", err)
	}

	_, err = mockClient.Do(req)
	if err.Error() != "circuit breaker is open" {
		t.Fatalf("Services Package Test - SingleRetryPolicy - Something went wrong sending the fake request to the non-existent server: %+v", err)
	}

	if mockClient.GetStatus() {
		t.Fatal("Services Package Test - SingleRetryPolicy - Mock Service should have been marked as down.")
	}

	// Constant-Retry Policy
	policy3 := NewConstantRetryPolicy(100*time.Millisecond, 5)
	mockClient.SetRetryPolicy(policy3)
	breaker3 := NewBreaker(BreakerConfig{
		ErrorThreshold:   len(policy3.Backoffs()),
		SuccessThreshold: 1,
		Timeout:          10 * time.Second,
	})
	mockClient.SetCircuitBreaker(breaker3)

	req, err = http.NewRequest("GET", "localhost:8080", nil)
	if err != nil {
		t.Fatalf("Services Package Test - ConstantRetryPolicy - Something went wrong with creating a fake request: %+v", err)
	}

	_, err = mockClient.Do(req)
	if err.Error() != "circuit breaker is open" {
		t.Fatalf("Services Package Test - ConstantRetryPolicy - Something went wrong sending the fake request to the non-existent server: %+v", err)
	}

	if mockClient.GetStatus() {
		t.Fatal("Services Package Test - ConstantRetryPolicy - Mock Service should have been marked as down.")
	}

	// Exponential-Retry Policy
	policy4 := NewExponentialRetryPolicy(DefaultBackoffConfig)
	mockClient.SetRetryPolicy(policy4)
	breaker4 := NewBreaker(BreakerConfig{
		ErrorThreshold:   len(policy4.Backoffs()),
		SuccessThreshold: 1,
		Timeout:          10 * time.Second,
	})
	mockClient.SetCircuitBreaker(breaker4)

	req, err = http.NewRequest("GET", "localhost:8080", nil)
	if err != nil {
		t.Fatalf("Services Package Test - ExponentialRetryPolicy - Something went wrong with creating a fake request: %+v", err)
	}

	_, err = mockClient.Do(req)
	if err.Error() != "circuit breaker is open" {
		t.Fatalf("Services Package Test - ExponentialRetryPolicy - Something went wrong sending the fake request to the non-existent server: %+v", err)
	}

	if mockClient.GetStatus() {
		t.Fatal("Services Package Test - ExponentialRetryPolicy - Mock Service should have been marked as down.")
	}

	// Fake Working Service ...
	policy5 := NewExponentialRetryPolicy(DefaultBackoffConfig)
	mockClient.SetRetryPolicy(policy5)
	breaker5 := NewBreaker(BreakerConfig{
		ErrorThreshold:   len(policy5.Backoffs()),
		SuccessThreshold: 1,
		Timeout:          10 * time.Second,
	})
	mockClient.SetCircuitBreaker(breaker5)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	req, err = http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatalf("Services Package Test - FakeService - Something went wrong with creating a fake request: %+v", err)
	}

	resp, err := mockClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("Services Package Test - FakeService - Something went wrong sending the fake request to the fake server: %+v", err)
	}

	if !mockClient.GetStatus() {
		t.Fatal("Services Package Test - FakeService - Mock Service should have been marked as up.")
	}
}
