package main

import "time"

// Appointment is a clinic appointment in the operational workflow.
type Appointment struct {
	ID          string `json:"id"`
	PatientID   string `json:"patientId,omitempty"`
	Patient     string `json:"patient"`
	Department  string `json:"department"`
	Doctor      string `json:"doctor"`
	ScheduledAt string `json:"scheduledAt"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// AppointmentEvent records every state transition for audit and queue replay.
type AppointmentEvent struct {
	ID            string `json:"id"`
	AppointmentID string `json:"appointmentId"`
	FromStatus    string `json:"fromStatus"`
	ToStatus      string `json:"toStatus"`
	Actor         string `json:"actor"`
	CreatedAt     string `json:"createdAt"`
}

// Department is a clinic service line.
type Department struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Doctor is an operational provider profile, not a medical record.
type Doctor struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Status     string `json:"status"`
	TodayCount int    `json:"todayCount"`
}

// Patient contains synthetic identifiers used by the demo workflow.
type Patient struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	LastVisit string `json:"lastVisit"`
}

// Followup is a non-diagnostic operational callback task.
type Followup struct {
	ID        string `json:"id"`
	PatientID string `json:"patientId,omitempty"`
	Patient   string `json:"patient"`
	Summary   string `json:"summary"`
	DueAt     string `json:"dueAt"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// CreateAppointmentInput is accepted by POST /appointments.
type CreateAppointmentInput struct {
	PatientID   string `json:"patientId"`
	Patient     string `json:"patient"`
	Department  string `json:"department"`
	Doctor      string `json:"doctor"`
	ScheduledAt string `json:"scheduledAt"`
}

// UpdateAppointmentStatusInput is accepted by POST /appointments/:id/status.
type UpdateAppointmentStatusInput struct {
	Status string `json:"status" binding:"required"`
	Actor  string `json:"actor"`
}

// CreateFollowupInput is accepted by POST /followups.
type CreateFollowupInput struct {
	PatientID string `json:"patientId"`
	Patient   string `json:"patient"`
	Summary   string `json:"summary"`
	DueAt     string `json:"dueAt"`
}

// Dashboard contains operational KPIs used by admin and mobile clients.
type Dashboard struct {
	TodayAppointments  int `json:"todayAppointments"`
	AverageWaitMinutes int `json:"averageWaitMinutes"`
	Completed          int `json:"completed"`
	CheckedIn          int `json:"checkedIn"`
	PendingFollowups   int `json:"pendingFollowups"`
}

// ContentItem is a synthetic content topic that moves through the editorial pipeline.
type ContentItem struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Channel   string          `json:"channel"`
	Owner     string          `json:"owner"`
	PlannedAt string          `json:"plannedAt"`
	Status    string          `json:"status"`
	CreatedAt string          `json:"createdAt"`
	UpdatedAt string          `json:"updatedAt"`
	Script    *ContentScript  `json:"script,omitempty"`
	Publish   *PublishRecord  `json:"publish,omitempty"`
	Metrics   *ContentMetrics `json:"metrics,omitempty"`
	Events    []ContentEvent  `json:"events,omitempty"`
}

// ContentScript is the latest script draft for a content item.
type ContentScript struct {
	ID            string `json:"id"`
	ContentItemID string `json:"contentItemId"`
	Body          string `json:"body"`
	UpdatedAt     string `json:"updatedAt"`
}

// PublishRecord captures the actor and timestamp used to publish a content item.
type PublishRecord struct {
	ID            string `json:"id"`
	ContentItemID string `json:"contentItemId"`
	PublishedAt   string `json:"publishedAt"`
	Actor         string `json:"actor"`
	CreatedAt     string `json:"createdAt"`
}

// ContentMetrics stores post-publication performance counters.
type ContentMetrics struct {
	ID            string `json:"id"`
	ContentItemID string `json:"contentItemId"`
	Views         int    `json:"views"`
	Likes         int    `json:"likes"`
	Comments      int    `json:"comments"`
	Shares        int    `json:"shares"`
	RecordedAt    string `json:"recordedAt"`
}

// ContentEvent is the immutable audit trail for content state changes and actions.
type ContentEvent struct {
	ID            string `json:"id"`
	ContentItemID string `json:"contentItemId"`
	FromStatus    string `json:"fromStatus"`
	ToStatus      string `json:"toStatus"`
	Action        string `json:"action"`
	Actor         string `json:"actor"`
	CreatedAt     string `json:"createdAt"`
}

// CreateContentItemInput is accepted by POST /content-items.
type CreateContentItemInput struct {
	Title     string `json:"title"`
	Channel   string `json:"channel"`
	Owner     string `json:"owner"`
	PlannedAt string `json:"plannedAt"`
}

// SaveContentScriptInput is accepted by POST /content-items/:id/script.
type SaveContentScriptInput struct {
	Body string `json:"body"`
}

// SubmitContentReviewInput is accepted by POST /content-items/:id/submit-review.
type SubmitContentReviewInput struct {
	Actor string `json:"actor"`
}

// PublishContentInput is accepted by POST /content-items/:id/publish.
type PublishContentInput struct {
	PublishedAt string `json:"publishedAt"`
	Actor       string `json:"actor"`
}

// RecordContentMetricsInput is accepted by POST /content-items/:id/metrics.
type RecordContentMetricsInput struct {
	Views    int `json:"views"`
	Likes    int `json:"likes"`
	Comments int `json:"comments"`
	Shares   int `json:"shares"`
}

const (
	AppointmentPending     = "待排期"
	AppointmentChecked     = "已排期"
	AppointmentWaiting     = "待制作"
	AppointmentServing     = "制作中"
	AppointmentCompleted   = "已发布"
	AppointmentCancelled   = "已取消"
	FollowupPending        = "待完成"
	FollowupCompleted      = "已完成"
	ContentStatusTopic     = "待选题"
	ContentStatusWriting   = "写作中"
	ContentStatusProducing = "制作中"
	ContentStatusReviewing = "待审核"
	ContentStatusPublished = "已发布"
	ContentStatusReviewed  = "已复盘"
)

var contentTransitions = map[string]map[string]bool{
	ContentStatusTopic:     {ContentStatusWriting: true},
	ContentStatusWriting:   {ContentStatusProducing: true},
	ContentStatusProducing: {ContentStatusReviewing: true},
	ContentStatusReviewing: {ContentStatusPublished: true},
	ContentStatusPublished: {ContentStatusReviewed: true},
	ContentStatusReviewed:  {},
}

var appointmentTransitions = map[string]map[string]bool{
	AppointmentPending:   {AppointmentChecked: true, AppointmentCancelled: true},
	AppointmentChecked:   {AppointmentWaiting: true, AppointmentCancelled: true},
	AppointmentWaiting:   {AppointmentServing: true, AppointmentCancelled: true},
	AppointmentServing:   {AppointmentCompleted: true},
	AppointmentCompleted: {},
	AppointmentCancelled: {},
}

func nowUTC() string { return time.Now().UTC().Format(time.RFC3339Nano) }
