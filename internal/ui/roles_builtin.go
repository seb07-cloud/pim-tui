package ui

// BuiltInRolePermissions maps role definition IDs to their key permissions
// This data is sourced from Microsoft documentation for common Entra ID built-in roles
// https://learn.microsoft.com/en-us/entra/identity/role-based-access-control/permissions-reference
var BuiltInRolePermissions = map[string][]string{
	// Global Administrator
	"62e90394-69f5-4237-9190-012177145e10": {
		"microsoft.directory/*/allTasks",
		"microsoft.azure.advancedThreatProtection/allEntities/allTasks",
		"microsoft.azure.informationProtection/allEntities/allTasks",
		"microsoft.azure.serviceHealth/allEntities/allTasks",
		"microsoft.commerce.billing/allEntities/allTasks",
		"microsoft.office365.complianceManager/allEntities/allTasks",
		"microsoft.office365.serviceHealth/allEntities/allTasks",
	},
	// Application Administrator
	"9b895d92-2cd3-44c7-9d02-a6ac2d5ea5c3": {
		"microsoft.directory/applications/create",
		"microsoft.directory/applications/delete",
		"microsoft.directory/applications/appRoles/update",
		"microsoft.directory/applications/credentials/update",
		"microsoft.directory/applications/owners/update",
		"microsoft.directory/applications/permissions/update",
		"microsoft.directory/servicePrincipals/allProperties/allTasks",
	},
	// Cloud Application Administrator
	"158c047a-c907-4556-b7ef-446551a6b5f7": {
		"microsoft.directory/applications/create",
		"microsoft.directory/applications/delete",
		"microsoft.directory/applications/appRoles/update",
		"microsoft.directory/applications/credentials/update",
		"microsoft.directory/servicePrincipals/allProperties/allTasks",
	},
	// User Administrator
	"fe930be7-5e62-47db-91af-98c3a49a38b1": {
		"microsoft.directory/users/create",
		"microsoft.directory/users/delete",
		"microsoft.directory/users/password/update",
		"microsoft.directory/users/basicProfile/update",
		"microsoft.directory/users/manager/update",
		"microsoft.directory/groups/members/update",
		"microsoft.directory/groups/create",
	},
	// Helpdesk Administrator
	"729827e3-9c14-49f7-bb1b-9608f156bbb8": {
		"microsoft.directory/users/password/update",
		"microsoft.directory/users/invalidateAllRefreshTokens",
		"microsoft.azure.serviceHealth/allEntities/allTasks",
		"microsoft.office365.serviceHealth/allEntities/allTasks",
	},
	// Security Administrator
	"194ae4cb-b126-40b2-bd5b-6091b380977d": {
		"microsoft.directory/conditionalAccessPolicies/allProperties/allTasks",
		"microsoft.directory/identityProtection/allProperties/allTasks",
		"microsoft.directory/privilegedIdentityManagement/allProperties/read",
		"microsoft.azure.advancedThreatProtection/allEntities/allTasks",
		"microsoft.office365.protectionCenter/allEntities/allTasks",
	},
	// Security Reader
	"5d6b6bb7-de71-4623-b4af-96380a352509": {
		"microsoft.directory/auditLogs/allProperties/read",
		"microsoft.directory/conditionalAccessPolicies/allProperties/read",
		"microsoft.directory/identityProtection/allProperties/read",
		"microsoft.directory/signInReports/allProperties/read",
	},
	// Privileged Role Administrator
	"e8611ab8-c189-46e8-94e1-60213ab1f814": {
		"microsoft.directory/directoryRoles/allProperties/allTasks",
		"microsoft.directory/privilegedIdentityManagement/allEntities/allTasks",
		"microsoft.directory/roleAssignments/allProperties/allTasks",
		"microsoft.directory/roleDefinitions/allProperties/allTasks",
	},
	// Privileged Authentication Administrator
	"7be44c8a-adaf-4e2a-84d6-ab2649e08a13": {
		"microsoft.directory/users/authenticationMethods/allProperties/allTasks",
		"microsoft.directory/users/password/update",
		"microsoft.directory/users/invalidateAllRefreshTokens",
	},
	// Authentication Administrator
	"c4e39bd9-1100-46d3-8c65-fb160da0071f": {
		"microsoft.directory/users/authenticationMethods/basic/update",
		"microsoft.directory/users/password/update",
		"microsoft.directory/users/invalidateAllRefreshTokens",
	},
	// Conditional Access Administrator
	"b1be1c3e-b65d-4f19-8427-f6fa0d97feb9": {
		"microsoft.directory/conditionalAccessPolicies/create",
		"microsoft.directory/conditionalAccessPolicies/delete",
		"microsoft.directory/conditionalAccessPolicies/allProperties/allTasks",
		"microsoft.directory/namedLocations/allProperties/allTasks",
	},
	// Exchange Administrator
	"29232cdf-9323-42fd-ade2-1d097af3e4de": {
		"microsoft.office365.exchange/allEntities/allTasks",
		"microsoft.directory/groups/members/update",
		"microsoft.azure.serviceHealth/allEntities/allTasks",
	},
	// SharePoint Administrator
	"f28a1f50-f6e7-4571-818b-6a12f2af6b6c": {
		"microsoft.office365.sharePoint/allEntities/allTasks",
		"microsoft.directory/groups/members/update",
		"microsoft.azure.serviceHealth/allEntities/allTasks",
	},
	// Teams Administrator
	"69091246-20e8-4a56-aa4d-066075b2a7a8": {
		"microsoft.teams/allEntities/allProperties/allTasks",
		"microsoft.directory/groups/members/update",
		"microsoft.azure.serviceHealth/allEntities/allTasks",
	},
	// Intune Administrator
	"3a2c62db-5318-420d-8d74-23affee5d9d5": {
		"microsoft.intune/allEntities/allTasks",
		"microsoft.directory/devices/allProperties/allTasks",
		"microsoft.directory/groups/members/update",
	},
	// Billing Administrator
	"b0f54661-2d74-4c50-afa3-1ec803f12efe": {
		"microsoft.commerce.billing/allEntities/allTasks",
		"microsoft.directory/organization/basic/update",
		"microsoft.azure.serviceHealth/allEntities/allTasks",
	},
	// Global Reader
	"f2ef992c-3afb-46b9-b7cf-a126ee74c451": {
		"microsoft.directory/*/read",
		"microsoft.azure.serviceHealth/allEntities/allTasks",
		"microsoft.office365.serviceHealth/allEntities/allTasks",
	},
	// Directory Readers
	"88d8e3e3-8f55-4a1e-953a-9b9898b8876b": {
		"microsoft.directory/administrativeUnits/standard/read",
		"microsoft.directory/applications/standard/read",
		"microsoft.directory/groups/standard/read",
		"microsoft.directory/users/standard/read",
	},
	// Groups Administrator
	"fdd7a751-b60b-444a-984c-02652fe8fa1c": {
		"microsoft.directory/groups/create",
		"microsoft.directory/groups/delete",
		"microsoft.directory/groups/members/update",
		"microsoft.directory/groups/owners/update",
		"microsoft.directory/groups/allProperties/allTasks",
	},
	// License Administrator
	"4d6ac14f-3453-41d0-bef9-a3e0c569773a": {
		"microsoft.directory/users/assignLicense",
		"microsoft.directory/users/reprocessLicenseAssignment",
		"microsoft.directory/groups/assignLicense",
		"microsoft.azure.serviceHealth/allEntities/allTasks",
	},
	// Password Administrator
	"966707d0-3269-4727-9be2-8c3a10f19b9d": {
		"microsoft.directory/users/password/update",
	},
	// Compliance Administrator
	"17315797-102d-40b4-93e0-432062caca18": {
		"microsoft.azure.informationProtection/allEntities/allTasks",
		"microsoft.azure.serviceHealth/allEntities/allTasks",
		"microsoft.office365.complianceManager/allEntities/allTasks",
		"microsoft.office365.serviceHealth/allEntities/allTasks",
	},
	// Reports Reader
	"4a5d8f65-41da-4de4-8968-e035b65339cf": {
		"microsoft.directory/auditLogs/allProperties/read",
		"microsoft.directory/signInReports/allProperties/read",
		"microsoft.azure.serviceHealth/allEntities/allTasks",
		"microsoft.office365.usageReports/allEntities/allProperties/read",
	},
	// Azure DevOps Administrator
	"e3973bdf-4987-49ae-837a-ba8e231c7286": {
		"microsoft.azure.devOps/allEntities/allTasks",
	},
	// Power Platform Administrator
	"11648597-926c-4cf3-9c36-bcebb0ba8dcc": {
		"microsoft.powerApps/allEntities/allTasks",
		"microsoft.flow/allEntities/allTasks",
		"microsoft.dynamics365/allEntities/allTasks",
	},
	// Fabric Administrator
	"a9ea8996-122f-4c74-9520-8edcd192826c": {
		"microsoft.fabric/allEntities/allTasks",
		"microsoft.powerBI/allEntities/allTasks",
	},
	// Identity Governance Administrator
	"45d8d3c5-c802-45c6-b32a-1d70b5e1e86e": {
		"microsoft.directory/accessReviews/allProperties/allTasks",
		"microsoft.directory/entitlementManagement/allProperties/allTasks",
		"microsoft.directory/privilegedIdentityManagement/allProperties/read",
	},
	// Knowledge Administrator
	"b5a8dcf3-09d5-43a9-a639-8e29ef291470": {
		"microsoft.office365.knowledge/allEntities/allTasks",
		"microsoft.office365.sharePoint.search/allEntities/allTasks",
	},
	// Cloud Device Administrator
	"7698a772-787b-4ac8-901f-60d6b08affd2": {
		"microsoft.directory/devices/delete",
		"microsoft.directory/devices/disable",
		"microsoft.directory/devices/enable",
		"microsoft.directory/bitlockerKeys/key/read",
	},
	// Hybrid Identity Administrator
	"8ac3fc64-6eca-42ea-9e69-59f4c7b60eb2": {
		"microsoft.directory/applications/audience/update",
		"microsoft.directory/applications/authentication/update",
		"microsoft.directory/cloudProvisioning/allProperties/allTasks",
		"microsoft.directory/domains/federation/update",
	},
}

// GetRolePermissions returns the permissions for a given role definition ID
func GetRolePermissions(roleDefinitionID string) []string {
	if perms, ok := BuiltInRolePermissions[roleDefinitionID]; ok {
		return perms
	}
	return nil
}
