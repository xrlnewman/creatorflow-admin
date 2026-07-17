package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

func cloneContentItem(item ContentItem) ContentItem {
	out := item
	if item.Script != nil {
		copy := *item.Script
		out.Script = &copy
	}
	if item.Publish != nil {
		copy := *item.Publish
		out.Publish = &copy
	}
	if item.Metrics != nil {
		copy := *item.Metrics
		out.Metrics = &copy
	}
	out.Events = append([]ContentEvent(nil), item.Events...)
	return out
}

func seedContentItems(s *MemoryStore) {
	statuses := []string{ContentStatusTopic, ContentStatusWriting, ContentStatusProducing, ContentStatusReviewing, ContentStatusPublished, ContentStatusReviewed}
	channels := []string{"短视频", "图文专栏", "直播栏目", "品牌合作"}
	owners := []string{"林编辑", "沈编辑", "赵编辑", "周编辑"}
	for i := 1; i <= 24; i++ {
		id := fmt.Sprintf("CF-0718-%03d", i)
		status := statuses[(i-1)%len(statuses)]
		item := ContentItem{
			ID: id, Title: []string{"城市夜行：下班后的十五分钟", "一周好物：把桌面整理成工作流", "品牌访谈：小店如何留住老客", "夏日直播：创作者增长公开课"}[(i-1)%4],
			Channel: channels[(i-1)%len(channels)], Owner: owners[(i-1)%len(owners)], PlannedAt: fmt.Sprintf("2026-07-%02dT%02d:00:00+08:00", 18+(i%8), 9+(i%10)), Status: status,
			CreatedAt: nowUTC(), UpdatedAt: nowUTC(),
		}
		s.contentItems[id] = item
		for _, to := range []string{ContentStatusWriting, ContentStatusProducing, ContentStatusReviewing, ContentStatusPublished, ContentStatusReviewed} {
			if !contentReached(status, to) {
				break
			}
			action := "write_script"
			if to == ContentStatusProducing {
				action = "start_production"
			} else if to == ContentStatusReviewing {
				action = "submit_review"
			} else if to == ContentStatusPublished {
				action = "publish"
			} else if to == ContentStatusReviewed {
				action = "record_metrics"
			}
			from := item.Status
			if from == to {
				break
			}
			item.Status = to
			s.contentItems[id] = item
			s.contentEvents[id] = append(s.contentEvents[id], ContentEvent{ID: fmt.Sprintf("%s-EV-%d", id, len(s.contentEvents[id])+1), ContentItemID: id, FromStatus: from, ToStatus: to, Action: action, Actor: item.Owner, CreatedAt: nowUTC()})
		}
		item.Status = status
		if status == ContentStatusWriting || status == ContentStatusProducing || status == ContentStatusReviewing || status == ContentStatusPublished || status == ContentStatusReviewed {
			item.Script = &ContentScript{ID: id + "-SCRIPT", ContentItemID: id, Body: "开场钩子、三段主体和结尾行动号召。", UpdatedAt: nowUTC()}
		}
		if status == ContentStatusPublished || status == ContentStatusReviewed {
			item.Publish = &PublishRecord{ID: id + "-PUB", ContentItemID: id, PublishedAt: "2026-07-18T18:00:00+08:00", Actor: "主编", CreatedAt: nowUTC()}
		}
		if status == ContentStatusReviewed {
			item.Metrics = &ContentMetrics{ID: id + "-METRIC", ContentItemID: id, Views: 8000 + i*731, Likes: 420 + i*19, Comments: 18 + i, Shares: 31 + i*2, RecordedAt: nowUTC()}
		}
		item.UpdatedAt = nowUTC()
		s.contentItems[id] = item
	}
}

func contentReached(current, target string) bool {
	order := map[string]int{ContentStatusTopic: 0, ContentStatusWriting: 1, ContentStatusProducing: 2, ContentStatusReviewing: 3, ContentStatusPublished: 4, ContentStatusReviewed: 5}
	return order[current] >= order[target]
}

func (s *MemoryStore) ListContentItems(_ context.Context, page, pageSize int, status, owner, plannedAt, publishedAt string) ([]ContentItem, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := make([]ContentItem, 0, len(s.contentItems))
	for _, item := range s.contentItems {
		if status != "" && item.Status != status {
			continue
		}
		if owner != "" && item.Owner != owner {
			continue
		}
		if plannedAt != "" && !strings.HasPrefix(item.PlannedAt, plannedAt) {
			continue
		}
		if publishedAt != "" && (item.Publish == nil || !strings.HasPrefix(item.Publish.PublishedAt, publishedAt)) {
			continue
		}
		all = append(all, cloneContentItem(item))
	}
	sort.Slice(all, func(i, j int) bool { return all[i].PlannedAt < all[j].PlannedAt })
	return paginate(all, page, pageSize)
}

func (s *MemoryStore) GetContentItem(_ context.Context, id string) (ContentItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.contentItems[id]
	if !ok {
		return ContentItem{}, ErrNotFound
	}
	item = cloneContentItem(item)
	item.Events = append([]ContentEvent(nil), s.contentEvents[id]...)
	return item, nil
}

func (s *MemoryStore) ListContentEvents(_ context.Context, id string) ([]ContentEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.contentItems[id]; !ok {
		return nil, ErrNotFound
	}
	out := append([]ContentEvent(nil), s.contentEvents[id]...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt < out[j].CreatedAt })
	return out, nil
}

func (s *MemoryStore) CreateContentItem(_ context.Context, item ContentItem) (ContentItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if item.ID == "" {
		item.ID = s.next("CF")
	}
	if item.Status == "" {
		item.Status = ContentStatusTopic
	}
	if item.CreatedAt == "" {
		item.CreatedAt = nowUTC()
	}
	item.UpdatedAt = item.CreatedAt
	s.contentItems[item.ID] = cloneContentItem(item)
	s.contentEvents[item.ID] = append(s.contentEvents[item.ID], ContentEvent{ID: s.next("CFE"), ContentItemID: item.ID, FromStatus: "", ToStatus: item.Status, Action: "create", Actor: item.Owner, CreatedAt: item.CreatedAt})
	return cloneContentItem(item), nil
}

func (s *MemoryStore) SaveContentScript(_ context.Context, id string, script ContentScript) (ContentItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.contentItems[id]
	if !ok {
		return ContentItem{}, ErrNotFound
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
	script.ID = id + "-SCRIPT"
	script.ContentItemID = id
	script.UpdatedAt = nowUTC()
	item.Script = &script
	item.UpdatedAt = script.UpdatedAt
	s.contentItems[id] = item
	s.contentEvents[id] = append(s.contentEvents[id], ContentEvent{ID: s.next("CFE"), ContentItemID: id, FromStatus: from, ToStatus: item.Status, Action: action, Actor: item.Owner, CreatedAt: script.UpdatedAt})
	return cloneContentItem(item), nil
}

func (s *MemoryStore) SubmitContentReview(_ context.Context, id, actor string) (ContentItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.contentItems[id]
	if !ok {
		return ContentItem{}, ErrNotFound
	}
	if !contentTransitions[item.Status][ContentStatusReviewing] {
		return ContentItem{}, ErrInvalidTransition
	}
	old, now := item.Status, nowUTC()
	item.Status, item.UpdatedAt = ContentStatusReviewing, now
	s.contentItems[id] = item
	s.contentEvents[id] = append(s.contentEvents[id], ContentEvent{ID: s.next("CFE"), ContentItemID: id, FromStatus: old, ToStatus: item.Status, Action: "submit_review", Actor: actor, CreatedAt: now})
	return cloneContentItem(item), nil
}

func (s *MemoryStore) PublishContent(_ context.Context, id string, publish PublishRecord) (ContentItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.contentItems[id]
	if !ok {
		return ContentItem{}, ErrNotFound
	}
	if !contentTransitions[item.Status][ContentStatusPublished] {
		return ContentItem{}, ErrInvalidTransition
	}
	now := nowUTC()
	publish.ID, publish.ContentItemID, publish.CreatedAt = id+"-PUB", id, now
	item.Status, item.UpdatedAt, item.Publish = ContentStatusPublished, now, &publish
	s.contentItems[id] = item
	s.contentEvents[id] = append(s.contentEvents[id], ContentEvent{ID: s.next("CFE"), ContentItemID: id, FromStatus: ContentStatusReviewing, ToStatus: item.Status, Action: "publish", Actor: publish.Actor, CreatedAt: now})
	return cloneContentItem(item), nil
}

func (s *MemoryStore) RecordContentMetrics(_ context.Context, id string, metrics ContentMetrics) (ContentItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.contentItems[id]
	if !ok {
		return ContentItem{}, ErrNotFound
	}
	if item.Status != ContentStatusPublished && item.Status != ContentStatusReviewed {
		return ContentItem{}, ErrInvalidTransition
	}
	from, now := item.Status, nowUTC()
	if item.Status == ContentStatusPublished {
		item.Status = ContentStatusReviewed
	}
	metrics.ID, metrics.ContentItemID, metrics.RecordedAt = id+"-METRIC", id, now
	item.Metrics, item.UpdatedAt = &metrics, now
	s.contentItems[id] = item
	s.contentEvents[id] = append(s.contentEvents[id], ContentEvent{ID: s.next("CFE"), ContentItemID: id, FromStatus: from, ToStatus: item.Status, Action: "record_metrics", Actor: "运营人员", CreatedAt: now})
	return cloneContentItem(item), nil
}
