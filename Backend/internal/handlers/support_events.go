package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const supportEventBufferSize = 32

const (
	supportEventTypeMessage  = "message"
	supportEventTypePresence = "presence"
	supportEventTypeStatus   = "status"
)

type SupportEvent struct {
	EventType      string `json:"event_type"`
	ConversationID int64  `json:"conversation_id"`
	MessageID      int64  `json:"message_id"`
	UserID         int64  `json:"user_id"`
	SenderRole     string `json:"sender_role"`
	AdminOnline    bool   `json:"admin_online"`
}

type SupportEventBroker struct {
	mu         sync.Mutex
	adminFeeds map[chan SupportEvent]struct{}
	userFeeds  map[int64]map[chan SupportEvent]struct{}
}

func NewSupportEventBroker() *SupportEventBroker {
	return &SupportEventBroker{
		adminFeeds: make(map[chan SupportEvent]struct{}),
		userFeeds:  make(map[int64]map[chan SupportEvent]struct{}),
	}
}

func (b *SupportEventBroker) SubscribeAdmin() (<-chan SupportEvent, func()) {
	if b == nil {
		return nil, func() {}
	}

	feed := make(chan SupportEvent, supportEventBufferSize)
	b.mu.Lock()
	wasOffline := len(b.adminFeeds) == 0
	b.adminFeeds[feed] = struct{}{}
	b.mu.Unlock()
	if wasOffline {
		b.PublishAdminPresence(true)
	}

	return feed, func() {
		b.mu.Lock()
		delete(b.adminFeeds, feed)
		isOffline := len(b.adminFeeds) == 0
		b.mu.Unlock()
		if isOffline {
			b.PublishAdminPresence(false)
		}
	}
}

func (b *SupportEventBroker) SubscribeUser(userID int64) (<-chan SupportEvent, func()) {
	if b == nil {
		return nil, func() {}
	}

	feed := make(chan SupportEvent, supportEventBufferSize)
	b.mu.Lock()
	if b.userFeeds[userID] == nil {
		b.userFeeds[userID] = make(map[chan SupportEvent]struct{})
	}
	b.userFeeds[userID][feed] = struct{}{}
	adminOnline := len(b.adminFeeds) > 0
	b.mu.Unlock()
	sendSupportEvent(feed, SupportEvent{
		EventType:   supportEventTypePresence,
		UserID:      userID,
		AdminOnline: adminOnline,
	})

	return feed, func() {
		b.mu.Lock()
		if feeds := b.userFeeds[userID]; feeds != nil {
			delete(feeds, feed)
			if len(feeds) == 0 {
				delete(b.userFeeds, userID)
			}
		}
		b.mu.Unlock()
	}
}

func (b *SupportEventBroker) Publish(event SupportEvent) {
	if b == nil {
		return
	}
	if event.EventType == "" {
		event.EventType = supportEventTypeMessage
	}

	b.mu.Lock()
	adminFeeds := make([]chan SupportEvent, 0, len(b.adminFeeds))
	for feed := range b.adminFeeds {
		adminFeeds = append(adminFeeds, feed)
	}
	userFeeds := make([]chan SupportEvent, 0, len(b.userFeeds[event.UserID]))
	for feed := range b.userFeeds[event.UserID] {
		userFeeds = append(userFeeds, feed)
	}
	b.mu.Unlock()

	for _, feed := range append(adminFeeds, userFeeds...) {
		sendSupportEvent(feed, event)
	}
}

func (b *SupportEventBroker) PublishAdminPresence(adminOnline bool) {
	if b == nil {
		return
	}

	b.mu.Lock()
	userFeeds := make([]chan SupportEvent, 0)
	for _, feeds := range b.userFeeds {
		for feed := range feeds {
			userFeeds = append(userFeeds, feed)
		}
	}
	b.mu.Unlock()

	for _, feed := range userFeeds {
		sendSupportEvent(feed, SupportEvent{
			EventType:   supportEventTypePresence,
			AdminOnline: adminOnline,
		})
	}
}

func sendSupportEvent(feed chan SupportEvent, event SupportEvent) {
	select {
	case feed <- event:
	default:
	}
}

func writeSupportEventStream(w http.ResponseWriter, r *http.Request, events <-chan SupportEvent) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			payload, err := json.Marshal(event)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(w, "event: support_message\ndata: %s\n\n", payload)
			flusher.Flush()
		case <-heartbeat.C:
			_, _ = fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
