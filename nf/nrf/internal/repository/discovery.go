package repository

import (
	"strings"
)

// DiscoveryQuery represents NF discovery search criteria (TS 29.510)
type DiscoveryQuery struct {
	NFType        NFType   `json:"target-nf-type,omitempty"`
	Limit         int      `json:"limit,omitempty"`
	PLMNID        *PLMNID  `json:"requester-plmn,omitempty"`
	SNSSAIs       []SNSSAI `json:"target-snssai-list,omitempty"`
	ServiceNames  []string `json:"service-names,omitempty"`
	RequesterFQDN string   `json:"requester-nf-fqdn,omitempty"`
	TargetNFID    string   `json:"target-nf-instance-id,omitempty"`

	// AMF-specific
	GUAMIs      []GUAMI `json:"guamis,omitempty"`
	AMFRegionID string  `json:"target-amf-region-id,omitempty"`
	AMFSetID    string  `json:"target-amf-set-id,omitempty"`

	// SMF-specific
	DNN string `json:"dnn,omitempty"`
	TAI *TAI   `json:"tai,omitempty"`

	// UPF-specific
	UPFCapability string `json:"upf-iwk-eps-ind,omitempty"`
}

// Matches checks if an NF profile matches the discovery query
func (q *DiscoveryQuery) Matches(profile *NFProfile) bool {
	// Only return registered NFs
	if profile.NFStatus != NFStatusRegistered {
		return false
	}

	// Check if expired
	if profile.IsExpired() {
		return false
	}

	// Match NF Type
	if q.NFType != "" && profile.NFType != q.NFType {
		return false
	}

	// Match specific NF instance ID
	if q.TargetNFID != "" && profile.NFInstanceID != q.TargetNFID {
		return false
	}

	// Match PLMN ID
	if q.PLMNID != nil && profile.PLMNID != nil {
		if !q.matchesPLMNID(profile.PLMNID) {
			return false
		}
	}

	// Match S-NSSAIs
	if len(q.SNSSAIs) > 0 {
		if !q.matchesSNSSAIs(profile.SNSSAIs) {
			return false
		}
	}

	// Match Service Names
	if len(q.ServiceNames) > 0 {
		if !q.matchesServiceNames(profile.NFServices) {
			return false
		}
	}

	// AMF-specific matching
	if q.NFType == NFTypeAMF && profile.AMFInfo != nil {
		if q.AMFRegionID != "" && profile.AMFInfo.AMFRegionID != q.AMFRegionID {
			return false
		}
		if q.AMFSetID != "" && profile.AMFInfo.AMFSetID != q.AMFSetID {
			return false
		}
		if len(q.GUAMIs) > 0 && !q.matchesGUAMIs(profile.AMFInfo.GUAMIList) {
			return false
		}
		if q.TAI != nil && !q.matchesTAI(profile.AMFInfo.TaiList) {
			return false
		}
	}

	// SMF-specific matching
	if q.NFType == NFTypeSMF && profile.SMFInfo != nil {
		if q.DNN != "" && !q.matchesDNN(profile.SMFInfo) {
			return false
		}
		if q.TAI != nil && !q.matchesTAI(profile.SMFInfo.TaiList) {
			return false
		}
	}

	return true
}

// matchesPLMNID checks if PLMN IDs match
func (q *DiscoveryQuery) matchesPLMNID(plmnID *PLMNID) bool {
	return q.PLMNID.MCC == plmnID.MCC && q.PLMNID.MNC == plmnID.MNC
}

// matchesSNSSAIs checks if any S-NSSAI matches
func (q *DiscoveryQuery) matchesSNSSAIs(snssais []SNSSAI) bool {
	for _, querySnssai := range q.SNSSAIs {
		for _, profileSnssai := range snssais {
			if querySnssai.SST == profileSnssai.SST {
				// If SD is specified, it must match
				if querySnssai.SD == "" || querySnssai.SD == profileSnssai.SD {
					return true
				}
			}
		}
	}
	return false
}

// matchesServiceNames checks if any service name matches
func (q *DiscoveryQuery) matchesServiceNames(services []NFService) bool {
	for _, queryService := range q.ServiceNames {
		for _, profileService := range services {
			if strings.EqualFold(queryService, profileService.ServiceName) {
				return true
			}
		}
	}
	return false
}

// matchesGUAMIs checks if any GUAMI matches
func (q *DiscoveryQuery) matchesGUAMIs(guamis []GUAMI) bool {
	for _, queryGuami := range q.GUAMIs {
		for _, profileGuami := range guamis {
			if q.matchesGUAMI(&queryGuami, &profileGuami) {
				return true
			}
		}
	}
	return false
}

// matchesGUAMI checks if a single GUAMI matches
func (q *DiscoveryQuery) matchesGUAMI(query, profile *GUAMI) bool {
	return query.PLMNID.MCC == profile.PLMNID.MCC &&
		query.PLMNID.MNC == profile.PLMNID.MNC &&
		query.AMFRegionID == profile.AMFRegionID &&
		query.AMFSetID == profile.AMFSetID
}

// matchesTAI checks if any TAI matches
func (q *DiscoveryQuery) matchesTAI(tais []TAI) bool {
	for _, profileTAI := range tais {
		if q.TAI.PLMNID.MCC == profileTAI.PLMNID.MCC &&
			q.TAI.PLMNID.MNC == profileTAI.PLMNID.MNC &&
			q.TAI.TAC == profileTAI.TAC {
			return true
		}
	}
	return false
}

// matchesDNN checks if the SMF supports the requested DNN
func (q *DiscoveryQuery) matchesDNN(smfInfo *SMFInfo) bool {
	for _, snssaiInfo := range smfInfo.SMFInfoList {
		for _, dnn := range snssaiInfo.DNNList {
			if dnn == q.DNN {
				return true
			}
		}
	}
	return false
}
