package httpclient

import (
"bytes"
"context"
"fmt"
"io"
"net/http"
"os"
"time"
)

const (
defaultLibraryServiceURL = "http://localhost:3000"
defaultInternalAPIKey    = "local_internal_key"
)

type LibraryClient struct {
baseURL    string
apiKey     string
httpClient *http.Client
}

func NewLibraryClient() *LibraryClient {
baseURL := os.Getenv("LIBRARY_SERVICE_URL")
if baseURL == "" {
baseURL = defaultLibraryServiceURL
}

apiKey := os.Getenv("INTERNAL_API_KEY")
if apiKey == "" {
apiKey = defaultInternalAPIKey
}

return &LibraryClient{
baseURL: baseURL,
apiKey:  apiKey,
httpClient: &http.Client{
Timeout: 5 * time.Second,
},
}
}

func (c *LibraryClient) ReserveBook(ctx context.Context, bookID string) error {
url := fmt.Sprintf("%s/internal/books/%s/reserve", c.baseURL, bookID)
return c.postInternal(ctx, url)
}

func (c *LibraryClient) ReleaseBook(ctx context.Context, bookID string) error {
url := fmt.Sprintf("%s/internal/books/%s/release", c.baseURL, bookID)
return c.postInternal(ctx, url)
}

func (c *LibraryClient) postInternal(ctx context.Context, url string) error {
req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(nil))
if err != nil {
return err
}

req.Header.Set("x-internal-api-key", c.apiKey)

res, err := c.httpClient.Do(req)
if err != nil {
return err
}
defer res.Body.Close()

if res.StatusCode >= 200 && res.StatusCode <= 299 {
return nil
}

body, _ := io.ReadAll(res.Body)

return fmt.Errorf("library service returned status %d: %s", res.StatusCode, string(body))
}
