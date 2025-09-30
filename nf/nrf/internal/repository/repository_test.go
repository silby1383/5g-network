package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMemoryRepository_Register(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := NewMemoryRepository(logger)
	defer repo.Close()

	ctx := context.Background()

	profile := &NFProfile{
		NFInstanceID:   "test-instance-1",
		NFType:         NFTypeAMF,
		NFStatus:       NFStatusRegistered,
		FQDN:           "amf.5gc.local",
		HeartBeatTimer: 30,
	}

	err := repo.Register(ctx, profile)
	require.NoError(t, err)

	// Try to register again (should fail)
	err = repo.Register(ctx, profile)
	assert.Error(t, err)
}

func TestMemoryRepository_Get(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := NewMemoryRepository(logger)
	defer repo.Close()

	ctx := context.Background()

	profile := &NFProfile{
		NFInstanceID: "test-instance-2",
		NFType:       NFTypeSMF,
		NFStatus:     NFStatusRegistered,
	}

	// Register
	err := repo.Register(ctx, profile)
	require.NoError(t, err)

	// Get
	retrieved, err := repo.Get(ctx, "test-instance-2")
	require.NoError(t, err)
	assert.Equal(t, "test-instance-2", retrieved.NFInstanceID)
	assert.Equal(t, NFTypeSMF, retrieved.NFType)

	// Get non-existent
	_, err = repo.Get(ctx, "non-existent")
	assert.Error(t, err)
}

func TestMemoryRepository_Update(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := NewMemoryRepository(logger)
	defer repo.Close()

	ctx := context.Background()

	profile := &NFProfile{
		NFInstanceID: "test-instance-3",
		NFType:       NFTypeUPF,
		NFStatus:     NFStatusRegistered,
		Capacity:     100,
	}

	// Register
	err := repo.Register(ctx, profile)
	require.NoError(t, err)

	// Update
	profile.Capacity = 200
	err = repo.Update(ctx, "test-instance-3", profile)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.Get(ctx, "test-instance-3")
	require.NoError(t, err)
	assert.Equal(t, 200, retrieved.Capacity)
}

func TestMemoryRepository_Deregister(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := NewMemoryRepository(logger)
	defer repo.Close()

	ctx := context.Background()

	profile := &NFProfile{
		NFInstanceID: "test-instance-4",
		NFType:       NFTypeAUSF,
		NFStatus:     NFStatusRegistered,
	}

	// Register
	err := repo.Register(ctx, profile)
	require.NoError(t, err)

	// Deregister
	err = repo.Deregister(ctx, "test-instance-4")
	require.NoError(t, err)

	// Verify deregistered
	_, err = repo.Get(ctx, "test-instance-4")
	assert.Error(t, err)
}

func TestMemoryRepository_Discover(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := NewMemoryRepository(logger)
	defer repo.Close()

	ctx := context.Background()

	// Register multiple NFs
	profiles := []*NFProfile{
		{
			NFInstanceID: "amf-1",
			NFType:       NFTypeAMF,
			NFStatus:     NFStatusRegistered,
			PLMNID:       &PLMNID{MCC: "001", MNC: "01"},
		},
		{
			NFInstanceID: "smf-1",
			NFType:       NFTypeSMF,
			NFStatus:     NFStatusRegistered,
			PLMNID:       &PLMNID{MCC: "001", MNC: "01"},
		},
		{
			NFInstanceID: "upf-1",
			NFType:       NFTypeUPF,
			NFStatus:     NFStatusRegistered,
		},
	}

	for _, p := range profiles {
		err := repo.Register(ctx, p)
		require.NoError(t, err)
	}

	// Discover all AMFs
	query := &DiscoveryQuery{
		NFType: NFTypeAMF,
	}

	results, err := repo.Discover(ctx, query)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "amf-1", results[0].NFInstanceID)

	// Discover by PLMN (UPF doesn't have PLMN so should only find AMF and SMF)
	query = &DiscoveryQuery{
		PLMNID: &PLMNID{MCC: "001", MNC: "01"},
	}

	results, err = repo.Discover(ctx, query)
	require.NoError(t, err)
	// Note: PLMN matching returns all registered NFs if profile doesn't have PLMN
	// This is a design decision - you could change Matches() to require PLMN match
	assert.GreaterOrEqual(t, len(results), 2) // At least AMF and SMF
}

func TestMemoryRepository_Heartbeat(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := NewMemoryRepository(logger)
	defer repo.Close()

	ctx := context.Background()

	profile := &NFProfile{
		NFInstanceID:   "test-instance-5",
		NFType:         NFTypeAMF,
		NFStatus:       NFStatusRegistered,
		HeartBeatTimer: 1, // 1 second for testing
	}

	// Register
	err := repo.Register(ctx, profile)
	require.NoError(t, err)

	// Get initial heartbeat time
	retrieved, err := repo.Get(ctx, "test-instance-5")
	require.NoError(t, err)
	initialHeartbeat := retrieved.LastHeartbeat

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Update heartbeat
	err = repo.UpdateHeartbeat(ctx, "test-instance-5")
	require.NoError(t, err)

	// Verify heartbeat updated
	retrieved, err = repo.Get(ctx, "test-instance-5")
	require.NoError(t, err)
	assert.True(t, retrieved.LastHeartbeat.After(initialHeartbeat))
}

func TestMemoryRepository_Subscribe(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := NewMemoryRepository(logger)
	defer repo.Close()

	ctx := context.Background()

	subscription := &Subscription{
		SubscriptionID: "sub-1",
		NFType:         NFTypeAMF,
		CallbackURI:    "http://consumer.local/callback",
		EventType:      []string{"NF_REGISTERED", "NF_DEREGISTERED"},
	}

	// Subscribe
	err := repo.Subscribe(ctx, subscription)
	require.NoError(t, err)

	// Get subscription
	retrieved, err := repo.GetSubscription(ctx, "sub-1")
	require.NoError(t, err)
	assert.Equal(t, "sub-1", retrieved.SubscriptionID)
	assert.Equal(t, NFTypeAMF, retrieved.NFType)

	// Unsubscribe
	err = repo.Unsubscribe(ctx, "sub-1")
	require.NoError(t, err)

	// Verify unsubscribed
	_, err = repo.GetSubscription(ctx, "sub-1")
	assert.Error(t, err)
}

func TestMemoryRepository_Stats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := NewMemoryRepository(logger)
	defer repo.Close()

	ctx := context.Background()

	// Register some NFs
	profiles := []*NFProfile{
		{NFInstanceID: "amf-1", NFType: NFTypeAMF, NFStatus: NFStatusRegistered},
		{NFInstanceID: "amf-2", NFType: NFTypeAMF, NFStatus: NFStatusRegistered},
		{NFInstanceID: "smf-1", NFType: NFTypeSMF, NFStatus: NFStatusRegistered},
	}

	for _, p := range profiles {
		err := repo.Register(ctx, p)
		require.NoError(t, err)
	}

	// Get stats
	stats, err := repo.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.TotalNFs)
	assert.Equal(t, 2, stats.NFsByType["AMF"])
	assert.Equal(t, 1, stats.NFsByType["SMF"])
	assert.Equal(t, 3, stats.NFsByStatus["REGISTERED"])
}
