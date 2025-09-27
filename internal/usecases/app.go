package usecases

import (
	"github.com/architeacher/svc-web-analyzer/internal/infrastructure"
	"github.com/architeacher/svc-web-analyzer/internal/service"
	"github.com/architeacher/svc-web-analyzer/internal/shared/decorator"
	"github.com/architeacher/svc-web-analyzer/internal/usecases/commands"
	"github.com/architeacher/svc-web-analyzer/internal/usecases/queries"
	otelTrace "go.opentelemetry.io/otel/trace"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	AnalyzeCommandHandler commands.AnalyzeCommandHandler
}

type Queries struct {
	FetchAnalysisQueryHandler        queries.FetchAnalysisQueryHandler
	FetchAnalysisEventsQueryHandler  queries.FetchAnalysisEventsQueryHandler
	FetchReadinessReportQueryHandler queries.FetchReadinessReportQueryHandler
	FetchLivenessReportQueryHandler  queries.FetchLivenessReportQueryHandler
	FetchHealthReportQueryHandler    queries.FetchHealthReportQueryHandler
}

func NewApplication(
	appService service.ApplicationService,
	logger *infrastructure.Logger,
	tracerProvider otelTrace.TracerProvider,
	metricsClient decorator.MetricsClient,
) Application {
	return Application{
		Commands: Commands{
			AnalyzeCommandHandler: commands.NewAnalyzeCommandHandler(appService, logger, tracerProvider, metricsClient),
		},
		Queries: Queries{
			FetchAnalysisQueryHandler:        queries.NewFetchAnalysisQueryHandler(appService, logger, tracerProvider, metricsClient),
			FetchAnalysisEventsQueryHandler:  queries.NewFetchAnalysisEventsQueryHandler(appService, logger, tracerProvider, metricsClient),
			FetchReadinessReportQueryHandler: queries.NewFetchReadinessReportQueryHandler(appService, logger, tracerProvider, metricsClient),
			FetchLivenessReportQueryHandler:  queries.NewFetchLivenessReportQueryHandler(appService, logger, tracerProvider, metricsClient),
			FetchHealthReportQueryHandler:    queries.NewFetchHealthReportQueryHandler(appService, logger, tracerProvider, metricsClient),
		},
	}
}
