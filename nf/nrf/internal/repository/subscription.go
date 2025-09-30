package repository

import (
	"time"
)

// Subscription represents an NF status notification subscription (TS 29.510)
type Subscription struct {
	SubscriptionID string    `json:"subscriptionId"`
	NFInstanceID   string    `json:"nfInstanceId,omitempty"` // Subscribe to specific NF
	NFType         NFType    `json:"nfType,omitempty"`       // Subscribe to NF type
	CallbackURI    string    `json:"nfStatusNotificationUri"`
	ValidityTime   time.Time `json:"validityTime,omitempty"`

	// Notification conditions
	EventType []string `json:"reqNotifEvents,omitempty"` // e.g., ["NF_REGISTERED", "NF_DEREGISTERED"]

	// Metadata
	CreatedAt time.Time `json:"createdAt"`
}

// IsExpired checks if the subscription has expired
func (s *Subscription) IsExpired() bool {
	if s.ValidityTime.IsZero() {
		return false // No expiry
	}
	return time.Now().After(s.ValidityTime)
}

// MatchesEvent checks if the subscription is interested in an event
func (s *Subscription) MatchesEvent(eventType string) bool {
	if len(s.EventType) == 0 {
		return true // Subscribe to all events
	}

	for _, e := range s.EventType {
		if e == eventType {
			return true
		}
	}
	return false
}

// MatchesProfile checks if the subscription matches an NF profile
func (s *Subscription) MatchesProfile(profile *NFProfile) bool {
	// If specific NF instance is subscribed
	if s.NFInstanceID != "" {
		return s.NFInstanceID == profile.NFInstanceID
	}

	// If NF type is specified
	if s.NFType != "" {
		return s.NFType == profile.NFType
	}

	// Subscribe to all NFs
	return true
}
