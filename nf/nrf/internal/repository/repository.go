package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Repository manages NF profiles
type Repository interface {
	// NF Profile Management
	Register(ctx context.Context, profile *NFProfile) error
	Update(ctx context.Context, nfInstanceID string, profile *NFProfile) error
	Deregister(ctx context.Context, nfInstanceID string) error
	Get(ctx context.Context, nfInstanceID string) (*NFProfile, error)
	GetAll(ctx context.Context) ([]*NFProfile, error)

	// Discovery
	Discover(ctx context.Context, query *DiscoveryQuery) ([]*NFProfile, error)

	// Subscription Management
	Subscribe(ctx context.Context, subscription *Subscription) error
	Unsubscribe(ctx context.Context, subscriptionID string) error
	GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error)
	GetSubscriptionsByNFInstanceID(ctx context.Context, nfInstanceID string) ([]*Subscription, error)

	// Heartbeat
	UpdateHeartbeat(ctx context.Context, nfInstanceID string) error

	// Health
	GetStats(ctx context.Context) (*Stats, error)
}

// MemoryRepository is an in-memory implementation of Repository
type MemoryRepository struct {
	mu            sync.RWMutex
	profiles      map[string]*NFProfile    // nfInstanceID -> NFProfile
	subscriptions map[string]*Subscription // subscriptionID -> Subscription
	logger        *zap.Logger

	// Cleanup goroutine
	stopChan      chan struct{}
	cleanupTicker *time.Ticker
}

// NewMemoryRepository creates a new in-memory repository
func NewMemoryRepository(logger *zap.Logger) *MemoryRepository {
	repo := &MemoryRepository{
		profiles:      make(map[string]*NFProfile),
		subscriptions: make(map[string]*Subscription),
		logger:        logger,
		stopChan:      make(chan struct{}),
		cleanupTicker: time.NewTicker(30 * time.Second),
	}

	// Start cleanup goroutine
	go repo.cleanup()

	return repo
}

// Register registers a new NF profile
func (r *MemoryRepository) Register(ctx context.Context, profile *NFProfile) error {
	if !profile.IsValid() {
		return fmt.Errorf("invalid NF profile")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already exists
	if _, exists := r.profiles[profile.NFInstanceID]; exists {
		return fmt.Errorf("NF instance already registered: %s", profile.NFInstanceID)
	}

	// Set timestamps
	now := time.Now()
	profile.CreatedAt = now
	profile.UpdatedAt = now
	profile.LastHeartbeat = now
	profile.NFStatus = NFStatusRegistered

	r.profiles[profile.NFInstanceID] = profile

	r.logger.Info("NF registered",
		zap.String("nf_instance_id", profile.NFInstanceID),
		zap.String("nf_type", string(profile.NFType)),
	)

	// Notify subscribers
	go r.notifySubscribers(profile, "NF_REGISTERED")

	return nil
}

// Update updates an existing NF profile
func (r *MemoryRepository) Update(ctx context.Context, nfInstanceID string, profile *NFProfile) error {
	if !profile.IsValid() {
		return fmt.Errorf("invalid NF profile")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.profiles[nfInstanceID]
	if !exists {
		return fmt.Errorf("NF instance not found: %s", nfInstanceID)
	}

	// Preserve timestamps
	profile.CreatedAt = existing.CreatedAt
	profile.UpdatedAt = time.Now()
	profile.LastHeartbeat = existing.LastHeartbeat

	r.profiles[nfInstanceID] = profile

	r.logger.Info("NF profile updated",
		zap.String("nf_instance_id", nfInstanceID),
	)

	// Notify subscribers
	go r.notifySubscribers(profile, "NF_PROFILE_CHANGED")

	return nil
}

// Deregister removes an NF profile
func (r *MemoryRepository) Deregister(ctx context.Context, nfInstanceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	profile, exists := r.profiles[nfInstanceID]
	if !exists {
		return fmt.Errorf("NF instance not found: %s", nfInstanceID)
	}

	delete(r.profiles, nfInstanceID)

	r.logger.Info("NF deregistered",
		zap.String("nf_instance_id", nfInstanceID),
		zap.String("nf_type", string(profile.NFType)),
	)

	// Notify subscribers
	go r.notifySubscribers(profile, "NF_DEREGISTERED")

	return nil
}

// Get retrieves an NF profile by instance ID
func (r *MemoryRepository) Get(ctx context.Context, nfInstanceID string) (*NFProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	profile, exists := r.profiles[nfInstanceID]
	if !exists {
		return nil, fmt.Errorf("NF instance not found: %s", nfInstanceID)
	}

	// Return a copy
	profileCopy := *profile
	return &profileCopy, nil
}

// GetAll retrieves all NF profiles
func (r *MemoryRepository) GetAll(ctx context.Context) ([]*NFProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	profiles := make([]*NFProfile, 0, len(r.profiles))
	for _, profile := range r.profiles {
		profileCopy := *profile
		profiles = append(profiles, &profileCopy)
	}

	return profiles, nil
}

// Discover searches for NF profiles based on query criteria
func (r *MemoryRepository) Discover(ctx context.Context, query *DiscoveryQuery) ([]*NFProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*NFProfile

	for _, profile := range r.profiles {
		if query.Matches(profile) {
			profileCopy := *profile
			results = append(results, &profileCopy)
		}
	}

	r.logger.Debug("NF discovery",
		zap.Int("total_profiles", len(r.profiles)),
		zap.Int("matched", len(results)),
	)

	return results, nil
}

// UpdateHeartbeat updates the last heartbeat time for an NF
func (r *MemoryRepository) UpdateHeartbeat(ctx context.Context, nfInstanceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	profile, exists := r.profiles[nfInstanceID]
	if !exists {
		return fmt.Errorf("NF instance not found: %s", nfInstanceID)
	}

	profile.UpdateHeartbeat()

	return nil
}

// Subscribe creates a new subscription
func (r *MemoryRepository) Subscribe(ctx context.Context, subscription *Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if subscription.SubscriptionID == "" {
		return fmt.Errorf("subscription ID is required")
	}

	subscription.CreatedAt = time.Now()
	r.subscriptions[subscription.SubscriptionID] = subscription

	r.logger.Info("Subscription created",
		zap.String("subscription_id", subscription.SubscriptionID),
		zap.String("callback_uri", subscription.CallbackURI),
	)

	return nil
}

// Unsubscribe removes a subscription
func (r *MemoryRepository) Unsubscribe(ctx context.Context, subscriptionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.subscriptions[subscriptionID]; !exists {
		return fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	delete(r.subscriptions, subscriptionID)

	r.logger.Info("Subscription removed",
		zap.String("subscription_id", subscriptionID),
	)

	return nil
}

// GetSubscription retrieves a subscription
func (r *MemoryRepository) GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	subscription, exists := r.subscriptions[subscriptionID]
	if !exists {
		return nil, fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	subCopy := *subscription
	return &subCopy, nil
}

// GetSubscriptionsByNFInstanceID retrieves subscriptions for a specific NF instance
func (r *MemoryRepository) GetSubscriptionsByNFInstanceID(ctx context.Context, nfInstanceID string) ([]*Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*Subscription
	for _, sub := range r.subscriptions {
		if sub.NFInstanceID == nfInstanceID {
			subCopy := *sub
			results = append(results, &subCopy)
		}
	}

	return results, nil
}

// GetStats returns repository statistics
func (r *MemoryRepository) GetStats(ctx context.Context) (*Stats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := &Stats{
		TotalNFs:           len(r.profiles),
		TotalSubscriptions: len(r.subscriptions),
		NFsByType:          make(map[string]int),
		NFsByStatus:        make(map[string]int),
	}

	for _, profile := range r.profiles {
		stats.NFsByType[string(profile.NFType)]++
		stats.NFsByStatus[string(profile.NFStatus)]++
	}

	return stats, nil
}

// cleanup periodically removes expired NF profiles
func (r *MemoryRepository) cleanup() {
	for {
		select {
		case <-r.cleanupTicker.C:
			r.performCleanup()
		case <-r.stopChan:
			return
		}
	}
}

// performCleanup removes expired profiles
func (r *MemoryRepository) performCleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	var expired []string
	for id, profile := range r.profiles {
		if profile.IsExpired() {
			expired = append(expired, id)
		}
	}

	for _, id := range expired {
		profile := r.profiles[id]
		delete(r.profiles, id)

		r.logger.Warn("NF profile expired and removed",
			zap.String("nf_instance_id", id),
			zap.String("nf_type", string(profile.NFType)),
		)

		// Notify subscribers
		go r.notifySubscribers(profile, "NF_DEREGISTERED")
	}

	if len(expired) > 0 {
		r.logger.Info("Cleanup completed",
			zap.Int("expired_count", len(expired)),
		)
	}
}

// notifySubscribers notifies all relevant subscribers about an event
func (r *MemoryRepository) notifySubscribers(profile *NFProfile, eventType string) {
	// TODO: Implement notification sending to subscribers
	// This would involve making HTTP POST requests to callback URIs
	r.logger.Debug("Subscriber notification",
		zap.String("event_type", eventType),
		zap.String("nf_instance_id", profile.NFInstanceID),
	)
}

// Close stops the repository
func (r *MemoryRepository) Close() {
	close(r.stopChan)
	r.cleanupTicker.Stop()
}

// Stats represents repository statistics
type Stats struct {
	TotalNFs           int            `json:"total_nfs"`
	TotalSubscriptions int            `json:"total_subscriptions"`
	NFsByType          map[string]int `json:"nfs_by_type"`
	NFsByStatus        map[string]int `json:"nfs_by_status"`
}
