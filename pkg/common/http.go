// Package common provides shared types and utilities for HTTP operations across all Langfuse clients.
//
// This package contains common data structures used by multiple API clients,
// such as pagination metadata and shared HTTP utilities.
package common

// ListMetadata contains pagination information for list operations.
//
// This structure is included in API responses that return lists of items,
// providing information about the current page, limits, and total counts.
type ListMetadata struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}
