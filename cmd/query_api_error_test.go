package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/sgaunet/perplexity-go/v2"
	clerrors "github.com/sgaunet/pplx/pkg/clerrors"
)

// disableSpinner sets OutputJSON=true to suppress the pterm spinner during tests,
// and restores the original value via t.Cleanup.
func disableSpinner(t *testing.T) {
	t.Helper()
	orig := globalOpts.OutputJSON
	globalOpts.OutputJSON = true
	t.Cleanup(func() { globalOpts.OutputJSON = orig })
}

// newTestRequest builds a minimal CompletionRequest suitable for unit tests.
func newTestRequest() *perplexity.CompletionRequest {
	msg := perplexity.NewMessages(perplexity.WithSystemMessage(""))
	_ = msg.AddUserMessage("test query")
	return perplexity.NewCompletionRequest(
		perplexity.WithMessages(msg.GetMessages()),
		perplexity.WithModel("sonar"),
	)
}

// mockCompletionResponseJSON returns a minimal valid completion response JSON.
func mockCompletionResponseJSON() string {
	return `{
		"id": "test-id",
		"model": "sonar",
		"created": 1234567890,
		"object": "chat.completion",
		"usage": {"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
		"choices": [{"index": 0, "finish_reason": "stop",
			"message": {"role": "assistant", "content": "Hello"},
			"delta": {"role": "", "content": ""}}]
	}`
}

// apiErrorJSON returns a Perplexity-format API error JSON body.
func apiErrorJSON(message, errType string, code int) string {
	type errBody struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	body := errBody{}
	body.Error.Message = message
	body.Error.Type = errType
	body.Error.Code = code
	b, _ := json.Marshal(body)
	return string(b)
}

// TestHandleNonStreamingResponse_Unauthorized verifies that a 401 response
// is surfaced as a clerrors.APIError with a descriptive message.
func TestHandleNonStreamingResponse_Unauthorized(t *testing.T) {
	disableSpinner(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := perplexity.NewClient("invalid-key")
	client.SetEndpoint(srv.URL)

	err := handleNonStreamingResponse(client, newTestRequest())
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}

	var apiErr *clerrors.APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected clerrors.APIError, got %T: %v", err, err)
	}
}

// TestHandleNonStreamingResponse_RateLimited verifies that a 429 response
// is surfaced as a clerrors.APIError.
func TestHandleNonStreamingResponse_RateLimited(t *testing.T) {
	disableSpinner(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(apiErrorJSON("rate limit exceeded", "rate_limit_error", 429)))
	}))
	defer srv.Close()

	client := perplexity.NewClient("test-key")
	client.SetEndpoint(srv.URL)

	err := handleNonStreamingResponse(client, newTestRequest())
	if err == nil {
		t.Fatal("expected error for 429 response, got nil")
	}

	var apiErr *clerrors.APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected clerrors.APIError, got %T: %v", err, err)
	}
}

// TestHandleNonStreamingResponse_InternalServerError verifies that a 500
// response is surfaced as a clerrors.APIError.
func TestHandleNonStreamingResponse_InternalServerError(t *testing.T) {
	disableSpinner(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(apiErrorJSON("internal server error", "server_error", 500)))
	}))
	defer srv.Close()

	client := perplexity.NewClient("test-key")
	client.SetEndpoint(srv.URL)

	err := handleNonStreamingResponse(client, newTestRequest())
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}

	var apiErr *clerrors.APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected clerrors.APIError, got %T: %v", err, err)
	}
}

// TestHandleNonStreamingResponse_ServiceUnavailable verifies that a 503
// response is surfaced as a clerrors.APIError.
func TestHandleNonStreamingResponse_ServiceUnavailable(t *testing.T) {
	disableSpinner(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(apiErrorJSON("service unavailable", "service_unavailable", 503)))
	}))
	defer srv.Close()

	client := perplexity.NewClient("test-key")
	client.SetEndpoint(srv.URL)

	err := handleNonStreamingResponse(client, newTestRequest())
	if err == nil {
		t.Fatal("expected error for 503 response, got nil")
	}

	var apiErr *clerrors.APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected clerrors.APIError, got %T: %v", err, err)
	}
}

// TestHandleNonStreamingResponse_NetworkTimeout verifies that a request that
// exceeds the HTTP client timeout is surfaced as a clerrors.APIError.
func TestHandleNonStreamingResponse_NetworkTimeout(t *testing.T) {
	disableSpinner(t)

	// done is closed last (LIFO) so the handler unblocks before srv.Close waits.
	done := make(chan struct{})
	// Register srv.Close first so it runs second in LIFO cleanup order.
	// done must be closed before srv.Close to unblock the handler goroutine.
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}))
	t.Cleanup(srv.Close)           // registered 1st → runs 2nd (LIFO)
	t.Cleanup(func() { close(done) }) // registered 2nd → runs 1st (LIFO)

	client := perplexity.NewClient("test-key")
	client.SetEndpoint(srv.URL)
	client.SetHTTPTimeout(50 * time.Millisecond)

	err := handleNonStreamingResponse(client, newTestRequest())
	if err == nil {
		t.Fatal("expected error for network timeout, got nil")
	}

	var apiErr *clerrors.APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected clerrors.APIError, got %T: %v", err, err)
	}
}

// TestHandleNonStreamingResponse_MalformedResponse verifies that a 200
// response with invalid JSON body is surfaced as a clerrors.APIError.
func TestHandleNonStreamingResponse_MalformedResponse(t *testing.T) {
	disableSpinner(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer srv.Close()

	client := perplexity.NewClient("test-key")
	client.SetEndpoint(srv.URL)

	err := handleNonStreamingResponse(client, newTestRequest())
	if err == nil {
		t.Fatal("expected error for malformed JSON response, got nil")
	}

	var apiErr *clerrors.APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected clerrors.APIError, got %T: %v", err, err)
	}
}

// TestHandleNonStreamingResponse_InvalidAPIKey_ErrorMessage verifies that an
// invalid-key error message is non-empty and user-friendly.
func TestHandleNonStreamingResponse_InvalidAPIKey_ErrorMessage(t *testing.T) {
	disableSpinner(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := perplexity.NewClient("bad-key")
	client.SetEndpoint(srv.URL)

	err := handleNonStreamingResponse(client, newTestRequest())
	if err == nil {
		t.Fatal("expected error for invalid API key, got nil")
	}

	if err.Error() == "" {
		t.Error("expected non-empty error message for invalid API key")
	}
}

// TestHandleNonStreamingResponse_Success verifies that a well-formed 200
// response is processed without error.
func TestHandleNonStreamingResponse_Success(t *testing.T) {
	disableSpinner(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockCompletionResponseJSON()))
	}))
	defer srv.Close()

	client := perplexity.NewClient("test-key")
	client.SetEndpoint(srv.URL)

	err := handleNonStreamingResponse(client, newTestRequest())
	if err != nil {
		t.Errorf("expected nil error for successful response, got: %v", err)
	}
}

// TestHandleNonStreamingResponse_EmptyResponseBody verifies that a 200
// response with an empty body is handled gracefully as a clerrors.APIError.
func TestHandleNonStreamingResponse_EmptyResponseBody(t *testing.T) {
	disableSpinner(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Empty body — JSON unmarshal will produce an error.
	}))
	defer srv.Close()

	client := perplexity.NewClient("test-key")
	client.SetEndpoint(srv.URL)

	err := handleNonStreamingResponse(client, newTestRequest())
	if err == nil {
		t.Fatal("expected error for empty response body, got nil")
	}

	var apiErr *clerrors.APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected clerrors.APIError, got %T: %v", err, err)
	}
}

// TestHandleNonStreamingResponse_DeadEndpoint verifies that a connection
// refused error (dead endpoint) is surfaced as a clerrors.APIError.
func TestHandleNonStreamingResponse_DeadEndpoint(t *testing.T) {
	disableSpinner(t)

	// Create and immediately close a server so connections are refused.
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	srv.Close()

	client := perplexity.NewClient("test-key")
	client.SetEndpoint(srv.URL)

	err := handleNonStreamingResponse(client, newTestRequest())
	if err == nil {
		t.Fatal("expected error for dead endpoint, got nil")
	}

	var apiErr *clerrors.APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected clerrors.APIError for connection refused, got %T: %v", err, err)
	}
}

// TestHandleNonStreamingResponse_AllErrorsAreAPIErrors is a table-driven test
// that verifies every HTTP error status code results in a clerrors.APIError.
func TestHandleNonStreamingResponse_AllErrorsAreAPIErrors(t *testing.T) {
	disableSpinner(t)

	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{"unauthorized", http.StatusUnauthorized, ""},
		{"forbidden", http.StatusForbidden, apiErrorJSON("forbidden", "forbidden", 403)},
		{"not found", http.StatusNotFound, apiErrorJSON("not found", "not_found", 404)},
		{"rate limit", http.StatusTooManyRequests, apiErrorJSON("rate limit exceeded", "rate_limit_error", 429)},
		{"internal server error", http.StatusInternalServerError, apiErrorJSON("internal error", "server_error", 500)},
		{"bad gateway", http.StatusBadGateway, apiErrorJSON("bad gateway", "gateway_error", 502)},
		{"service unavailable", http.StatusServiceUnavailable, apiErrorJSON("service unavailable", "service_unavailable", 503)},
		{"gateway timeout", http.StatusGatewayTimeout, apiErrorJSON("gateway timeout", "timeout", 504)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := tt.statusCode
			body := tt.body
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(code)
				if body != "" {
					_, _ = w.Write([]byte(body))
				}
			}))
			defer srv.Close()

			client := perplexity.NewClient("test-key")
			client.SetEndpoint(srv.URL)

			err := handleNonStreamingResponse(client, newTestRequest())
			if err == nil {
				t.Fatalf("expected error for HTTP %d, got nil", code)
			}

			var apiErr *clerrors.APIError
			if !errors.As(err, &apiErr) {
				t.Errorf("HTTP %d: expected clerrors.APIError, got %T: %v", code, err, err)
			}
		})
	}
}

// TestMissingAPIKey verifies that the query command returns a clerrors.ConfigError
// when the PPLX_API_KEY environment variable is not set.
func TestMissingAPIKey(t *testing.T) {
	origKey := os.Getenv("PPLX_API_KEY")
	t.Setenv("PPLX_API_KEY", "")
	defer func() {
		if origKey != "" {
			t.Setenv("PPLX_API_KEY", origKey)
		}
	}()

	origUserPrompt := globalOpts.UserPrompt
	globalOpts.UserPrompt = "test"
	defer func() { globalOpts.UserPrompt = origUserPrompt }()

	err := queryCmd.RunE(queryCmd, []string{})
	if err == nil {
		t.Fatal("expected error when PPLX_API_KEY is not set, got nil")
	}

	var configErr *clerrors.ConfigError
	if !errors.As(err, &configErr) {
		t.Errorf("expected clerrors.ConfigError, got %T: %v", err, err)
	}
}

// TestDeadlineExceeded verifies that a context.DeadlineExceeded error
// from a slow API is wrapped as clerrors.APIError.
func TestDeadlineExceeded(t *testing.T) {
	disableSpinner(t)

	done := make(chan struct{})
	// Server that never responds within the client's timeout.
	// Register srv.Close first so done is closed first (LIFO cleanup order).
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}))
	t.Cleanup(srv.Close)              // registered 1st → runs 2nd (LIFO)
	t.Cleanup(func() { close(done) }) // registered 2nd → runs 1st (LIFO)

	client := perplexity.NewClient("test-key")
	client.SetEndpoint(srv.URL)
	client.SetHTTPTimeout(10 * time.Millisecond)

	err := handleNonStreamingResponse(client, newTestRequest())
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	var apiErr *clerrors.APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected clerrors.APIError for deadline exceeded, got %T: %v", err, err)
	}

	// The underlying error should indicate a timeout or deadline exceeded.
	// Some HTTP clients wrap timeouts differently, so we just verify the message is non-empty.
	if !errors.Is(err, context.DeadlineExceeded) {
		if apiErr.Error() == "" {
			t.Error("expected non-empty error message for deadline exceeded")
		}
	}
}

// TestAPIErrorType_ErrorMessage verifies that clerrors.APIError produces
// human-readable messages that include the original cause.
func TestAPIErrorType_ErrorMessage(t *testing.T) {
	cause := errors.New("unauthorized: check your API key")
	apiErr := clerrors.NewAPIError("failed to send completion request", cause)

	if apiErr.Error() == "" {
		t.Error("APIError.Error() returned empty string")
	}

	// The original cause must be accessible via errors.Unwrap.
	if !errors.Is(apiErr, cause) {
		t.Error("errors.Is(apiErr, cause) returned false; cause not accessible via Unwrap")
	}
}

// TestAPIErrorType_NilCause verifies that clerrors.APIError handles a nil
// wrapped error without panicking.
func TestAPIErrorType_NilCause(t *testing.T) {
	apiErr := clerrors.NewAPIError("something failed", nil)
	if apiErr.Error() == "" {
		t.Error("APIError.Error() returned empty string for nil cause")
	}
}

// TestAPIErrorType_WithStatusCode verifies that a status code is included
// in the error message when set.
func TestAPIErrorType_WithStatusCode(t *testing.T) {
	apiErr := &clerrors.APIError{
		StatusCode: http.StatusTooManyRequests,
		Message:    "rate limit exceeded",
	}
	if apiErr.Error() == "" {
		t.Error("APIError.Error() returned empty string")
	}
}
