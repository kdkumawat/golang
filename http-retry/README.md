# GoLang - HTTP Retries

When making HTTP requests, it is common to encounter errors due to network issues or server unavailability. To handle these errors, we can implement a retry mechanism.

This article details how to extend the Golang HTTP client to include retry functionality. We will cover how to implement retry count, backoff strategy, retry on network errors, and response status codes (502, 503, 504). Additionally, we will explore how to prevent the request body from being closed and how to drain the body to use the same connection.

## Retry Count

For example, let's say we want to retry a request three times before giving up. We can create a `RetryCount` variable and set it to 3.

```go
const RetryCount = 3
```

## Backoff Strategy

A backoff strategy is a method for delaying retries after a failed request. The idea is to increase the delay between retries to give the server time to recover.

For example, we can start with a delay of one second and double the delay after each retry. We can implement this using an exponential backoff strategy.

```go
func backoff(retries int) time.Duration {
    return time.Duration(math.Pow(2, float64(retries))) * time.Second
}
```

## Retry on Network Errors and Response Status Codes

We can also implement retry logic for specific network errors and response status codes. For example, if we encounter a network error, we can retry the request. Similarly, if we receive a 502, 503, or 504 status code, we can retry the request.

```go
func shouldRetry(err error, resp *http.Response) bool {
    if err != nil {
        return true
    }

    if resp.StatusCode == http.StatusBadGateway ||
        resp.StatusCode == http.StatusServiceUnavailable ||
        resp.StatusCode == http.StatusGatewayTimeout {
        return true
    }

    return false
}
```

## Drain Body to Use Same Connection

To reuse the same connection when retrying requests. To do this, we need to drain the response body before closing the connection.

```go
func drainBody(resp *http.Response) {
    if resp.Body != nil {
        io.Copy(ioutil.Discard, resp.Body)
        resp.Body.Close()
    }
}
```

## Prevent Request Body from Being Closed

By default, the Golang HTTP client will close the request body after a request is sent. This can cause issues when retrying requests since the body may have already been closed. To prevent this from happening, we can create a custom `RoundTripper` that wraps the default `Transport` and prevents the request body from being closed.

```go
type retryableTransport struct {
    transport http.RoundTripper
}

func (t *retryableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    // Clone the request body
    var bodyBytes []byte
    if req.Body != nil {
        bodyBytes, _ = ioutil.ReadAll(req.Body)
        req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
    }

    // Send the request
    resp, err := t.transport.RoundTrip(req)

    // Retry logic
    retries := 0
    for shouldRetry(err, resp) && retries < RetryCount {
        // Wait for the specified backoff period
        time.Sleep(backoff(retries))

				// We're going to retry, consume any response to reuse the connection.
				drainBody(resp)

        // Clone the request body again
        if req.Body != nil {
            req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
        }

        // Retry the request
        resp, err = t.transport.RoundTrip(req)
        retries++
    }

    // Return the response
    return resp, err
}
```

With these methods in place, we can now create our custom `http.Client` that includes retry functionality.

```go
func NewRetryableClient() *http.Client {
    transport := &retryableTransport{
        transport: &http.Transport{},
    }

    return &http.Client{
        Transport: transport,
    }
}
```

We can now use our new `http.Client` to make requests that automatically retry on failure.

```go
client := NewRetryableClient()
resp, err := client.Get("https://reqres.in/api/users/2")
if err != nil {
	fmt.Println("err", err)
}

defer resp.Body.Close()

body, err := ioutil.ReadAll(resp.Body)
if err != nil {
	fmt.Println("err reading body : %w", err)
}

fmt.Println("resp", string(body))

```

Implementing these features can be extremely useful in production environments where network instability and server unavailability can be common. By having a retry mechanism in place, we can greatly improve the reliability and resilience of our applications.
