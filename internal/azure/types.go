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
	ID              string
	DisplayName     string
	Description     string
	Status          ActivationStatus
	ExpiresAt       *time.Time
	MaxDuration     time.Duration
	LinkedRoles     []LinkedRole     // Entra ID roles tied to this group
	LinkedAzureRBac []LinkedAzureRole // Azure RBAC roles tied to this group
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
	CustomerTenant  string
	LinkedGroupID   string
	LinkedGroupName string
	Status          ActivationStatus
}
