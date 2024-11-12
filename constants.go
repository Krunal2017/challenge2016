package main

const (
	ErrDistributorNotExist   = "Error: distributor does not exist"
	ErrDistributorExists     = "Error: distributor already exists"
	ErrInvalidJSONPayload    = "Invalid JSON payload"
	ErrInvalidJSONResponse    = "Invalid JSON response"
	ErrMissingParams         = "Missing 'distributor' or 'location' parameter"
	ErrMethodNotAllowed      = "Only GET & POST methods are allowed"
	MsgDistributorCreated    = "Distributor created successfully"
	MsgAccessGranted         = "YES"
	MsgAccessDenied          = "NO"
)
