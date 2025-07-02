package x

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

// HttpMethod represents HTTP methods as a string type.
type HttpMethod string

// HTTP method constants.
const (
	HttpMethodGET     HttpMethod = "GET"
	HttpMethodPOST    HttpMethod = "POST"
	HttpMethodPUT     HttpMethod = "PUT"
	HttpMethodDELETE  HttpMethod = "DELETE"
	HttpMethodPATCH   HttpMethod = "PATCH"
	HttpMethodOPTIONS HttpMethod = "OPTIONS"
	HttpMethodHEAD    HttpMethod = "HEAD"
	HttpMethodCONNECT HttpMethod = "CONNECT"
	HttpMethodTRACE   HttpMethod = "TRACE"
)

func AddCookies(client *http.Client, targetURL string, cookies []*http.Cookie) error {
	// Validate inputs
	if client == nil {
		return fmt.Errorf("http client cannot be nil")
	}

	if len(cookies) == 0 {
		return nil // Nothing to do
	}

	// Parse URL string
	cookieURL, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	// If the client doesn't have a jar, create one
	if client.Jar == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return fmt.Errorf("failed to create cookie jar: %w", err)
		}
		client.Jar = jar
	}

	// Set the cookies in the client.jar
	client.Jar.SetCookies(cookieURL, cookies)

	return nil
}

func FetchJson(
	method HttpMethod,
	url string,
	headers map[string]string,
	cookies map[string]string,
	body []byte,
	timeout time.Duration,
	val any,
) error {
	// create a new http client
	client := &http.Client{}
	defer client.CloseIdleConnections()

	if bodyBytes, _, err := DoRequest(
		context.Background(), client, method, url,
		headers, cookies, body, timeout,
	); err != nil {
		return err
	} else {
		return json.Unmarshal(bodyBytes, val)
	}
}

// DoRequest sends an HTTP request with given parameters, supports common compression
// encodings (gzip, deflate, br, zstd), and returns the decompressed response body,
// status code, and error.
// It requires external dependencies:
// go get github.com/andybalholm/brotli
// go get github.com/klauspost/compress/zstd
func DoRequest(
	ctx context.Context,
	client *http.Client,
	method HttpMethod,
	url string,
	headers map[string]string,
	cookies map[string]string,
	body []byte,
	timeout time.Duration,
) ([]byte, int, error) {

	var data io.Reader = nil
	if body != nil {
		data = bytes.NewReader(body)
	}

	// Add timeout control (if timeout > 0)
	reqCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel() // Ensure resources are released after timeout or completion
	}

	// Validate HTTP method
	switch method {
	case HttpMethodGET, HttpMethodPOST, HttpMethodPUT, HttpMethodDELETE,
		HttpMethodPATCH, HttpMethodOPTIONS, HttpMethodHEAD, HttpMethodCONNECT, HttpMethodTRACE:
	// Valid method
	default:
		// Use http.StatusBadRequest for invalid input like method
		return nil, http.StatusBadRequest, fmt.Errorf("invalid HTTP method: %s", method)
	}

	// Create request with context
	req, err := http.NewRequestWithContext(reqCtx, string(method), url, data)
	if err != nil {
		// Use http.StatusInternalServerError for internal errors
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Accept-Encoding header to indicate support for compression
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add cookies
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	// Set default Content-Type if not provided and body exists
	if data != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// --- Send Request ---
	resp, err := client.Do(req)
	if err != nil {
		// Check for context deadline exceeded (timeout)
		if ctxErr := reqCtx.Err(); ctxErr == context.DeadlineExceeded {
			return nil, http.StatusGatewayTimeout, fmt.Errorf("request timed out after %v: %w", timeout, err)
		}
		// Use http.StatusServiceUnavailable or map specific network errors if possible
		return nil, http.StatusServiceUnavailable, fmt.Errorf("failed to send request: %w", err)
	}
	// Ensure the original response body is always closed eventually.
	// Decompressor wrappers might have their own Close methods, handled below.
	defer resp.Body.Close()

	// --- Handle Response Body Decompression ---
	var reader io.ReadCloser = resp.Body // Start with the original body by default
	var decompressErr error

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, decompressErr = gzip.NewReader(resp.Body)
		if decompressErr != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to create gzip reader: %w", decompressErr)
		}
		// gzip.Reader implements Close, which will close the underlying resp.Body
		defer reader.Close()
	case "deflate":
		reader, decompressErr = zlib.NewReader(resp.Body)
		if decompressErr != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to create deflate reader: %w", decompressErr)
		}
		// zlib.Reader implements Close, closing the underlying resp.Body
		defer reader.Close()
	case "br":
		// brotli.NewReader returns an io.Reader, not io.ReadCloser.
		// We still need to ensure the original resp.Body is closed via the outer defer.
		reader = io.NopCloser(brotli.NewReader(resp.Body))
		// No defer reader.Close() needed here as NopCloser's Close does nothing,
		// and brotli.Reader doesn't have Close(). The original resp.Body.Close() handles it.
	case "zstd":
		var zstdDecoder *zstd.Decoder
		zstdDecoder, decompressErr = zstd.NewReader(resp.Body)
		if decompressErr != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to create zstd reader: %w", decompressErr)
		}
		// zstd.Decoder has a Close method to release resources.
		// .IOReadCloser() gives us an io.ReadCloser wrapper.
		reader = zstdDecoder.IOReadCloser()
		defer reader.Close() // Close the zstd reader wrapper
	default:
		// No recognized Content-Encoding or no encoding; use resp.Body directly (already assigned to reader)
	}

	// --- Process Response ---

	// It's often better to read the body even for non-200 responses for debugging,
	// but we'll stick to the original logic of failing fast for non-OK status.
	// The previous "Poe homepage" error message was too specific.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { // Check for non-successful status codes
		// Attempt to read a small part of the body for error context? Optional.
		// bodyPreview, _ := io.ReadAll(io.LimitReader(reader, 1024)) // Example
		// Discard rest of body to allow connection reuse
		_, _ = io.Copy(io.Discard, reader)
		return nil, resp.StatusCode, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	// Read the potentially decompressed body using the final 'reader'
	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		// Check if the error occurred during decompression (e.g., corrupted stream)
		if decompressErr != nil && err == io.ErrUnexpectedEOF { // Or other decompression specific errors
			return nil, http.StatusInternalServerError, fmt.Errorf("failed reading decompressed body (possible corrupted stream): %w", err)
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to read response body: %w", err)
	}

	// Success
	return bodyBytes, resp.StatusCode, nil
}

// DownloadFile sends a GET request to the specified URL and returns the response body as a byte slice.
func DownloadFile(ctx context.Context, urlStr string, timeout time.Duration) ([]byte, error) {
	// Create HTTP client
	client := &http.Client{}
	defer client.CloseIdleConnections()

	// Send GET request
	respBytes, _, err := DoRequest(context.Background(), client, HttpMethodGET, urlStr, nil, nil, nil, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", err)
	}

	return respBytes, nil
}

func GetURLBody(urlStr string) string {
	ret, _ := DownloadFile(context.Background(), urlStr, time.Minute)
	return string(ret)
}
