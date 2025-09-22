package adapters

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/architeacher/svc-web-analyzer/internal/handlers"
	"github.com/architeacher/svc-web-analyzer/internal/infrastructure"
	"github.com/architeacher/svc-web-analyzer/internal/usecases"
	"github.com/architeacher/svc-web-analyzer/internal/usecases/commands"
	"github.com/architeacher/svc-web-analyzer/internal/usecases/queries"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type RequestHandler struct {
	logger *infrastructure.Logger
	app    usecases.Application
}

func NewRequestHandler(
	a usecases.Application,
) *RequestHandler {
	return &RequestHandler{
		app: a,
	}
}

// AnalyzeURL implements ServerInterface.AnalyzeURL
func (h *RequestHandler) AnalyzeURL(w http.ResponseWriter, r *http.Request, params handlers.AnalyzeURLParams) {
	// Parse request body
	var req handlers.AnalyzeURLJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "bad_request", "Invalid request body", err.Error())
		return
	}

	// Validate required URL field
	if req.Url == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "bad_request", "URL is required", "url field cannot be empty")
		return
	}

	// Execute command
	result, err := h.app.Commands.AnalyzeCommandHandler.Handle(
		r.Context(),
		commands.AnalyzeCommand{
			URL: req.Url,
		},
	)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "internal_server_error", "Failed to start analysis", err.Error())
		return
	}

	// Write success response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Version", "v1")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(result)
}

// GetAnalysis implements ServerInterface.GetAnalysis
func (h *RequestHandler) GetAnalysis(w http.ResponseWriter, r *http.Request, analysisId openapi_types.UUID, params handlers.GetAnalysisParams) {
	// Execute query
	result, err := h.app.Queries.FetchAnalysisQueryHandler.Execute(
		r.Context(),
		queries.FetchAnalysisQuery{AnalysisID: analysisId.String()},
	)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "not_found", "Analysis not found", err.Error())
		return
	}

	// Write response based on analysis status
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Version", "v1")

	// Determine response status based on analysis state
	if result != nil {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}

	json.NewEncoder(w).Encode(result)
}

// GetAnalysisEvents implements ServerInterface.GetAnalysisEvents
func (h *RequestHandler) GetAnalysisEvents(w http.ResponseWriter, r *http.Request, analysisId openapi_types.UUID, params handlers.GetAnalysisEventsParams) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("API-Version", "v1")
	w.WriteHeader(http.StatusOK)

	// Execute SSE query
	_, err := h.app.Queries.FetchAnalysisEventsQueryHandler.Execute(
		r.Context(),
		queries.FetchAnalysisEventsQuery{AnalysisID: analysisId.String()},
	)
	if err != nil {
		// Write error as SSE event
		w.Write([]byte("event: error\n"))
		w.Write([]byte("data: {\"error\": \"Failed to fetch events\"}\n\n"))
	}
}
func (h *RequestHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now()

	// Tell application service to fetch readiness report
	readinessResult, err := h.app.Queries.FetchReadinessReportQueryHandler.Execute(
		ctx,
		queries.FetchReadinessReportQuery{},
	)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "internal_server_error", "Failed to check readiness", err.Error())
		return
	}

	// Create the readiness response with real data
	readinessResp := handlers.ReadinessResponse{
		Status:    readinessResult.OverallStatus,
		Timestamp: now,
		Version:   stringPtr("1.0.0"),
		Checks: handlers.ReadinessResponse_Checks{
			Storage: &struct {
				Status handlers.ReadinessResponseChecksStorageStatus `json:"status"`
			}{
				Status: handlers.ReadinessResponseChecksStorageStatus(readinessResult.Storage.Status),
			},
			Cache: &struct {
				Status handlers.ReadinessResponseChecksCacheStatus `json:"status"`
			}{
				Status: handlers.ReadinessResponseChecksCacheStatus(readinessResult.Cache.Status),
			},
			Queue: &struct {
				Status handlers.ReadinessResponseChecksQueueStatus `json:"status"`
			}{
				Status: handlers.ReadinessResponseChecksQueueStatus(readinessResult.Queue.Status),
			},
		},
	}

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if readinessResult.OverallStatus == handlers.DOWN {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(readinessResp)
}

func (h *RequestHandler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Tell application service to fetch liveness report
	livenessResult, err := h.app.Queries.FetchLivenessReportQueryHandler.Execute(
		ctx,
		queries.FetchLivenessReportQuery{},
	)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "internal_server_error", "Failed to check liveness", err.Error())
		return
	}

	livenessResp := handlers.LivenessResponse{
		Status:    livenessResult.OverallStatus,
		Timestamp: time.Now(),
	}

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if livenessResult.OverallStatus == handlers.LivenessResponseStatusDOWN {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(livenessResp)
}

// HealthCheck implements ServerInterface.HealthCheck
func (h *RequestHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now()

	// Tell application service to fetch health report
	healthResult, err := h.app.Queries.FetchHealthReportQueryHandler.Execute(
		ctx,
		queries.FetchHealthReportQuery{},
	)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "internal_server_error", "Failed to check health", err.Error())
		return
	}

	// Create health response with real data
	healthResp := handlers.HealthResponse{
		Status:    healthResult.OverallStatus,
		Timestamp: now,
		Version:   stringPtr("1.0.0"),
		Uptime:    &healthResult.Uptime,
		Checks: handlers.HealthResponse_Checks{
			Storage: &struct {
				Details      *map[string]interface{}                    `json:"details,omitempty"`
				Error        *string                                    `json:"error,omitempty"`
				LastChecked  *time.Time                                 `json:"last_checked,omitempty"`
				ResponseTime *float32                                   `json:"response_time,omitempty"`
				Status       handlers.HealthResponseChecksStorageStatus `json:"status"`
			}{
				Status:       handlers.HealthResponseChecksStorageStatus(healthResult.Storage.Status),
				ResponseTime: &healthResult.Storage.ResponseTime,
				LastChecked:  &healthResult.Storage.LastChecked,
				Error: func() *string {
					if healthResult.Storage.Error != "" {
						return &healthResult.Storage.Error
					} else {
						return nil
					}
				}(),
			},
			Cache: &struct {
				Details      *map[string]interface{}                  `json:"details,omitempty"`
				Error        *string                                  `json:"error,omitempty"`
				LastChecked  *time.Time                               `json:"last_checked,omitempty"`
				ResponseTime *float32                                 `json:"response_time,omitempty"`
				Status       handlers.HealthResponseChecksCacheStatus `json:"status"`
			}{
				Status:       handlers.HealthResponseChecksCacheStatus(healthResult.Cache.Status),
				ResponseTime: &healthResult.Cache.ResponseTime,
				LastChecked:  &healthResult.Cache.LastChecked,
				Error: func() *string {
					if healthResult.Cache.Error != "" {
						return &healthResult.Cache.Error
					} else {
						return nil
					}
				}(),
			},
			Queue: &struct {
				Details      *map[string]interface{}                  `json:"details,omitempty"`
				Error        *string                                  `json:"error,omitempty"`
				LastChecked  *time.Time                               `json:"last_checked,omitempty"`
				ResponseTime *float32                                 `json:"response_time,omitempty"`
				Status       handlers.HealthResponseChecksQueueStatus `json:"status"`
			}{
				Status:       handlers.HealthResponseChecksQueueStatus(healthResult.Queue.Status),
				ResponseTime: &healthResult.Queue.ResponseTime,
				LastChecked:  &healthResult.Queue.LastChecked,
				Error: func() *string {
					if healthResult.Queue.Error != "" {
						return &healthResult.Queue.Error
					} else {
						return nil
					}
				}(),
			},
		},
	}

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if healthResult.OverallStatus == handlers.HealthResponseStatusDOWN {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(healthResp)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func float32Ptr(f float32) *float32 {
	return &f
}

// writeErrorResponse writes a standardized error response
func (h *RequestHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorType, message, details string) {
	errorResp := handlers.ErrorResponse{
		Error:      &errorType,
		Message:    &message,
		Details:    &details,
		StatusCode: &statusCode,
		Timestamp:  &[]time.Time{time.Now()}[0],
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResp)
}
