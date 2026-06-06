package handlers

import (
	"testing"
	"time"
)

func TestSupportEventBrokerPublishesAdminPresence(t *testing.T) {
	broker := NewSupportEventBroker()

	userEvents, unsubscribeUser := broker.SubscribeUser(42)
	defer unsubscribeUser()

	initialEvent := receiveSupportEvent(t, userEvents)
	if initialEvent.EventType != supportEventTypePresence || initialEvent.AdminOnline {
		t.Fatalf("expected initial offline presence, got %+v", initialEvent)
	}

	_, unsubscribeAdmin := broker.SubscribeAdmin()
	onlineEvent := receiveSupportEvent(t, userEvents)
	if onlineEvent.EventType != supportEventTypePresence || !onlineEvent.AdminOnline {
		t.Fatalf("expected online presence, got %+v", onlineEvent)
	}

	unsubscribeAdmin()
	offlineEvent := receiveSupportEvent(t, userEvents)
	if offlineEvent.EventType != supportEventTypePresence || offlineEvent.AdminOnline {
		t.Fatalf("expected offline presence, got %+v", offlineEvent)
	}
}

func receiveSupportEvent(t *testing.T, events <-chan SupportEvent) SupportEvent {
	t.Helper()

	select {
	case event := <-events:
		return event
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for support event")
	}

	return SupportEvent{}
}
