package context

import (
	"fmt"
	"sync"
)

// SMFContext manages all PDU sessions and UPF associations
type SMFContext struct {
	mu sync.RWMutex

	// PDU Sessions indexed by SUPI + PDU Session ID
	sessions map[string]*PDUSession

	// UPF Associations (simplified - one default UPF for now)
	upfNodeID    string
	upfN4Address string

	// Statistics
	stats Statistics
}

// Statistics tracks SMF statistics
type Statistics struct {
	TotalSessions    int `json:"totalSessions"`
	ActiveSessions   int `json:"activeSessions"`
	ReleasedSessions int `json:"releasedSessions"`
}

// NewSMFContext creates a new SMF context manager
func NewSMFContext(upfNodeID, upfN4Address string) *SMFContext {
	return &SMFContext{
		sessions:     make(map[string]*PDUSession),
		upfNodeID:    upfNodeID,
		upfN4Address: upfN4Address,
	}
}

// sessionKey generates a unique key for a PDU session
func sessionKey(supi string, pduSessionID uint8) string {
	return fmt.Sprintf("%s-%d", supi, pduSessionID)
}

// AddSession adds a new PDU session
func (c *SMFContext) AddSession(session *PDUSession) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := sessionKey(session.SUPI, session.PDUSessionID)
	if _, exists := c.sessions[key]; exists {
		return fmt.Errorf("session already exists: %s", key)
	}

	c.sessions[key] = session
	c.stats.TotalSessions++

	return nil
}

// GetSession retrieves a PDU session
func (c *SMFContext) GetSession(supi string, pduSessionID uint8) (*PDUSession, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := sessionKey(supi, pduSessionID)
	session, exists := c.sessions[key]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", key)
	}

	return session, nil
}

// RemoveSession removes a PDU session
func (c *SMFContext) RemoveSession(supi string, pduSessionID uint8) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := sessionKey(supi, pduSessionID)
	if _, exists := c.sessions[key]; !exists {
		return fmt.Errorf("session not found: %s", key)
	}

	delete(c.sessions, key)
	c.stats.ReleasedSessions++

	return nil
}

// GetAllSessions returns all PDU sessions for a SUPI
func (c *SMFContext) GetAllSessions(supi string) []*PDUSession {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var sessions []*PDUSession
	for key, session := range c.sessions {
		if session.SUPI == supi {
			sessions = append(sessions, session)
		}
		_ = key // avoid unused variable
	}

	return sessions
}

// GetActiveSessions returns all active PDU sessions
func (c *SMFContext) GetActiveSessions() []*PDUSession {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var sessions []*PDUSession
	for _, session := range c.sessions {
		if session.GetState() == PDUSessionStateActive {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// GetStatistics returns current statistics
func (c *SMFContext) GetStatistics() Statistics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Count active sessions
	activeCount := 0
	for _, session := range c.sessions {
		if session.GetState() == PDUSessionStateActive {
			activeCount++
		}
	}

	c.mu.RUnlock()
	c.mu.Lock()
	c.stats.ActiveSessions = activeCount
	stats := c.stats
	c.mu.Unlock()
	c.mu.RLock()

	return stats
}

// GetUPFInfo returns default UPF information
func (c *SMFContext) GetUPFInfo() (nodeID, n4Address string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.upfNodeID, c.upfN4Address
}
