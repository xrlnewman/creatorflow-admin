package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterContentPipelineEnvelopeAndValidation(t *testing.T) {
	r := NewRouter(NewMemoryStore(), newMemoryIdempotency())
	create := func(body, key string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/content-items", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)
		return res
	}
	if res := create(`{"channel":"短视频","owner":"林编辑","plannedAt":"2026-07-18"}`, "handler-invalid"); res.Code != http.StatusBadRequest {
		t.Fatalf("missing title status = %d, body=%s", res.Code, res.Body.String())
	}
	res := create(`{"title":"城市夜行","channel":"短视频","owner":"林编辑","plannedAt":"2026-07-18T09:00:00+08:00"}`, "handler-create")
	if res.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body=%s", res.Code, res.Body.String())
	}
	var envelope struct {
		Code    int         `json:"code"`
		TraceID string      `json:"traceId"`
		Data    ContentItem `json:"data"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Code != 0 || envelope.TraceID == "" || envelope.Data.ID == "" {
		t.Fatalf("bad envelope: %+v", envelope)
	}
	duplicate := create(`{"title":"城市夜行","channel":"短视频","owner":"林编辑","plannedAt":"2026-07-18T09:00:00+08:00"}`, "handler-create")
	var duplicateEnvelope struct {
		Data ContentItem `json:"data"`
	}
	if err := json.Unmarshal(duplicate.Body.Bytes(), &duplicateEnvelope); err != nil {
		t.Fatal(err)
	}
	if duplicateEnvelope.Data.ID != envelope.Data.ID {
		t.Fatalf("duplicate id = %q, want %q", duplicateEnvelope.Data.ID, envelope.Data.ID)
	}

	id := envelope.Data.ID
	post := func(path, body, key string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/content-items/"+id+path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		out := httptest.NewRecorder()
		r.ServeHTTP(out, req)
		return out
	}
	if out := post("/publish", `{"publishedAt":"2026-07-18T18:00:00+08:00","actor":"主编"}`, "publish-early"); out.Code != http.StatusConflict {
		t.Fatalf("early publish status = %d, body=%s", out.Code, out.Body.String())
	}
	if out := post("/script", `{"body":"人物在夜色中完成一次城市漫游。"}`, "handler-script-1"); out.Code != http.StatusOK {
		t.Fatalf("script status = %d, body=%s", out.Code, out.Body.String())
	}
	if out := post("/script", `{"body":"人物在夜色中完成一次城市漫游，加入旁白。"}`, "handler-script-2"); out.Code != http.StatusOK {
		t.Fatalf("production status = %d, body=%s", out.Code, out.Body.String())
	}
	if out := post("/submit-review", `{"actor":"主编"}`, "handler-review"); out.Code != http.StatusOK {
		t.Fatalf("review status = %d, body=%s", out.Code, out.Body.String())
	}
	if out := post("/publish", `{"publishedAt":"2026-07-18T18:00:00+08:00","actor":"主编"}`, "handler-publish"); out.Code != http.StatusOK {
		t.Fatalf("publish status = %d, body=%s", out.Code, out.Body.String())
	}
	if out := post("/metrics", `{"views":-1,"likes":0,"comments":0,"shares":0}`, "handler-metrics-invalid"); out.Code != http.StatusBadRequest {
		t.Fatalf("negative metrics status = %d, body=%s", out.Code, out.Body.String())
	}
	if out := post("/metrics", `{"views":12000,"likes":680,"comments":42,"shares":93}`, "handler-metrics"); out.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, body=%s", out.Code, out.Body.String())
	}
	get := httptest.NewRecorder()
	r.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/v1/content-items/"+id, nil))
	if get.Code != http.StatusOK || !bytes.Contains(get.Body.Bytes(), []byte("已复盘")) || !bytes.Contains(get.Body.Bytes(), []byte("城市夜行")) {
		t.Fatalf("detail response = %d, body=%s", get.Code, get.Body.String())
	}
	events := httptest.NewRecorder()
	r.ServeHTTP(events, httptest.NewRequest(http.MethodGet, "/api/v1/content-items/"+id+"/events", nil))
	if events.Code != http.StatusOK || !bytes.Contains(events.Body.Bytes(), []byte("record_metrics")) {
		t.Fatalf("events response = %d, body=%s", events.Code, events.Body.String())
	}
	list := httptest.NewRecorder()
	r.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/v1/content-items?status=已复盘&owner=林编辑&page=1&pageSize=20", nil))
	if list.Code != http.StatusOK || !bytes.Contains(list.Body.Bytes(), []byte(id)) {
		t.Fatalf("filtered list = %d, body=%s", list.Code, list.Body.String())
	}
}
