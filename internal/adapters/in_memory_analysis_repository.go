package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/architeacher/svc-web-analyzer/internal/domain"
	"github.com/architeacher/svc-web-analyzer/internal/infrastructure"
	"github.com/google/uuid"
)

type InMemoryAnalysisRepository struct {
	data   map[uuid.UUID]*domain.Analysis
	logger *infrastructure.Logger
}

func NewInMemoryAnalysisRepository(logger *infrastructure.Logger) *InMemoryAnalysisRepository {
	return &InMemoryAnalysisRepository{
		data:   make(map[uuid.UUID]*domain.Analysis),
		logger: logger,
	}
}

func (r *InMemoryAnalysisRepository) Create(ctx context.Context, analysis *domain.Analysis) error {
	if analysis.ID == uuid.Nil {
		return fmt.Errorf("analysis ID cannot be nil")
	}

	if _, exists := r.data[analysis.ID]; exists {
		return fmt.Errorf("analysis with ID %s already exists", analysis.ID.String())
	}

	// Deep copy to avoid reference issues
	analysisCopy := r.deepCopyAnalysis(analysis)
	r.data[analysis.ID] = analysisCopy

	r.logger.Debug().
		Str("id", analysis.ID.String()).
		Str("url", analysis.URL).
		Msg("Analysis created in repository")

	return nil
}

func (r *InMemoryAnalysisRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Analysis, error) {
	analysis, exists := r.data[id]
	if !exists {
		return nil, domain.ErrAnalysisNotFound
	}

	// Return a deep copy to avoid mutation
	return r.deepCopyAnalysis(analysis), nil
}

func (r *InMemoryAnalysisRepository) Update(ctx context.Context, analysis *domain.Analysis) error {
	if analysis.ID == uuid.Nil {
		return fmt.Errorf("analysis ID cannot be nil")
	}

	if _, exists := r.data[analysis.ID]; !exists {
		return domain.ErrAnalysisNotFound
	}

	// Deep copy to avoid reference issues
	analysisCopy := r.deepCopyAnalysis(analysis)
	r.data[analysis.ID] = analysisCopy

	r.logger.Debug().
		Str("id", analysis.ID.String()).
		Str("status", string(analysis.Status)).
		Msg("Analysis updated in repository")

	return nil
}

func (r *InMemoryAnalysisRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, exists := r.data[id]; !exists {
		return domain.ErrAnalysisNotFound
	}

	delete(r.data, id)

	r.logger.Debug().Str("id", id.String()).Msg("Analysis deleted from repository")

	return nil
}

func (r *InMemoryAnalysisRepository) List(ctx context.Context, limit, offset int) ([]*domain.Analysis, error) {
	analyses := make([]*domain.Analysis, 0, len(r.data))

	for _, analysis := range r.data {
		analyses = append(analyses, r.deepCopyAnalysis(analysis))
	}

	// Simple pagination
	start := offset
	if start > len(analyses) {
		return []*domain.Analysis{}, nil
	}

	end := start + limit
	if end > len(analyses) {
		end = len(analyses)
	}

	return analyses[start:end], nil
}

func (r *InMemoryAnalysisRepository) GetByStatus(ctx context.Context, status domain.AnalysisStatus) ([]*domain.Analysis, error) {
	var analyses []*domain.Analysis

	for _, analysis := range r.data {
		if analysis.Status == status {
			analyses = append(analyses, r.deepCopyAnalysis(analysis))
		}
	}

	return analyses, nil
}

func (r *InMemoryAnalysisRepository) GetByURL(ctx context.Context, url string) ([]*domain.Analysis, error) {
	var analyses []*domain.Analysis

	for _, analysis := range r.data {
		if analysis.URL == url {
			analyses = append(analyses, r.deepCopyAnalysis(analysis))
		}
	}

	return analyses, nil
}

func (r *InMemoryAnalysisRepository) Count(ctx context.Context) (int, error) {
	return len(r.data), nil
}

func (r *InMemoryAnalysisRepository) CountByStatus(ctx context.Context, status domain.AnalysisStatus) (int, error) {
	count := 0
	for _, analysis := range r.data {
		if analysis.Status == status {
			count++
		}
	}
	return count, nil
}

// Clean up completed analyses older than the specified duration
func (r *InMemoryAnalysisRepository) CleanupOldAnalyses(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)
	deleted := 0

	for id, analysis := range r.data {
		if analysis.Status == domain.StatusCompleted &&
			analysis.CompletedAt != nil &&
			analysis.CompletedAt.Before(cutoff) {
			delete(r.data, id)
			deleted++
		}
	}

	r.logger.Info().
		Int("deleted_count", deleted).
		Time("cutoff_time", cutoff).
		Msg("Cleaned up old analyses")

	return deleted, nil
}

// Helper method to deep copy an analysis to prevent reference issues
func (r *InMemoryAnalysisRepository) deepCopyAnalysis(analysis *domain.Analysis) *domain.Analysis {
	// Use JSON marshal/unmarshal for deep copy
	// This is not the most efficient but ensures complete independence
	data, err := json.Marshal(analysis)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to marshal analysis for deep copy")
		return analysis // fallback to shallow copy
	}

	var copy domain.Analysis
	if err := json.Unmarshal(data, &copy); err != nil {
		r.logger.Error().Err(err).Msg("Failed to unmarshal analysis for deep copy")
		return analysis // fallback to shallow copy
	}

	return &copy
}

// GetStats returns repository statistics
func (r *InMemoryAnalysisRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	total, _ := r.Count(ctx)
	requested, _ := r.CountByStatus(ctx, domain.StatusRequested)
	inProgress, _ := r.CountByStatus(ctx, domain.StatusInProgress)
	completed, _ := r.CountByStatus(ctx, domain.StatusCompleted)
	failed, _ := r.CountByStatus(ctx, domain.StatusFailed)

	stats["total"] = total
	stats["by_status"] = map[string]int{
		"requested":   requested,
		"in_progress": inProgress,
		"completed":   completed,
		"failed":      failed,
	}

	return stats, nil
}
