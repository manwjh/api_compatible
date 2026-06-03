package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExchangeRow struct {
	ID                  uuid.UUID
	DeploymentID        *uuid.UUID
	UserID              int
	TokenID             *int
	TapRequestID        string
	SessionKey          string
	Endpoint            string
	Wire                string
	IsStream            bool
	StatusCode          int
	LatencyMS           int
	ModelHeader         string
	ClientRequestURI    string
	UpstreamResponseURI string
	AssembledStreamURI  string
	ClientRequestSHA256 string
	UpstreamResponseSHA256 string
	Truncated           bool
	SkippedReason       string
	StoreError          string
}

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, databaseURL string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() { p.pool.Close() }

func (p *Postgres) EnsureDeployment(ctx context.Context, newapiImage, tapImage string) (uuid.UUID, error) {
	var id uuid.UUID
	err := p.pool.QueryRow(ctx,
		`INSERT INTO tap_deployment (newapi_image, tap_image) VALUES ($1, $2) RETURNING id`,
		newapiImage, tapImage,
	).Scan(&id)
	return id, err
}

func (p *Postgres) InsertExchange(ctx context.Context, row ExchangeRow) error {
	_, err := p.pool.Exec(ctx, `
INSERT INTO http_exchange (
  id, deployment_id, user_id, token_id, tap_request_id, session_key,
  endpoint, wire, is_stream, status_code, latency_ms, model_header,
  client_request_uri, upstream_response_uri, assembled_stream_uri,
  client_request_sha256, upstream_response_sha256,
  truncated, skipped_reason, store_error
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20
)`,
		row.ID, row.DeploymentID, row.UserID, row.TokenID, row.TapRequestID, row.SessionKey,
		row.Endpoint, row.Wire, row.IsStream, row.StatusCode, row.LatencyMS, row.ModelHeader,
		row.ClientRequestURI, row.UpstreamResponseURI, row.AssembledStreamURI,
		row.ClientRequestSHA256, row.UpstreamResponseSHA256,
		row.Truncated, nullString(row.SkippedReason), nullString(row.StoreError),
	)
	return err
}

func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (p *Postgres) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return p.pool.Ping(ctx)
}
