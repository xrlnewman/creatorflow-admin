package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// CareService owns validation, idempotency and lifecycle rules for CreatorFlow.
type CareService struct {
	store CareStore
	idem  idempotencyStore
}

func NewCareService(store CareStore, idem idempotencyStore) *CareService {
	return &CareService{store: store, idem: idem}
}

func (s *CareService) CreateAppointment(ctx context.Context, input CreateAppointmentInput, key string) (Appointment, error) {
	if strings.TrimSpace(key) == "" {
		return Appointment{}, ErrMissingIdempotencyKey
	}
	if strings.TrimSpace(input.Patient) == "" && strings.TrimSpace(input.PatientID) == "" {
		return Appointment{}, fmt.Errorf("%w: patient is required", ErrInvalidInput)
	}
	resourceKey := "appointment:create:" + key
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return Appointment{}, err
	} else if ok {
		return s.store.GetAppointment(ctx, existing)
	}
	release, err := s.idem.Lock(ctx, "appointment:create-lock", 10*time.Second)
	if err != nil {
		return Appointment{}, err
	}
	defer release()
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return Appointment{}, err
	} else if ok {
		return s.store.GetAppointment(ctx, existing)
	}
	a, err := s.store.CreateAppointment(ctx, Appointment{PatientID: input.PatientID, Patient: input.Patient, Department: input.Department, Doctor: input.Doctor, ScheduledAt: input.ScheduledAt, Status: AppointmentPending})
	if err != nil {
		return Appointment{}, err
	}
	if err := s.idem.Set(ctx, resourceKey, a.ID, 24*time.Hour); err != nil {
		return Appointment{}, err
	}
	return a, nil
}

func (s *CareService) CheckinAppointment(ctx context.Context, id, key string) (Appointment, error) {
	return s.UpdateAppointmentStatus(ctx, id, AppointmentChecked, "前台", key)
}

func (s *CareService) CreateFollowup(ctx context.Context, input CreateFollowupInput, key string) (Followup, error) {
	if strings.TrimSpace(key) == "" {
		return Followup{}, ErrMissingIdempotencyKey
	}
	if strings.TrimSpace(input.Patient) == "" && strings.TrimSpace(input.PatientID) == "" {
		return Followup{}, fmt.Errorf("%w: patient is required", ErrInvalidInput)
	}
	resourceKey := "followup:create:" + key
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return Followup{}, err
	} else if ok {
		return findFollowup(ctx, s.store, existing)
	}
	release, err := s.idem.Lock(ctx, "followup:create-lock", 10*time.Second)
	if err != nil {
		return Followup{}, err
	}
	defer release()
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return Followup{}, err
	} else if ok {
		return findFollowup(ctx, s.store, existing)
	}
	f, err := s.store.CreateFollowup(ctx, Followup{PatientID: input.PatientID, Patient: input.Patient, Summary: input.Summary, DueAt: input.DueAt, Status: FollowupPending})
	if err != nil {
		return Followup{}, err
	}
	if err := s.idem.Set(ctx, resourceKey, f.ID, 24*time.Hour); err != nil {
		return Followup{}, err
	}
	return f, nil
}

func (s *CareService) UpdateAppointmentStatus(ctx context.Context, id, status string, args ...string) (Appointment, error) {
	actor, key := "运营人员", ""
	if len(args) == 1 {
		key = args[0]
	}
	if len(args) >= 2 {
		actor, key = args[0], args[1]
	}
	if strings.TrimSpace(key) == "" {
		return Appointment{}, ErrMissingIdempotencyKey
	}
	status = strings.TrimSpace(status)
	if !validAppointmentStatus(status) {
		return Appointment{}, fmt.Errorf("%w: unknown status", ErrInvalidInput)
	}
	resourceKey := "appointment:status:" + id + ":" + key
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return Appointment{}, err
	} else if ok {
		return s.store.GetAppointment(ctx, existing)
	}
	release, err := s.idem.Lock(ctx, "appointment:status-lock:"+id, 10*time.Second)
	if err != nil {
		return Appointment{}, err
	}
	defer release()
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return Appointment{}, err
	} else if ok {
		return s.store.GetAppointment(ctx, existing)
	}
	if actor == "" {
		actor = "运营人员"
	}
	a, _, err := s.store.UpdateAppointmentStatus(ctx, id, status, actor)
	if err != nil {
		return Appointment{}, err
	}
	if err := s.idem.Set(ctx, resourceKey, a.ID, 24*time.Hour); err != nil {
		return Appointment{}, err
	}
	return a, nil
}

func (s *CareService) CompleteFollowup(ctx context.Context, id, key string) (Followup, error) {
	if strings.TrimSpace(key) == "" {
		return Followup{}, ErrMissingIdempotencyKey
	}
	resourceKey := "followup:complete:" + id + ":" + key
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return Followup{}, err
	} else if ok {
		return findFollowup(ctx, s.store, existing)
	}
	release, err := s.idem.Lock(ctx, "followup:complete-lock:"+id, 10*time.Second)
	if err != nil {
		return Followup{}, err
	}
	defer release()
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return Followup{}, err
	} else if ok {
		return findFollowup(ctx, s.store, existing)
	}
	f, err := s.store.CompleteFollowup(ctx, id)
	if err != nil {
		return Followup{}, err
	}
	if err := s.idem.Set(ctx, resourceKey, f.ID, 24*time.Hour); err != nil {
		return Followup{}, err
	}
	return f, nil
}

func findFollowup(ctx context.Context, store CareStore, id string) (Followup, error) {
	list, _, err := store.ListFollowups(ctx, 1, 100, "")
	if err != nil {
		return Followup{}, err
	}
	for _, f := range list {
		if f.ID == id {
			return f, nil
		}
	}
	return Followup{}, ErrNotFound
}

func validAppointmentStatus(status string) bool {
	switch status {
	case AppointmentPending, AppointmentChecked, AppointmentWaiting, AppointmentServing, AppointmentCompleted, AppointmentCancelled:
		return true
	}
	return false
}

// CreateContentItem validates and creates a topic with an idempotent request key.
func (s *CareService) CreateContentItem(ctx context.Context, input CreateContentItemInput, key string) (ContentItem, error) {
	if strings.TrimSpace(key) == "" {
		return ContentItem{}, ErrMissingIdempotencyKey
	}
	if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Channel) == "" || strings.TrimSpace(input.Owner) == "" {
		return ContentItem{}, fmt.Errorf("%w: title, channel and owner are required", ErrInvalidInput)
	}
	resourceKey := "content:create:" + key
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return ContentItem{}, err
	} else if ok {
		return s.store.GetContentItem(ctx, existing)
	}
	release, err := s.idem.Lock(ctx, "content:create-lock", 10*time.Second)
	if err != nil {
		return ContentItem{}, err
	}
	defer release()
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return ContentItem{}, err
	} else if ok {
		return s.store.GetContentItem(ctx, existing)
	}
	item, err := s.store.CreateContentItem(ctx, ContentItem{Title: strings.TrimSpace(input.Title), Channel: strings.TrimSpace(input.Channel), Owner: strings.TrimSpace(input.Owner), PlannedAt: strings.TrimSpace(input.PlannedAt), Status: ContentStatusTopic})
	if err != nil {
		return ContentItem{}, err
	}
	if err := s.idem.Set(ctx, resourceKey, item.ID, 24*time.Hour); err != nil {
		return ContentItem{}, err
	}
	return item, nil
}

// SaveContentScript stores a draft and advances the topic into writing/production.
func (s *CareService) SaveContentScript(ctx context.Context, id string, input SaveContentScriptInput, key string) (ContentItem, error) {
	if strings.TrimSpace(key) == "" {
		return ContentItem{}, ErrMissingIdempotencyKey
	}
	if strings.TrimSpace(input.Body) == "" {
		return ContentItem{}, fmt.Errorf("%w: script body is required", ErrInvalidInput)
	}
	resourceKey := "content:script:" + id + ":" + key
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return ContentItem{}, err
	} else if ok {
		return s.store.GetContentItem(ctx, existing)
	}
	release, err := s.idem.Lock(ctx, "content:script-lock:"+id, 10*time.Second)
	if err != nil {
		return ContentItem{}, err
	}
	defer release()
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return ContentItem{}, err
	} else if ok {
		return s.store.GetContentItem(ctx, existing)
	}
	item, err := s.store.SaveContentScript(ctx, id, ContentScript{Body: strings.TrimSpace(input.Body)})
	if err != nil {
		return ContentItem{}, err
	}
	if err := s.idem.Set(ctx, resourceKey, item.ID, 24*time.Hour); err != nil {
		return ContentItem{}, err
	}
	return item, nil
}

// SubmitContentReview moves a production-ready item into the review queue.
func (s *CareService) SubmitContentReview(ctx context.Context, id, actor, key string) (ContentItem, error) {
	if strings.TrimSpace(key) == "" {
		return ContentItem{}, ErrMissingIdempotencyKey
	}
	if strings.TrimSpace(actor) == "" {
		return ContentItem{}, fmt.Errorf("%w: actor is required", ErrInvalidInput)
	}
	resourceKey := "content:review:" + id + ":" + key
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return ContentItem{}, err
	} else if ok {
		return s.store.GetContentItem(ctx, existing)
	}
	release, err := s.idem.Lock(ctx, "content:review-lock:"+id, 10*time.Second)
	if err != nil {
		return ContentItem{}, err
	}
	defer release()
	item, err := s.store.SubmitContentReview(ctx, id, strings.TrimSpace(actor))
	if err != nil {
		return ContentItem{}, err
	}
	if err := s.idem.Set(ctx, resourceKey, item.ID, 24*time.Hour); err != nil {
		return ContentItem{}, err
	}
	return item, nil
}

// PublishContent publishes a reviewed item and records the responsible actor.
func (s *CareService) PublishContent(ctx context.Context, id string, input PublishContentInput, key string) (ContentItem, error) {
	if strings.TrimSpace(key) == "" {
		return ContentItem{}, ErrMissingIdempotencyKey
	}
	if strings.TrimSpace(input.PublishedAt) == "" || strings.TrimSpace(input.Actor) == "" {
		return ContentItem{}, fmt.Errorf("%w: publishedAt and actor are required", ErrInvalidInput)
	}
	resourceKey := "content:publish:" + id + ":" + key
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return ContentItem{}, err
	} else if ok {
		return s.store.GetContentItem(ctx, existing)
	}
	release, err := s.idem.Lock(ctx, "content:publish-lock:"+id, 10*time.Second)
	if err != nil {
		return ContentItem{}, err
	}
	defer release()
	item, err := s.store.PublishContent(ctx, id, PublishRecord{PublishedAt: strings.TrimSpace(input.PublishedAt), Actor: strings.TrimSpace(input.Actor)})
	if err != nil {
		return ContentItem{}, err
	}
	if err := s.idem.Set(ctx, resourceKey, item.ID, 24*time.Hour); err != nil {
		return ContentItem{}, err
	}
	return item, nil
}

// RecordContentMetrics validates and records post-publication counters.
func (s *CareService) RecordContentMetrics(ctx context.Context, id string, input RecordContentMetricsInput, key string) (ContentItem, error) {
	if strings.TrimSpace(key) == "" {
		return ContentItem{}, ErrMissingIdempotencyKey
	}
	if input.Views < 0 || input.Likes < 0 || input.Comments < 0 || input.Shares < 0 {
		return ContentItem{}, fmt.Errorf("%w: metrics cannot be negative", ErrInvalidInput)
	}
	resourceKey := "content:metrics:" + id + ":" + key
	if existing, ok, err := s.idem.Get(ctx, resourceKey); err != nil {
		return ContentItem{}, err
	} else if ok {
		return s.store.GetContentItem(ctx, existing)
	}
	release, err := s.idem.Lock(ctx, "content:metrics-lock:"+id, 10*time.Second)
	if err != nil {
		return ContentItem{}, err
	}
	defer release()
	item, err := s.store.RecordContentMetrics(ctx, id, ContentMetrics{Views: input.Views, Likes: input.Likes, Comments: input.Comments, Shares: input.Shares})
	if err != nil {
		return ContentItem{}, err
	}
	if err := s.idem.Set(ctx, resourceKey, item.ID, 24*time.Hour); err != nil {
		return ContentItem{}, err
	}
	return item, nil
}

func httpStatusForError(err error) int {
	switch {
	case errors.Is(err, ErrMissingIdempotencyKey), errors.Is(err, ErrInvalidInput):
		return 400
	case errors.Is(err, ErrNotFound):
		return 404
	case errors.Is(err, ErrInvalidTransition), errors.Is(err, ErrIdempotencyBusy):
		return 409
	default:
		return 500
	}
}
