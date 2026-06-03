package proxy

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"corpus-tap/internal/capture"
	"corpus-tap/internal/config"
	"corpus-tap/internal/enrich"
	"corpus-tap/internal/rules"

	"github.com/google/uuid"
)

type Server struct {
	cfg      config.Config
	upstream *url.URL
	proxy    *httputil.ReverseProxy
	recorder *capture.Recorder
	client   *http.Client
}

func New(cfg config.Config, recorder *capture.Recorder) (*Server, error) {
	u, err := url.Parse(cfg.Upstream)
	if err != nil {
		return nil, err
	}
	rp := httputil.NewSingleHostReverseProxy(u)
	rp.FlushInterval = -1
	return &Server{
		cfg:      cfg,
		upstream: u,
		proxy:    rp,
		recorder: recorder,
		client: &http.Client{
			Timeout: 0,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthz":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		case "/readyz":
			s.ready(w, r)
			return
		}

		if s.cfg.ProxyOnly || !rules.ShouldCapture(r) {
			s.proxy.ServeHTTP(w, r)
			return
		}

		s.handleCapture(w, r)
	})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.upstream.String()+"/api/status", nil)
	if err != nil {
		http.Error(w, "upstream config error", http.StatusServiceUnavailable)
		return
	}
	resp, err := s.client.Do(req)
	if err != nil {
		http.Error(w, "upstream unreachable", http.StatusServiceUnavailable)
		return
	}
	_ = resp.Body.Close()
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleCapture(w http.ResponseWriter, r *http.Request) {
	tapID := uuid.New().String()
	start := time.Now()

	subject := enrich.Resolve(r, s.cfg.DevUserID, s.cfg.DenyUserIDs)
	rec := capture.Record{
		TapRequestID: tapID,
		Endpoint:     r.URL.Path,
		Wire:         capture.Wire(r.URL.Path),
		SessionKey:   enrich.SessionKey(r, tapID),
	}

	if !subject.OK {
		rec.SkipReason = "enrich_failed"
		s.forwardAndRecord(w, r, nil, rec, start)
		return
	}
	rec.UserID = subject.UserID
	rec.TokenID = subject.TokenID

	body, truncated, err := readBody(r.Body, s.cfg.MaxBodyBytes)
	if err != nil {
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}
	rec.Truncated = truncated
	rec.ClientBody = body
	rec.IsStream = rules.IsStreamRequest(r, body)
	r.Body = io.NopCloser(bytes.NewReader(body))

	upReq, err := http.NewRequestWithContext(r.Context(), r.Method, s.upstream.ResolveReference(r.URL).String(), bytes.NewReader(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	copyHeaders(upReq.Header, r.Header)
	upReq.ContentLength = int64(len(body))

	upResp, err := s.client.Do(upReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		rec.SkipReason = "upstream_error"
		go s.recorder.Persist(context.Background(), rec)
		return
	}
	defer upResp.Body.Close()

	respBody, err := io.ReadAll(upResp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	rec.ResponseBody = respBody
	rec.StatusCode = upResp.StatusCode
	rec.LatencyMS = int(time.Since(start).Milliseconds())

	copyHeaders(w.Header(), upResp.Header)
	w.WriteHeader(upResp.StatusCode)
	_, _ = w.Write(respBody)

	go s.recorder.Persist(context.Background(), rec)
}

func (s *Server) forwardAndRecord(w http.ResponseWriter, r *http.Request, body []byte, rec capture.Record, start time.Time) {
	if body != nil {
		r.Body = io.NopCloser(bytes.NewReader(body))
	}
	s.proxy.ServeHTTP(w, r)
	rec.LatencyMS = int(time.Since(start).Milliseconds())
	go s.recorder.Persist(context.Background(), rec)
}

func readBody(rc io.ReadCloser, max int64) ([]byte, bool, error) {
	defer rc.Close()
	var buf bytes.Buffer
	lr := io.LimitReader(rc, max+1)
	n, err := io.Copy(&buf, lr)
	if err != nil {
		return nil, false, err
	}
	truncated := n > max
	if truncated {
		return buf.Bytes()[:max], true, nil
	}
	return buf.Bytes(), false, nil
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		if strings.EqualFold(k, "Connection") {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
