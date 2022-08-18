package memo

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/kofalt/go-memoize"
	"net/http"
	"strings"
	"time"
)

const (
	NoExpiration time.Duration = 0
	NoCleanup    time.Duration = 0
)

// Options contains optional parameters for NewMemo.
type Options struct {
	// placeholder for future optional parameters
}

// NewMemo creates a new instance of Azure SDK pipeline [policy.Policy] which
// caches all GET requests. Value of zero or less for expiration and cleanup interval
// mean no expiration and no cleanup.
//
// This type is useful to plug in caching for Azure SDK without touching much code.
//
// Note that even though HTTP requests are cached, the unmarshalling of the response body is not cached,
// and it still happens every time.
func NewMemo(expiration time.Duration, cleanupInterval time.Duration, options *Options) *Memo {
	return &Memo{
		cache: memoize.NewMemoizer(expiration, cleanupInterval),
	}
}

// Memo implements azure-sdk-for-go/sdk/azcore/policy pipeline [policy.Policy] interface
// which caches all GET responses.
// Don't use this type directly, use NewMemo() instead.
type Memo struct {
	cache *memoize.Memoizer
}

// Do is part of [policy.Policy] interface.
func (m *Memo) Do(req *policy.Request) (*http.Response, error) {
	raw := req.Raw()
	method := strings.ToUpper(raw.Method)
	key := fmt.Sprintf("%s %s", method, raw.URL.String())

	if method != "GET" {
		return req.Next()
	}

	var result interface{}
	var err error
	result, err, _ = m.cache.Memoize(key, func() (interface{}, error) {
		return req.Next()
	})

	if result == nil {
		return nil, err
	}

	return result.(*http.Response), err
}
