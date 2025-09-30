package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/5g-network/nf/udr/internal/clickhouse"
	"go.uber.org/zap"
)

// Repository defines the UDR data repository interface (TS 29.504, 29.505)
type Repository interface {
	// Subscriber Data Management (TS 29.505)
	CreateSubscriber(ctx context.Context, data *SubscriberData) error
	GetSubscriber(ctx context.Context, supi string) (*SubscriberData, error)
	UpdateSubscriber(ctx context.Context, supi string, data *SubscriberData) error
	DeleteSubscriber(ctx context.Context, supi string) error
	ListSubscribers(ctx context.Context, limit, offset int) ([]*SubscriberData, error)

	// Authentication Subscription Data (TS 29.503)
	CreateAuthenticationSubscription(ctx context.Context, data *AuthenticationSubscription) error
	GetAuthenticationSubscription(ctx context.Context, supi string) (*AuthenticationSubscription, error)
	UpdateAuthenticationSubscription(ctx context.Context, supi string, data *AuthenticationSubscription) error
	DeleteAuthenticationSubscription(ctx context.Context, supi string) error
	IncrementSQN(ctx context.Context, supi string) (uint64, error)

	// Session Management Subscription Data
	CreateSMSubscription(ctx context.Context, data *SessionManagementSubscriptionData) error
	GetSMSubscription(ctx context.Context, supi, dnn string) (*SessionManagementSubscriptionData, error)
	UpdateSMSubscription(ctx context.Context, supi, dnn string, data *SessionManagementSubscriptionData) error
	DeleteSMSubscription(ctx context.Context, supi, dnn string) error
	ListSMSubscriptions(ctx context.Context, supi string) ([]*SessionManagementSubscriptionData, error)

	// SDM Subscriptions (for notifications)
	CreateSDMSubscription(ctx context.Context, sub *SDMSubscription) error
	GetSDMSubscription(ctx context.Context, subscriptionID string) (*SDMSubscription, error)
	DeleteSDMSubscription(ctx context.Context, subscriptionID string) error

	// Policy Data
	CreatePolicyData(ctx context.Context, data *PolicyData) error
	GetPolicyData(ctx context.Context, supi string) (*PolicyData, error)
	UpdatePolicyData(ctx context.Context, supi string, data *PolicyData) error

	// Health
	Ping(ctx context.Context) error
	GetStats(ctx context.Context) (*Stats, error)
}

// ClickHouseRepository implements Repository using ClickHouse
type ClickHouseRepository struct {
	client *clickhouse.Client
	logger *zap.Logger
}

// NewClickHouseRepository creates a new ClickHouse-based repository
func NewClickHouseRepository(client *clickhouse.Client, logger *zap.Logger) *ClickHouseRepository {
	return &ClickHouseRepository{
		client: client,
		logger: logger,
	}
}

// CreateSubscriber creates a new subscriber
func (r *ClickHouseRepository) CreateSubscriber(ctx context.Context, data *SubscriberData) error {
	now := time.Now()
	data.CreatedAt = now
	data.UpdatedAt = now

	// Marshal NSSAI and DNN configs
	nssaiJSON, err := data.MarshalNSSAI()
	if err != nil {
		return fmt.Errorf("failed to marshal NSSAI: %w", err)
	}

	dnnJSON, err := data.MarshalDNNConfigurations()
	if err != nil {
		return fmt.Errorf("failed to marshal DNN configurations: %w", err)
	}

	query := `
		INSERT INTO udr.subscribers (
			supi, supi_type, plmn_id_mcc, plmn_id_mnc,
			subscriber_status, msisdn,
			subscribed_ue_ambr_uplink, subscribed_ue_ambr_downlink,
			nssai, dnn_configurations,
			roaming_allowed, roaming_areas,
			opc_key, authentication_method,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err = r.client.Exec(ctx, query,
		data.SUPI, data.SUPIType, data.PLMNIDmcc, data.PLMNIDmnc,
		data.SubscriberStatus, data.MSISDN,
		data.SubscribedUeAmbrUplink, data.SubscribedUeAmbrDownlink,
		nssaiJSON, dnnJSON,
		data.RoamingAllowed, data.RoamingAreas,
		data.OPCKey, data.AuthenticationMethod,
		data.CreatedAt, data.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create subscriber: %w", err)
	}

	r.logger.Info("Subscriber created", zap.String("supi", data.SUPI))
	return nil
}

// GetSubscriber retrieves a subscriber by SUPI
func (r *ClickHouseRepository) GetSubscriber(ctx context.Context, supi string) (*SubscriberData, error) {
	query := `
		SELECT 
			supi, supi_type, plmn_id_mcc, plmn_id_mnc,
			subscriber_status, msisdn,
			subscribed_ue_ambr_uplink, subscribed_ue_ambr_downlink,
			nssai, dnn_configurations,
			roaming_allowed, roaming_areas,
			opc_key, authentication_method,
			created_at, updated_at
		FROM udr.subscribers
		WHERE supi = ?
		ORDER BY updated_at DESC
		LIMIT 1
	`

	var data SubscriberData
	var nssaiJSON, dnnJSON string

	row := r.client.QueryRow(ctx, query, supi)
	err := row.Scan(
		&data.SUPI, &data.SUPIType, &data.PLMNIDmcc, &data.PLMNIDmnc,
		&data.SubscriberStatus, &data.MSISDN,
		&data.SubscribedUeAmbrUplink, &data.SubscribedUeAmbrDownlink,
		&nssaiJSON, &dnnJSON,
		&data.RoamingAllowed, &data.RoamingAreas,
		&data.OPCKey, &data.AuthenticationMethod,
		&data.CreatedAt, &data.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("subscriber not found: %w", err)
	}

	// Unmarshal JSON fields
	if err := data.UnmarshalNSSAI(nssaiJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal NSSAI: %w", err)
	}

	if err := data.UnmarshalDNNConfigurations(dnnJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DNN configurations: %w", err)
	}

	return &data, nil
}

// UpdateSubscriber updates an existing subscriber
func (r *ClickHouseRepository) UpdateSubscriber(ctx context.Context, supi string, data *SubscriberData) error {
	data.UpdatedAt = time.Now()

	// Marshal NSSAI and DNN configs
	nssaiJSON, err := data.MarshalNSSAI()
	if err != nil {
		return fmt.Errorf("failed to marshal NSSAI: %w", err)
	}

	dnnJSON, err := data.MarshalDNNConfigurations()
	if err != nil {
		return fmt.Errorf("failed to marshal DNN configurations: %w", err)
	}

	// In ClickHouse with ReplacingMergeTree, we INSERT with same key to update
	query := `
		INSERT INTO udr.subscribers (
			supi, supi_type, plmn_id_mcc, plmn_id_mnc,
			subscriber_status, msisdn,
			subscribed_ue_ambr_uplink, subscribed_ue_ambr_downlink,
			nssai, dnn_configurations,
			roaming_allowed, roaming_areas,
			opc_key, authentication_method,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err = r.client.Exec(ctx, query,
		data.SUPI, data.SUPIType, data.PLMNIDmcc, data.PLMNIDmnc,
		data.SubscriberStatus, data.MSISDN,
		data.SubscribedUeAmbrUplink, data.SubscribedUeAmbrDownlink,
		nssaiJSON, dnnJSON,
		data.RoamingAllowed, data.RoamingAreas,
		data.OPCKey, data.AuthenticationMethod,
		data.CreatedAt, data.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update subscriber: %w", err)
	}

	r.logger.Info("Subscriber updated", zap.String("supi", supi))
	return nil
}

// DeleteSubscriber deletes a subscriber (marks as deleted in ClickHouse)
func (r *ClickHouseRepository) DeleteSubscriber(ctx context.Context, supi string) error {
	// In ClickHouse, we typically don't delete but mark as inactive
	query := `
		ALTER TABLE udr.subscribers
		DELETE WHERE supi = ?
	`

	err := r.client.Exec(ctx, query, supi)
	if err != nil {
		return fmt.Errorf("failed to delete subscriber: %w", err)
	}

	r.logger.Info("Subscriber deleted", zap.String("supi", supi))
	return nil
}

// ListSubscribers lists subscribers with pagination
func (r *ClickHouseRepository) ListSubscribers(ctx context.Context, limit, offset int) ([]*SubscriberData, error) {
	query := `
		SELECT 
			supi, supi_type, plmn_id_mcc, plmn_id_mnc,
			subscriber_status, msisdn,
			subscribed_ue_ambr_uplink, subscribed_ue_ambr_downlink,
			nssai, dnn_configurations,
			roaming_allowed, roaming_areas,
			opc_key, authentication_method,
			created_at, updated_at
		FROM udr.subscribers
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.client.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscribers: %w", err)
	}
	defer rows.Close()

	var subscribers []*SubscriberData
	for rows.Next() {
		var data SubscriberData
		var nssaiJSON, dnnJSON string

		err := rows.Scan(
			&data.SUPI, &data.SUPIType, &data.PLMNIDmcc, &data.PLMNIDmnc,
			&data.SubscriberStatus, &data.MSISDN,
			&data.SubscribedUeAmbrUplink, &data.SubscribedUeAmbrDownlink,
			&nssaiJSON, &dnnJSON,
			&data.RoamingAllowed, &data.RoamingAreas,
			&data.OPCKey, &data.AuthenticationMethod,
			&data.CreatedAt, &data.UpdatedAt,
		)

		if err != nil {
			r.logger.Error("Failed to scan subscriber", zap.Error(err))
			continue
		}

		// Unmarshal JSON fields
		data.UnmarshalNSSAI(nssaiJSON)
		data.UnmarshalDNNConfigurations(dnnJSON)

		subscribers = append(subscribers, &data)
	}

	return subscribers, nil
}

// CreateAuthenticationSubscription creates authentication subscription data
func (r *ClickHouseRepository) CreateAuthenticationSubscription(ctx context.Context, data *AuthenticationSubscription) error {
	now := time.Now()
	data.CreatedAt = now
	data.UpdatedAt = now

	query := `
		INSERT INTO udr.authentication_subscription (
			supi, authentication_method,
			permanent_key, permanent_key_id,
			enc_algorithm, enc_opc, enc_op,
			sqn, sqn_scheme,
			authentication_management_field,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err := r.client.Exec(ctx, query,
		data.SUPI, data.AuthenticationMethod,
		data.PermanentKey, data.PermanentKeyID,
		data.EncAlgorithm, data.EncOPC, data.EncOP,
		data.SQN, data.SQNScheme,
		data.AuthenticationManagementField,
		data.CreatedAt, data.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create authentication subscription: %w", err)
	}

	r.logger.Info("Authentication subscription created", zap.String("supi", data.SUPI))
	return nil
}

// GetAuthenticationSubscription retrieves authentication subscription data
func (r *ClickHouseRepository) GetAuthenticationSubscription(ctx context.Context, supi string) (*AuthenticationSubscription, error) {
	query := `
		SELECT 
			supi, authentication_method,
			permanent_key, permanent_key_id,
			enc_algorithm, enc_opc, enc_op,
			sqn, sqn_scheme,
			authentication_management_field,
			created_at, updated_at
		FROM udr.authentication_subscription
		WHERE supi = ?
		ORDER BY updated_at DESC
		LIMIT 1
	`

	var data AuthenticationSubscription
	row := r.client.QueryRow(ctx, query, supi)

	err := row.Scan(
		&data.SUPI, &data.AuthenticationMethod,
		&data.PermanentKey, &data.PermanentKeyID,
		&data.EncAlgorithm, &data.EncOPC, &data.EncOP,
		&data.SQN, &data.SQNScheme,
		&data.AuthenticationManagementField,
		&data.CreatedAt, &data.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("authentication subscription not found: %w", err)
	}

	return &data, nil
}

// UpdateAuthenticationSubscription updates authentication subscription data
func (r *ClickHouseRepository) UpdateAuthenticationSubscription(ctx context.Context, supi string, data *AuthenticationSubscription) error {
	data.UpdatedAt = time.Now()

	query := `
		INSERT INTO udr.authentication_subscription (
			supi, authentication_method,
			permanent_key, permanent_key_id,
			enc_algorithm, enc_opc, enc_op,
			sqn, sqn_scheme,
			authentication_management_field,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err := r.client.Exec(ctx, query,
		data.SUPI, data.AuthenticationMethod,
		data.PermanentKey, data.PermanentKeyID,
		data.EncAlgorithm, data.EncOPC, data.EncOP,
		data.SQN, data.SQNScheme,
		data.AuthenticationManagementField,
		data.CreatedAt, data.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update authentication subscription: %w", err)
	}

	r.logger.Info("Authentication subscription updated", zap.String("supi", supi))
	return nil
}

// DeleteAuthenticationSubscription deletes authentication subscription
func (r *ClickHouseRepository) DeleteAuthenticationSubscription(ctx context.Context, supi string) error {
	query := `
		ALTER TABLE udr.authentication_subscription
		DELETE WHERE supi = ?
	`

	err := r.client.Exec(ctx, query, supi)
	if err != nil {
		return fmt.Errorf("failed to delete authentication subscription: %w", err)
	}

	r.logger.Info("Authentication subscription deleted", zap.String("supi", supi))
	return nil
}

// IncrementSQN atomically increments the SQN for a subscriber
func (r *ClickHouseRepository) IncrementSQN(ctx context.Context, supi string) (uint64, error) {
	// Get current SQN
	authSub, err := r.GetAuthenticationSubscription(ctx, supi)
	if err != nil {
		return 0, err
	}

	// Increment
	newSQN := authSub.SQN + 1
	authSub.SQN = newSQN

	// Update
	if err := r.UpdateAuthenticationSubscription(ctx, supi, authSub); err != nil {
		return 0, err
	}

	return newSQN, nil
}

// Ping checks database connectivity
func (r *ClickHouseRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx)
}

// GetStats returns repository statistics
func (r *ClickHouseRepository) GetStats(ctx context.Context) (*Stats, error) {
	query := `
		SELECT 
			COUNT(*) as total_subscribers,
			COUNT(DISTINCT plmn_id_mcc) as total_plmns
		FROM udr.subscribers
	`

	var stats Stats
	row := r.client.QueryRow(ctx, query)
	err := row.Scan(&stats.TotalSubscribers, &stats.TotalPLMNs)

	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return &stats, nil
}

// CreateSMSubscription, GetSMSubscription, etc. would follow similar patterns
// Implementing stubs for now

func (r *ClickHouseRepository) CreateSMSubscription(ctx context.Context, data *SessionManagementSubscriptionData) error {
	// TODO: Implement
	return nil
}

func (r *ClickHouseRepository) GetSMSubscription(ctx context.Context, supi, dnn string) (*SessionManagementSubscriptionData, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (r *ClickHouseRepository) UpdateSMSubscription(ctx context.Context, supi, dnn string, data *SessionManagementSubscriptionData) error {
	// TODO: Implement
	return nil
}

func (r *ClickHouseRepository) DeleteSMSubscription(ctx context.Context, supi, dnn string) error {
	// TODO: Implement
	return nil
}

func (r *ClickHouseRepository) ListSMSubscriptions(ctx context.Context, supi string) ([]*SessionManagementSubscriptionData, error) {
	// TODO: Implement
	return nil, nil
}

func (r *ClickHouseRepository) CreateSDMSubscription(ctx context.Context, sub *SDMSubscription) error {
	// TODO: Implement
	return nil
}

func (r *ClickHouseRepository) GetSDMSubscription(ctx context.Context, subscriptionID string) (*SDMSubscription, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (r *ClickHouseRepository) DeleteSDMSubscription(ctx context.Context, subscriptionID string) error {
	// TODO: Implement
	return nil
}

func (r *ClickHouseRepository) CreatePolicyData(ctx context.Context, data *PolicyData) error {
	// TODO: Implement
	return nil
}

func (r *ClickHouseRepository) GetPolicyData(ctx context.Context, supi string) (*PolicyData, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (r *ClickHouseRepository) UpdatePolicyData(ctx context.Context, supi string, data *PolicyData) error {
	// TODO: Implement
	return nil
}

// Stats represents repository statistics
type Stats struct {
	TotalSubscribers int `json:"total_subscribers"`
	TotalPLMNs       int `json:"total_plmns"`
}
