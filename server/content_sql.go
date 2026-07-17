package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const contentSelect = `SELECT id,title,channel,owner,planned_at,status,created_at,updated_at FROM content_items`

func (s *SQLStore) ListContentItems(ctx context.Context, page, pageSize int, status, owner, plannedAt, publishedAt string) ([]ContentItem, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	where := []string{"1=1"}
	args := []any{}
	if status != "" {
		where, args = append(where, "status=?"), append(args, status)
	}
	if owner != "" {
		where, args = append(where, "owner=?"), append(args, owner)
	}
	if plannedAt != "" {
		where, args = append(where, "planned_at LIKE CONCAT(?, '%')"), append(args, plannedAt)
	}
	if publishedAt != "" {
		where, args = append(where, "EXISTS (SELECT 1 FROM content_publish_records p WHERE p.content_item_id=content_items.id AND p.published_at LIKE CONCAT(?, '%'))"), append(args, publishedAt)
	}
	condition := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM content_items WHERE "+condition, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, pageSize, (page-1)*pageSize)
	rows, err := s.db.QueryContext(ctx, contentSelect+" WHERE "+condition+" ORDER BY planned_at ASC LIMIT ? OFFSET ?", queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []ContentItem{}
	for rows.Next() {
		item, err := scanContentItem(rows)
		if err != nil {
			return nil, 0, err
		}
		if err := s.loadContentDetails(ctx, &item, false); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *SQLStore) GetContentItem(ctx context.Context, id string) (ContentItem, error) {
	row := s.db.QueryRowContext(ctx, contentSelect+" WHERE id=?", id)
	item, err := scanContentItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return ContentItem{}, ErrNotFound
	}
	if err != nil {
		return ContentItem{}, err
	}
	if err := s.loadContentDetails(ctx, &item, true); err != nil {
		return ContentItem{}, err
	}
	return item, nil
}

func (s *SQLStore) ListContentEvents(ctx context.Context, id string) ([]ContentEvent, error) {
	var exists string
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM content_items WHERE id=?`, id).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,content_item_id,from_status,to_status,action,actor,created_at FROM content_events WHERE content_item_id=? ORDER BY created_at ASC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []ContentEvent{}
	for rows.Next() {
		var event ContentEvent
		if err := rows.Scan(&event.ID, &event.ContentItemID, &event.FromStatus, &event.ToStatus, &event.Action, &event.Actor, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *SQLStore) CreateContentItem(ctx context.Context, item ContentItem) (ContentItem, error) {
	if item.ID == "" {
		item.ID = fmt.Sprintf("CF-%d", nowUnixNano())
	}
	if item.Status == "" {
		item.Status = ContentStatusTopic
	}
	if item.CreatedAt == "" {
		item.CreatedAt = nowUTC()
	}
	item.UpdatedAt = item.CreatedAt
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ContentItem{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO content_items (id,title,channel,owner,planned_at,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?)`, item.ID, item.Title, item.Channel, item.Owner, item.PlannedAt, item.Status, item.CreatedAt, item.UpdatedAt); err != nil {
		return ContentItem{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO content_events (id,content_item_id,from_status,to_status,action,actor,created_at) VALUES (?,?,?,?,?,?,?)`, fmt.Sprintf("CFE-%d", nowUnixNano()), item.ID, "", item.Status, "create", item.Owner, item.CreatedAt); err != nil {
		return ContentItem{}, err
	}
	if err := tx.Commit(); err != nil {
		return ContentItem{}, err
	}
	return item, nil
}

func (s *SQLStore) SaveContentScript(ctx context.Context, id string, script ContentScript) (ContentItem, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ContentItem{}, err
	}
	defer tx.Rollback()
	item, err := scanContentItem(tx.QueryRowContext(ctx, contentSelect+" WHERE id=? FOR UPDATE", id))
	if errors.Is(err, sql.ErrNoRows) {
		return ContentItem{}, ErrNotFound
	}
	if err != nil {
		return ContentItem{}, err
	}
	from, action := item.Status, "save_script"
	switch item.Status {
	case ContentStatusTopic:
		item.Status, action = ContentStatusWriting, "write_script"
	case ContentStatusWriting:
		item.Status, action = ContentStatusProducing, "start_production"
	case ContentStatusProducing:
	default:
		return ContentItem{}, ErrInvalidTransition
	}
	now := nowUTC()
	if _, err := tx.ExecContext(ctx, `INSERT INTO content_scripts (id,content_item_id,body,updated_at) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE body=VALUES(body),updated_at=VALUES(updated_at)`, id+"-SCRIPT", id, script.Body, now); err != nil {
		return ContentItem{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE content_items SET status=?,updated_at=? WHERE id=?`, item.Status, now, id); err != nil {
		return ContentItem{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO content_events (id,content_item_id,from_status,to_status,action,actor,created_at) VALUES (?,?,?,?,?,?,?)`, fmt.Sprintf("CFE-%d", nowUnixNano()), id, from, item.Status, action, item.Owner, now); err != nil {
		return ContentItem{}, err
	}
	if err := tx.Commit(); err != nil {
		return ContentItem{}, err
	}
	item.Script = &ContentScript{ID: id + "-SCRIPT", ContentItemID: id, Body: script.Body, UpdatedAt: now}
	item.UpdatedAt = now
	return item, nil
}

func (s *SQLStore) SubmitContentReview(ctx context.Context, id, actor string) (ContentItem, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ContentItem{}, err
	}
	defer tx.Rollback()
	item, err := scanContentItem(tx.QueryRowContext(ctx, contentSelect+" WHERE id=? FOR UPDATE", id))
	if errors.Is(err, sql.ErrNoRows) {
		return ContentItem{}, ErrNotFound
	}
	if err != nil {
		return ContentItem{}, err
	}
	if !contentTransitions[item.Status][ContentStatusReviewing] {
		return ContentItem{}, ErrInvalidTransition
	}
	now := nowUTC()
	if _, err := tx.ExecContext(ctx, `UPDATE content_items SET status=?,updated_at=? WHERE id=?`, ContentStatusReviewing, now, id); err != nil {
		return ContentItem{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO content_events (id,content_item_id,from_status,to_status,action,actor,created_at) VALUES (?,?,?,?,?,?,?)`, fmt.Sprintf("CFE-%d", nowUnixNano()), id, item.Status, ContentStatusReviewing, "submit_review", actor, now); err != nil {
		return ContentItem{}, err
	}
	if err := tx.Commit(); err != nil {
		return ContentItem{}, err
	}
	item.Status, item.UpdatedAt = ContentStatusReviewing, now
	return item, nil
}

func (s *SQLStore) PublishContent(ctx context.Context, id string, publish PublishRecord) (ContentItem, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ContentItem{}, err
	}
	defer tx.Rollback()
	item, err := scanContentItem(tx.QueryRowContext(ctx, contentSelect+" WHERE id=? FOR UPDATE", id))
	if errors.Is(err, sql.ErrNoRows) {
		return ContentItem{}, ErrNotFound
	}
	if err != nil {
		return ContentItem{}, err
	}
	if !contentTransitions[item.Status][ContentStatusPublished] {
		return ContentItem{}, ErrInvalidTransition
	}
	now := nowUTC()
	if _, err := tx.ExecContext(ctx, `UPDATE content_items SET status=?,updated_at=? WHERE id=?`, ContentStatusPublished, now, id); err != nil {
		return ContentItem{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO content_publish_records (id,content_item_id,published_at,actor,created_at) VALUES (?,?,?,?,?)`, id+"-PUB", id, publish.PublishedAt, publish.Actor, now); err != nil {
		return ContentItem{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO content_events (id,content_item_id,from_status,to_status,action,actor,created_at) VALUES (?,?,?,?,?,?,?)`, fmt.Sprintf("CFE-%d", nowUnixNano()), id, item.Status, ContentStatusPublished, "publish", publish.Actor, now); err != nil {
		return ContentItem{}, err
	}
	if err := tx.Commit(); err != nil {
		return ContentItem{}, err
	}
	item.Status, item.UpdatedAt = ContentStatusPublished, now
	publish.ID, publish.ContentItemID, publish.CreatedAt = id+"-PUB", id, now
	item.Publish = &publish
	return item, nil
}

func (s *SQLStore) RecordContentMetrics(ctx context.Context, id string, metrics ContentMetrics) (ContentItem, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ContentItem{}, err
	}
	defer tx.Rollback()
	item, err := scanContentItem(tx.QueryRowContext(ctx, contentSelect+" WHERE id=? FOR UPDATE", id))
	if errors.Is(err, sql.ErrNoRows) {
		return ContentItem{}, ErrNotFound
	}
	if err != nil {
		return ContentItem{}, err
	}
	if item.Status != ContentStatusPublished && item.Status != ContentStatusReviewed {
		return ContentItem{}, ErrInvalidTransition
	}
	from, status, now := item.Status, item.Status, nowUTC()
	if status == ContentStatusPublished {
		status = ContentStatusReviewed
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO content_metrics (id,content_item_id,views,likes,comments,shares,recorded_at) VALUES (?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE views=VALUES(views),likes=VALUES(likes),comments=VALUES(comments),shares=VALUES(shares),recorded_at=VALUES(recorded_at)`, id+"-METRIC", id, metrics.Views, metrics.Likes, metrics.Comments, metrics.Shares, now); err != nil {
		return ContentItem{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE content_items SET status=?,updated_at=? WHERE id=?`, status, now, id); err != nil {
		return ContentItem{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO content_events (id,content_item_id,from_status,to_status,action,actor,created_at) VALUES (?,?,?,?,?,?,?)`, fmt.Sprintf("CFE-%d", nowUnixNano()), id, from, status, "record_metrics", "运营人员", now); err != nil {
		return ContentItem{}, err
	}
	if err := tx.Commit(); err != nil {
		return ContentItem{}, err
	}
	item.Status, item.UpdatedAt = status, now
	metrics.ID, metrics.ContentItemID, metrics.RecordedAt = id+"-METRIC", id, now
	item.Metrics = &metrics
	return item, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanContentItem(row rowScanner) (ContentItem, error) {
	var item ContentItem
	err := row.Scan(&item.ID, &item.Title, &item.Channel, &item.Owner, &item.PlannedAt, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (s *SQLStore) loadContentDetails(ctx context.Context, item *ContentItem, withEvents bool) error {
	var script ContentScript
	if err := s.db.QueryRowContext(ctx, `SELECT id,content_item_id,body,updated_at FROM content_scripts WHERE content_item_id=?`, item.ID).Scan(&script.ID, &script.ContentItemID, &script.Body, &script.UpdatedAt); err == nil {
		item.Script = &script
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var publish PublishRecord
	if err := s.db.QueryRowContext(ctx, `SELECT id,content_item_id,published_at,actor,created_at FROM content_publish_records WHERE content_item_id=?`, item.ID).Scan(&publish.ID, &publish.ContentItemID, &publish.PublishedAt, &publish.Actor, &publish.CreatedAt); err == nil {
		item.Publish = &publish
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var metrics ContentMetrics
	if err := s.db.QueryRowContext(ctx, `SELECT id,content_item_id,views,likes,comments,shares,recorded_at FROM content_metrics WHERE content_item_id=?`, item.ID).Scan(&metrics.ID, &metrics.ContentItemID, &metrics.Views, &metrics.Likes, &metrics.Comments, &metrics.Shares, &metrics.RecordedAt); err == nil {
		item.Metrics = &metrics
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if withEvents {
		events, err := s.ListContentEvents(ctx, item.ID)
		if err != nil {
			return err
		}
		item.Events = events
	}
	return nil
}

func nowUnixNano() int64 {
	return time.Now().UnixNano()
}
