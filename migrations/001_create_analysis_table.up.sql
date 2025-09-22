-- UUID extension is enabled in 000_create_database.up.sql

-- Create enum types for better type safety
CREATE TYPE analysis_status AS ENUM ('requested', 'in_progress', 'completed', 'failed');

-- Main analysis table with versioning support
CREATE TABLE analysis (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    url TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    content_hash VARCHAR(64), -- SHA-256 hash of page content for deduplication
    status analysis_status NOT NULL DEFAULT 'requested',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    duration_ms BIGINT, -- Duration in milliseconds

    -- Analysis results as JSON (PostgreSQL has excellent JSON support)
    results JSONB,

    -- Error information
    error_code TEXT,
    error_message TEXT,
    error_status_code INTEGER,
    error_details TEXT,

    -- Metadata
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure unique URL + version combination
    CONSTRAINT uk_analysis_url_version UNIQUE (url, version)
);

-- Create indexes for performance
CREATE INDEX idx_analysis_status ON analysis(status);
CREATE INDEX idx_analysis_created_at ON analysis(created_at);
CREATE INDEX idx_analysis_status_created ON analysis(status, created_at);

-- Index for finding latest version of a URL
CREATE INDEX idx_analysis_url_version_desc ON analysis(url, version DESC);

-- Index for version-specific queries
CREATE INDEX idx_analysis_version ON analysis(version);

-- Index for content hash deduplication
CREATE INDEX idx_analysis_content_hash ON analysis(content_hash) WHERE content_hash IS NOT NULL;

-- JSON indexes for common queries on results
CREATE INDEX idx_analysis_results_gin ON analysis USING GIN(results);
CREATE INDEX idx_analysis_html_version ON analysis((results->>'html_version')) WHERE results IS NOT NULL;

-- Partial indexes for performance
CREATE INDEX idx_analysis_completed ON analysis(completed_at) WHERE status = 'completed';
CREATE INDEX idx_analysis_failed ON analysis(created_at) WHERE status = 'failed';

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Function to automatically set version for new URL analyses
CREATE OR REPLACE FUNCTION set_analysis_version()
RETURNS TRIGGER AS $$
BEGIN
    -- If no version specified, get the next version for this URL
    IF NEW.version IS NULL OR NEW.version = 1 THEN
        SELECT COALESCE(MAX(version), 0) + 1
        INTO NEW.version
        FROM analysis
        WHERE url = NEW.url;
    END IF;

    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to automatically update updated_at
CREATE TRIGGER update_analysis_updated_at
    BEFORE UPDATE ON analysis
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger to automatically set version
CREATE TRIGGER set_analysis_version
    BEFORE INSERT ON analysis
    FOR EACH ROW
    EXECUTE FUNCTION set_analysis_version();

-- Add comments for documentation
COMMENT ON TABLE analysis IS 'Web page analysis results storage with versioning support';
COMMENT ON COLUMN analysis.url IS 'The URL being analyzed';
COMMENT ON COLUMN analysis.version IS 'Version number for the same URL (auto-incremented)';
COMMENT ON COLUMN analysis.content_hash IS 'SHA-256 hash of page content for deduplication';
COMMENT ON COLUMN analysis.results IS 'JSON structure containing analysis data including HTML version, title, heading counts, links, and forms';
COMMENT ON COLUMN analysis.duration_ms IS 'Analysis duration in milliseconds';
COMMENT ON CONSTRAINT uk_analysis_url_version ON analysis IS 'Ensures unique URL + version combinations';
COMMENT ON INDEX idx_analysis_results_gin IS 'GIN index for efficient JSON queries on results column';
COMMENT ON INDEX idx_analysis_url_version_desc IS 'Index for finding latest version of a URL efficiently';
COMMENT ON INDEX idx_analysis_content_hash IS 'Index for content hash-based deduplication queries';