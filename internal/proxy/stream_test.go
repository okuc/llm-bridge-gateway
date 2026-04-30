package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/okuc/llm-bridge-gateway/internal/converter"
	"github.com/okuc/llm-bridge-gateway/pkg/logger"
	opentrans "github.com/xy200303/OpenTrans"
)

func TestHandleStreamPassesThroughUpstreamErrorStatus(t *testing.T) {
	logger.Init("error", "text", "stdout", "")
	gin.SetMode(gin.TestMode)

	streamHandler := NewStreamHandler(converter.New())
	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header: http.Header{
			"Content-Type": []string{"text/event-stream"},
		},
		Body: io.NopCloser(strings.NewReader(`{"error":{"message":"rate limited"}}`)),
	}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	streamHandler.HandleStream(ctx, resp, opentrans.ProtocolClaude, opentrans.ProtocolOpenAI)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
	if got := w.Body.String(); !strings.Contains(got, "rate limited") {
		t.Fatalf("body = %q, want rate limited error", got)
	}
}
