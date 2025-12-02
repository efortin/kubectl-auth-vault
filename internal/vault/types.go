package vault

// OIDCTokenResponse represents the Vault response for OIDC token requests.
// This is the structure returned by Vault's identity/oidc/token endpoint.
type OIDCTokenResponse struct {
	Data OIDCTokenData `json:"data"`
}

// OIDCTokenData contains the OIDC token data from Vault.
type OIDCTokenData struct {
	Token string `json:"token"`
}
