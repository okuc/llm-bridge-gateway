package proxy

import (
	"context"
	"net/http"
	"time"

	"github.com/okuc/llm-bridge-gateway/internal/config"
	"github.com/okuc/llm-bridge-gateway/pkg/logger"
)

// RetryableFunc 可重试的函数类型
type RetryableFunc func() (*http.Response, error)

// WithRetry 带重试的执行
func WithRetry(ctx context.Context, route *config.RouteConfig, fn RetryableFunc) (*http.Response, error) {
	maxAttempts := route.Retry.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	backoff := route.Retry.Backoff
	if backoff <= 0 {
		backoff = 1 * time.Second
	}

	var lastErr error
	var lastResp *http.Response

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := fn()
		if err == nil {
			// 检查是否需要重试（5xx错误）
			if resp.StatusCode < 500 {
				return resp, nil
			}

			// 5xx错误，如果不是最后一次尝试，关闭响应体准备重试
			if attempt < maxAttempts {
				resp.Body.Close()
			}
			lastResp = resp
			lastErr = nil
		} else {
			lastErr = err
		}

		// 如果不是最后一次尝试，等待后重试
		if attempt < maxAttempts {
			waitTime := backoff * time.Duration(attempt)
			logger.Warn("Retrying request",
				"attempt", attempt,
				"max_attempts", maxAttempts,
				"wait_time", waitTime.String(),
				"error", lastErr,
			)

			// 使用 context 实现可取消的等待
			select {
			case <-time.After(waitTime):
			case <-ctx.Done():
				if lastResp != nil {
					lastResp.Body.Close()
				}
				return nil, ctx.Err()
			}
		}
	}

	// 所有重试都失败
	if lastErr != nil {
		return nil, lastErr
	}

	// 返回最后一次的响应（Body未关闭）
	return lastResp, nil
}
