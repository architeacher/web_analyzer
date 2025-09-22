package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/architeacher/svc-web-analyzer/internal/domain"
	"github.com/architeacher/svc-web-analyzer/internal/handlers"
	"github.com/architeacher/svc-web-analyzer/internal/infrastructure"
	"github.com/architeacher/svc-web-analyzer/internal/service"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type SSEHandlers struct {
	analysisService service.ApplicationService
	logger          *infrastructure.Logger
}

func NewSSEHandlers(analysisService service.ApplicationService, logger *infrastructure.Logger) *SSEHandlers {
	return &SSEHandlers{
		analysisService: analysisService,
		logger:          logger,
	}
}

func (h *SSEHandlers) HandleGetAnalysisEvents(w http.ResponseWriter, r *http.Request, analysisId openapi_types.UUID, params handlers.GetAnalysisEventsParams) {
	h.logger.Debug().
		Str("method", "GetAnalysisEvents").
		Str("analysis_id", analysisId.String()).
		Msg("Processing SSE analysis events query")

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "CacheClient-Control")
	w.Header().Set("API-Version", "v1")

	// Convert UUID
	id, err := uuid.Parse(analysisId.String())
	if err != nil {
		h.writeSSEError(w, "INVALID_ANALYSIS_ID", "Invalid analysis ID format")
		return
	}

	// Get event channel from analysis app
	eventChan, err := h.analysisService.FetchAnalysisEvents(r.Context(), id.String())
	if err != nil {
		if err == domain.ErrAnalysisNotFound {
			h.writeSSEError(w, "ANALYSIS_NOT_FOUND", "HTMLParser not found")
			return
		}
		h.writeSSEError(w, "INTERNAL_SERVER_ERROR", "Failed to get analysis events")
		return
	}

	// Create context for handling client disconnection
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Set up flusher for real-time streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.writeSSEError(w, "STREAMING_NOT_SUPPORTED", "Streaming not supported")
		return
	}

	// Send initial connection event
	h.writeSSEEvent(w, "connected", map[string]interface{}{
		"analysis_id": analysisId.String(),
		"timestamp":   time.Now().Format(time.RFC3339),
	})
	flusher.Flush()

	// Keep-alive ticker
	keepAliveTicker := time.NewTicker(30 * time.Second)
	defer keepAliveTicker.Stop()

	h.logger.Info().Str("analysis_id", analysisId.String()).Msg("SSE connection established")

	// Event streaming loop
	for {
		select {
		case <-ctx.Done():
			h.logger.Debug().Str("analysis_id", analysisId.String()).Msg("SSE connection closed by client")
			return

		case <-keepAliveTicker.C:
			// Send keep-alive event
			h.writeSSEEvent(w, "keepalive", map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
			})
			flusher.Flush()

		case event, ok := <-eventChan:
			if !ok {
				// Channel closed, analysis is complete
				h.writeSSEEvent(w, "stream_end", map[string]interface{}{
					"message":   "HTMLParser stream ended",
					"timestamp": time.Now().Format(time.RFC3339),
				})
				flusher.Flush()
				h.logger.Debug().Str("analysis_id", analysisId.String()).Msg("SSE stream ended")
				return
			}

			// Convert domain event to SSE event
			h.writeAnalysisEvent(w, event)
			flusher.Flush()

			// If this is a final event (completed or failed), close the stream
			if event.Type == domain.EventTypeCompleted || event.Type == domain.EventTypeFailed {
				h.logger.Debug().
					Str("analysis_id", analysisId.String()).
					Str("event_type", event.Type).
					Msg("Received final event, will close SSE stream")

				// Send a small delay before closing to ensure client receives the event
				time.Sleep(100 * time.Millisecond)
				return
			}
		}
	}
}

func (h *SSEHandlers) writeSSEEvent(w http.ResponseWriter, eventType string, data interface{}) {
	// Generate event ID
	eventID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Convert data to JSON
	dataJSON, err := json.Marshal(data)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to marshal SSE event data")
		return
	}

	// Write SSE event format
	fmt.Fprintf(w, "id: %s\n", eventID)
	fmt.Fprintf(w, "event: %s\n", eventType)
	fmt.Fprintf(w, "data: %s\n\n", dataJSON)
}

func (h *SSEHandlers) writeAnalysisEvent(w http.ResponseWriter, event domain.AnalysisEvent) {
	eventData := map[string]interface{}{
		"event_id":  event.EventID,
		"type":      event.Type,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	switch event.Type {
	case domain.EventTypeStarted:
		eventData["message"] = "HTMLParser started"
		if analysis, ok := event.Data.(*domain.Analysis); ok {
			eventData["analysis_id"] = analysis.ID.String()
			eventData["url"] = analysis.URL
		}

	case domain.EventTypeProgress:
		eventData["data"] = event.Data

	case domain.EventTypeCompleted:
		eventData["message"] = "HTMLParser completed"
		if analysis, ok := event.Data.(*domain.Analysis); ok {
			eventData["analysis_id"] = analysis.ID.String()
			eventData["results"] = analysis.Results
		}

	case domain.EventTypeFailed:
		eventData["message"] = "HTMLParser failed"
		if analysis, ok := event.Data.(*domain.Analysis); ok {
			eventData["analysis_id"] = analysis.ID.String()
			eventData["error"] = analysis.Error
		}

	default:
		eventData["data"] = event.Data
	}

	h.writeSSEEvent(w, "analysis_event", eventData)
}

func (h *SSEHandlers) writeSSEError(w http.ResponseWriter, errorCode, message string) {
	errorData := map[string]interface{}{
		"error":     errorCode,
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	h.writeSSEEvent(w, "error", errorData)

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	h.logger.Warn().
		Str("error_code", errorCode).
		Str("message", message).
		Msg("SSE error event sent")
}
