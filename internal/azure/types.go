package azure

import "time"

type ActivationStatus int

const (
	StatusInactive ActivationStatus = iota
	StatusActive
	StatusExpiringSoon // < 30 min remaining
	StatusPending      // awaiting approval
)

func (s ActivationStatus) String() string {
	switch s {
	case StatusActive:
		return "Active"
	case StatusExpiringSoon:
		return "Expiring Soon"
	case StatusPending:
		return "Pending"
	default:
		return "Inactive"
	}
}

// IsActive returns true if the status represents an active state (Active or ExpiringSoon)
func (s ActivationStatus) IsActive() bool {
	return s == StatusActive || s == StatusExpiringSoon
}

// StatusFromExpiry returns the activation status based on expiry time
func StatusFromExpiry(expiry *time.Time) ActivationStatus {
	if expiry == nil {
		return StatusInactive
	}
	if time.Until(*expiry) < 30*time.Minute {
		return StatusExpiringSoon
	}
	return StatusActive
}

type Tenant struct {
	ID          string
	DisplayName string
}

type Role struct {
	ID               string
	DisplayName      string
	Description      string
	RoleDefinitionID string
	DirectoryScopeID string
	Status           ActivationStatus
	ExpiresAt        *time.Time
	MaxDuration      time.Duration
	Permissions      []string // Permission actions for this role
}

type Group struct {
	ID               string
	DisplayName      string
	Description      string
	RoleDefinitionID string // "member" or "owner" from eligibility response
	Status           ActivationStatus
	ExpiresAt        *time.Time
	MaxDuration      time.Duration
	LinkedRoles      []LinkedRole      // Entra ID roles tied to this group
	LinkedAzureRBac  []LinkedAzureRole // Azure RBAC roles tied to this group
}

// LinkedRole represents an Entra ID role assignment linked to a group
type LinkedRole struct {
	DisplayName      string
	RoleDefinitionID string
	Status           ActivationStatus
}

// LinkedAzureRole represents an Azure RBAC role assignment linked to a group
type LinkedAzureRole struct {
	DisplayName      string
	RoleDefinitionID string
	Scope            string // subscription/resource group path
}

type LighthouseSubscription struct {
	ID              string
	DisplayName     string
	TenantID        string // Home tenant ID of the subscription
	TenantName      string // Display name of the home tenant
	CustomerTenant  string
	LinkedGroupID   string
	LinkedGroupName string
	Status          ActivationStatus
	EligibleRoles   []EligibleAzureRole // Azure RBAC roles that can be activated
}

// EligibleAzureRole represents an Azure RBAC role that can be activated via PIM
type EligibleAzureRole struct {
	RoleDefinitionID   string // Full resource ID of the role definition
	RoleDefinitionName string // Display name (e.g., "Contributor", "Reader")
	RoleEligibilityID  string // The roleEligibilityScheduleInstance ID (needed for activation)
	Scope              string // Subscription or resource group scope
	Status             ActivationStatus
	ExpiresAt          *time.Time
}
