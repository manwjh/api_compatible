package capture

import (
	"context"
	"log"
	"time"

	"corpus-tap/internal/config"
	"corpus-tap/internal/redact"
	"corpus-tap/internal/rules"
	"corpus-tap/internal/store"

	"github.com/google/uuid"
)

type Recorder struct {
	cfg          config.Config
	pg           *store.Postgres
	deploymentID uuid.UUID
}

func NewRecorder(cfg config.Config, pg *store.Postgres, deploymentID uuid.UUID) *Recorder {
	return &Recorder{cfg: cfg, pg: pg, deploymentID: deploymentID}
}

type Record struct {
	TapRequestID string
	UserID       int
	TokenID      int
	SessionKey   string
	Endpoint     string
	Wire         string
	IsStream     bool
	StatusCode   int
	LatencyMS    int
	ModelHeader  string
	ClientBody   []byte
	ResponseBody []byte
	Truncated    bool
	SkipReason   string
}

func (r *Recorder) Persist(ctx context.Context, rec Record) {
	if rec.SkipReason != "" {
		r.insertMeta(ctx, rec, "", "", "", "", "", rec.SkipReason, "")
		return
	}
	deploy := r.cfg.DeploymentID
	if deploy == "" {
		deploy = r.deploymentID.String()
	}
	exID := uuid.New()
	client := redact.Body(rec.ClientBody)
	resp := redact.Body(rec.ResponseBody)

	var reqURI, respURI, asmURI, reqSHA, respSHA string
	var storeErr string

	if r.cfg.LocalDataDir != "" {
		if b, err := store.WriteLocalGzip(r.cfg.LocalDataDir, deploy, rec.UserID, exID.String(), "client_request", client); err != nil {
			storeErr = err.Error()
		} else {
			reqURI, reqSHA = b.URI, b.SHA256
		}
		if len(resp) > 0 {
			role := "upstream_response"
			if rec.IsStream {
				role = "assembled_stream"
				asmURI = ""
			}
			if b, err := store.WriteLocalGzip(r.cfg.LocalDataDir, deploy, rec.UserID, exID.String(), role, resp); err != nil {
				if storeErr == "" {
					storeErr = err.Error()
				}
			} else {
				if rec.IsStream {
					asmURI = b.URI
				} else {
					respURI, respSHA = b.URI, b.SHA256
				}
			}
		}
	}

	r.insertMeta(ctx, rec, reqURI, respURI, asmURI, reqSHA, respSHA, "", storeErr)
}

func (r *Recorder) insertMeta(ctx context.Context, rec Record, reqURI, respURI, asmURI, reqSHA, respSHA, skipped, storeErr string) {
	if r.pg == nil {
		if storeErr != "" {
			log.Printf("corpus-tap: store user=%d tap_id=%s err=%s", rec.UserID, rec.TapRequestID, storeErr)
		}
		return
	}
	var dep *uuid.UUID
	if r.deploymentID != uuid.Nil {
		dep = &r.deploymentID
	}
	row := store.ExchangeRow{
		ID:                     uuid.New(),
		DeploymentID:           dep,
		UserID:                 rec.UserID,
		TapRequestID:             rec.TapRequestID,
		SessionKey:             rec.SessionKey,
		Endpoint:               rec.Endpoint,
		Wire:                   rec.Wire,
		IsStream:               rec.IsStream,
		StatusCode:             rec.StatusCode,
		LatencyMS:              rec.LatencyMS,
		ModelHeader:            rec.ModelHeader,
		ClientRequestURI:       reqURI,
		UpstreamResponseURI:    respURI,
		AssembledStreamURI:     asmURI,
		ClientRequestSHA256:    reqSHA,
		UpstreamResponseSHA256: respSHA,
		Truncated:              rec.Truncated,
		SkippedReason:          skipped,
		StoreError:             storeErr,
	}
	if rec.TokenID > 0 {
		row.TokenID = &rec.TokenID
	}
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := r.pg.InsertExchange(ctx2, row); err != nil {
		log.Printf("corpus-tap: pg insert: %v", err)
	}
}

func ModelFromBody(path string, body []byte) string {
	_ = path
	_ = body
	return ""
}

func Wire(path string) string { return rules.WireFormat(path) }
