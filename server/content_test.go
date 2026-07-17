package main

import (
	"context"
	"errors"
	"testing"
)

func TestContentPipelineLifecycleAndEventOrder(t *testing.T) {
	store := NewMemoryStore()
	svc := NewCareService(store, newMemoryIdempotency())
	ctx := context.Background()

	item, err := svc.CreateContentItem(ctx, CreateContentItemInput{
		Title: "春日通勤短片", Channel: "短视频", Owner: "林编辑", PlannedAt: "2026-07-18T09:00:00+08:00",
	}, "content-create-1")
	if err != nil {
		t.Fatal(err)
	}
	if item.Status != ContentStatusTopic {
		t.Fatalf("initial status = %q", item.Status)
	}
	if _, err := svc.SubmitContentReview(ctx, item.ID, "主编", "review-before-script"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("review before script error = %v", err)
	}

	item, err = svc.SaveContentScript(ctx, item.ID, SaveContentScriptInput{Body: "镜头一：晨光穿过地铁站；镜头二：人物完成出发。"}, "script-1")
	if err != nil || item.Status != ContentStatusWriting {
		t.Fatalf("save script status = %q, err = %v", item.Status, err)
	}
	item, err = svc.SaveContentScript(ctx, item.ID, SaveContentScriptInput{Body: "镜头一：晨光穿过地铁站；镜头二：人物完成出发。"}, "script-2")
	if err != nil || item.Status != ContentStatusProducing {
		t.Fatalf("start production status = %q, err = %v", item.Status, err)
	}
	item, err = svc.SubmitContentReview(ctx, item.ID, "主编", "review-1")
	if err != nil || item.Status != ContentStatusReviewing {
		t.Fatalf("submit review status = %q, err = %v", item.Status, err)
	}
	item, err = svc.PublishContent(ctx, item.ID, PublishContentInput{PublishedAt: "2026-07-18T18:00:00+08:00", Actor: "主编"}, "publish-1")
	if err != nil || item.Status != ContentStatusPublished {
		t.Fatalf("publish status = %q, err = %v", item.Status, err)
	}
	duplicate, err := svc.PublishContent(ctx, item.ID, PublishContentInput{PublishedAt: "2026-07-18T18:00:00+08:00", Actor: "主编"}, "publish-1")
	if err != nil || duplicate.ID != item.ID {
		t.Fatalf("idempotent publish = %+v, err = %v", duplicate, err)
	}
	item, err = svc.RecordContentMetrics(ctx, item.ID, RecordContentMetricsInput{Views: 12000, Likes: 680, Comments: 42, Shares: 93}, "metrics-1")
	if err != nil || item.Status != ContentStatusReviewed {
		t.Fatalf("metrics status = %q, err = %v", item.Status, err)
	}
	if item.Metrics == nil || item.Metrics.Views != 12000 {
		t.Fatalf("metrics = %+v", item.Metrics)
	}
	events, err := store.ListContentEvents(ctx, item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 6 {
		t.Fatalf("events = %d, want 6", len(events))
	}
	for i := 1; i < len(events); i++ {
		if events[i-1].CreatedAt > events[i].CreatedAt {
			t.Fatalf("events not ordered: %+v", events)
		}
	}
	if events[len(events)-1].Action != "record_metrics" || events[len(events)-1].ToStatus != ContentStatusReviewed {
		t.Fatalf("last event = %+v", events[len(events)-1])
	}
}

func TestContentItemValidationAndMetrics(t *testing.T) {
	store := NewMemoryStore()
	svc := NewCareService(store, newMemoryIdempotency())
	ctx := context.Background()
	for _, input := range []CreateContentItemInput{
		{Channel: "短视频", Owner: "林编辑", PlannedAt: "2026-07-18"},
		{Title: "缺渠道", Owner: "林编辑", PlannedAt: "2026-07-18"},
		{Title: "缺负责人", Channel: "短视频", PlannedAt: "2026-07-18"},
	} {
		if _, err := svc.CreateContentItem(ctx, input, "required-"+input.Title); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("input %+v error = %v", input, err)
		}
	}
	item, err := svc.CreateContentItem(ctx, CreateContentItemInput{Title: "指标校验", Channel: "图文专栏", Owner: "沈编辑", PlannedAt: "2026-07-18"}, "metrics-create")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RecordContentMetrics(ctx, item.ID, RecordContentMetricsInput{Views: -1}, "metrics-negative"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("negative metrics error = %v", err)
	}
	if _, err := svc.SaveContentScript(ctx, item.ID, SaveContentScriptInput{Body: ""}, "script-empty"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("empty script error = %v", err)
	}
}

func TestContentListFiltersByOwnerStatusAndPublishedAt(t *testing.T) {
	store := NewMemoryStore()
	items, total, err := store.ListContentItems(context.Background(), 1, 20, ContentStatusPublished, "林编辑", "", "2026-07-18")
	if err != nil {
		t.Fatal(err)
	}
	if total == 0 || len(items) == 0 {
		t.Fatalf("filtered items = %d/%d", len(items), total)
	}
	for _, item := range items {
		if item.Status != ContentStatusPublished || item.Owner != "林编辑" || item.Publish == nil || item.Publish.PublishedAt[:10] != "2026-07-18" {
			t.Fatalf("unexpected filtered item: %+v", item)
		}
	}
}
