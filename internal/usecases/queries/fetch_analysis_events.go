package queries

import (
	"context"

	"github.com/architeacher/svc-web-analyzer/internal/domain"
	"github.com/architeacher/svc-web-analyzer/internal/infrastructure"
	"github.com/architeacher/svc-web-analyzer/internal/service"
	"github.com/architeacher/svc-web-analyzer/internal/shared/decorator"
	"go.opentelemetry.io/otel/trace"
)

type (
	FetchAnalysisEventsQuery struct {
		AnalysisID string
	}

	FetchAnalysisEventsQueryHandler decorator.QueryHandler[FetchAnalysisEventsQuery, <-chan domain.AnalysisEvent]

	fetchAnalysisWithEventsQueryHandler struct {
		appService service.ApplicationService
	}
)

func NewFetchAnalysisEventsQueryHandler(
	appService service.ApplicationService,
	logger *infrastructure.Logger,
	tracerProvider trace.TracerProvider,
	metricsClient decorator.MetricsClient,
) decorator.QueryHandler[FetchAnalysisEventsQuery, <-chan domain.AnalysisEvent] {
	return decorator.ApplyQueryDecorators[FetchAnalysisEventsQuery, <-chan domain.AnalysisEvent](
		fetchAnalysisWithEventsQueryHandler{
			appService: appService,
		},
		logger,
		tracerProvider,
		metricsClient,
	)
}

func (h fetchAnalysisWithEventsQueryHandler) Execute(ctx context.Context, q FetchAnalysisEventsQuery) (<-chan domain.AnalysisEvent, error) {
	return h.appService.FetchAnalysisEvents(ctx, q.AnalysisID)
}
