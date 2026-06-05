package response

import (
	"fmt"
	"time"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool         `json:"success"`
	Error   *ErrorDetail `json:"error"`
}

// ErrorDetail contains detailed error information
type ErrorDetail struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []ErrorIssue `json:"details,omitempty"`
	Chain   []string     `json:"chain,omitempty"`
	// TraceID dual-emits as both the legacy camelCase "traceId" and the
	// canonical snake_case "trace_id". The two keys always carry the same
	// value; "traceId" is retained as a deprecated alias so existing
	// consumers do not break. The constructors (Error and its helpers)
	// populate both fields; prefer them over building ErrorDetail directly.
	TraceID      string `json:"traceId"`
	TraceIDSnake string `json:"trace_id"`
	Timestamp    string `json:"timestamp"`
}

// ErrorIssue represents a specific error issue
type ErrorIssue struct {
	Service string `json:"service"`
	Issue   string `json:"issue"`
	Err     string `json:"err,omitempty"`
	Code    string `json:"code,omitempty"`
}

// Pagination represents pagination information.
//
// It dual-emits the canonical 8-field snake_case shape alongside the three
// legacy aliases (total, page, page_size). Build it with NewPagination, which
// populates every field consistently; constructing it by hand risks the
// canonical and legacy values drifting apart.
//
// Canonical fields:
//
//	current_page   - the requested page (1-based)
//	items_per_page - page size
//	total_items    - total record count across all pages
//	total_pages    - number of pages
//	has_next_page  - whether a page after current_page exists
//	has_prev_page  - whether a page before current_page exists
//	next_page      - the next page number, or null when has_next_page is false
//	prev_page      - the prev page number, or null when has_prev_page is false
//
// Legacy aliases (deprecated, kept so existing consumers do not break):
//
//	total     == total_items
//	page      == current_page
//	page_size == items_per_page
type Pagination struct {
	// Canonical 8-field snake_case shape.
	CurrentPage  int  `json:"current_page"`
	ItemsPerPage int  `json:"items_per_page"`
	TotalItems   int  `json:"total_items"`
	TotalPages   int  `json:"total_pages"`
	HasNextPage  bool `json:"has_next_page"`
	HasPrevPage  bool `json:"has_prev_page"`
	NextPage     *int `json:"next_page"`
	PrevPage     *int `json:"prev_page"`

	// Legacy aliases (deprecated): retained for backward compatibility.
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// PaginatedData represents paginated list data
type PaginatedData struct {
	Items      interface{} `json:"items"`
	Pagination Pagination  `json:"pagination"`
}

// AsyncJobData represents async job response data
type AsyncJobData struct {
	JobID            string `json:"job_id"`
	Status           string `json:"status"`
	EstimatedSeconds int    `json:"estimated_seconds,omitempty"`
}

// Success creates a successful response with data
func Success(data interface{}) *SuccessResponse {
	return &SuccessResponse{
		Success: true,
		Data:    data,
	}
}

// SuccessList creates a successful paginated list response
func SuccessList(items interface{}, pagination Pagination) *SuccessResponse {
	return &SuccessResponse{
		Success: true,
		Data: PaginatedData{
			Items:      items,
			Pagination: pagination,
		},
	}
}

// SuccessAsync creates a successful async job response
func SuccessAsync(jobID, status string, estimatedSeconds int) *SuccessResponse {
	return &SuccessResponse{
		Success: true,
		Data: AsyncJobData{
			JobID:            jobID,
			Status:           status,
			EstimatedSeconds: estimatedSeconds,
		},
	}
}

// Error creates an error response
func Error(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return &ErrorResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:         code,
			Message:      message,
			Details:      details,
			TraceID:      traceID,
			TraceIDSnake: traceID,
			Timestamp:    time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// NewPagination creates a pagination object with calculated total pages.
//
// It populates the canonical 8-field snake_case shape (current_page,
// items_per_page, total_items, total_pages, has_next_page, has_prev_page,
// next_page, prev_page) and the three legacy aliases (total, page, page_size)
// with consistent values. next_page/prev_page are nil (serialized as JSON null)
// at the first/last page boundaries.
func NewPagination(total, page, pageSize int) Pagination {
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}
	if totalPages == 0 && total > 0 {
		totalPages = 1
	}

	hasNext := page < totalPages
	hasPrev := page > 1 && totalPages > 0

	var nextPage, prevPage *int
	if hasNext {
		n := page + 1
		nextPage = &n
	}
	if hasPrev {
		p := page - 1
		prevPage = &p
	}

	return Pagination{
		// Canonical 8-field shape.
		CurrentPage:  page,
		ItemsPerPage: pageSize,
		TotalItems:   total,
		TotalPages:   totalPages,
		HasNextPage:  hasNext,
		HasPrevPage:  hasPrev,
		NextPage:     nextPage,
		PrevPage:     prevPage,

		// Legacy aliases.
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// Common Error Codes
//
// Format: {ZONE}-{SERVICE}-{HTTP_STATUS}
//
//	Zone B - Web Application / BFF
//	Zone C - Core Services
//	Zone X - Integration Services
//	Zone W - Workers / Background Jobs
//
// Every service block covers the full HTTP status spectrum the service may emit.
// Add new services by copying an existing block and renaming the prefix.
const (
	// ---------------------------------------------------------------
	// Zone B - Web Application (BFF)
	// B-NYB - Nayoo Web BFF
	// ---------------------------------------------------------------
	BNYBBadRequest           = "B-NYB-400"
	BNYBUnauthorized         = "B-NYB-401"
	BNYBForbidden            = "B-NYB-403"
	BNYBNotFound             = "B-NYB-404"
	BNYBNotAcceptable        = "B-NYB-406"
	BNYBRequestTimeout       = "B-NYB-408"
	BNYBConflict             = "B-NYB-409"
	BNYBGone                 = "B-NYB-410"
	BNYBPayloadTooLarge      = "B-NYB-413"
	BNYBUnsupportedMediaType = "B-NYB-415"
	BNYBValidationFailed     = "B-NYB-422"
	BNYBLocked               = "B-NYB-423"
	BNYBPreconditionRequired = "B-NYB-428"
	BNYBTooManyRequests      = "B-NYB-429"
	BNYBInternalServerError  = "B-NYB-500"
	BNYBNotImplemented       = "B-NYB-501"
	BNYBBadGateway           = "B-NYB-502"
	BNYBServiceUnavailable   = "B-NYB-503"
	BNYBGatewayTimeout       = "B-NYB-504"
	BNYBInsufficientStorage  = "B-NYB-507"

	// B-BOB - Backoffice BFF
	BBOBBadRequest           = "B-BOB-400"
	BBOBUnauthorized         = "B-BOB-401"
	BBOBForbidden            = "B-BOB-403"
	BBOBNotFound             = "B-BOB-404"
	BBOBNotAcceptable        = "B-BOB-406"
	BBOBRequestTimeout       = "B-BOB-408"
	BBOBConflict             = "B-BOB-409"
	BBOBGone                 = "B-BOB-410"
	BBOBPayloadTooLarge      = "B-BOB-413"
	BBOBUnsupportedMediaType = "B-BOB-415"
	BBOBValidationFailed     = "B-BOB-422"
	BBOBLocked               = "B-BOB-423"
	BBOBPreconditionRequired = "B-BOB-428"
	BBOBTooManyRequests      = "B-BOB-429"
	BBOBInternalServerError  = "B-BOB-500"
	BBOBNotImplemented       = "B-BOB-501"
	BBOBBadGateway           = "B-BOB-502"
	BBOBServiceUnavailable   = "B-BOB-503"
	BBOBGatewayTimeout       = "B-BOB-504"
	BBOBInsufficientStorage  = "B-BOB-507"

	// ---------------------------------------------------------------
	// Zone C - Core Services
	// C-AUT - Auth Service
	// ---------------------------------------------------------------
	CAUTBadRequest           = "C-AUT-400"
	CAUTUnauthorized         = "C-AUT-401"
	CAUTForbidden            = "C-AUT-403"
	CAUTNotFound             = "C-AUT-404"
	CAUTNotAcceptable        = "C-AUT-406"
	CAUTRequestTimeout       = "C-AUT-408"
	CAUTConflict             = "C-AUT-409"
	CAUTGone                 = "C-AUT-410"
	CAUTPayloadTooLarge      = "C-AUT-413"
	CAUTUnsupportedMediaType = "C-AUT-415"
	CAUTValidationFailed     = "C-AUT-422"
	CAUTLocked               = "C-AUT-423"
	CAUTPreconditionRequired = "C-AUT-428"
	CAUTTooManyRequests      = "C-AUT-429"
	CAUTInternalServerError  = "C-AUT-500"
	CAUTNotImplemented       = "C-AUT-501"
	CAUTBadGateway           = "C-AUT-502"
	CAUTServiceUnavailable   = "C-AUT-503"
	CAUTGatewayTimeout       = "C-AUT-504"
	CAUTInsufficientStorage  = "C-AUT-507"

	// C-LST - Listing Service
	CLSTBadRequest           = "C-LST-400"
	CLSTUnauthorized         = "C-LST-401"
	CLSTForbidden            = "C-LST-403"
	CLSTNotFound             = "C-LST-404"
	CLSTNotAcceptable        = "C-LST-406"
	CLSTRequestTimeout       = "C-LST-408"
	CLSTConflict             = "C-LST-409"
	CLSTGone                 = "C-LST-410"
	CLSTPayloadTooLarge      = "C-LST-413"
	CLSTUnsupportedMediaType = "C-LST-415"
	CLSTValidationFailed     = "C-LST-422"
	CLSTLocked               = "C-LST-423"
	CLSTPreconditionRequired = "C-LST-428"
	CLSTTooManyRequests      = "C-LST-429"
	CLSTInternalServerError  = "C-LST-500"
	CLSTNotImplemented       = "C-LST-501"
	CLSTBadGateway           = "C-LST-502"
	CLSTServiceUnavailable   = "C-LST-503"
	CLSTGatewayTimeout       = "C-LST-504"
	CLSTInsufficientStorage  = "C-LST-507"

	// C-MED - Media Service
	CMEDBadRequest           = "C-MED-400"
	CMEDUnauthorized         = "C-MED-401"
	CMEDForbidden            = "C-MED-403"
	CMEDNotFound             = "C-MED-404"
	CMEDNotAcceptable        = "C-MED-406"
	CMEDRequestTimeout       = "C-MED-408"
	CMEDConflict             = "C-MED-409"
	CMEDGone                 = "C-MED-410"
	CMEDPayloadTooLarge      = "C-MED-413"
	CMEDUnsupportedMediaType = "C-MED-415"
	CMEDValidationFailed     = "C-MED-422"
	CMEDLocked               = "C-MED-423"
	CMEDPreconditionRequired = "C-MED-428"
	CMEDTooManyRequests      = "C-MED-429"
	CMEDInternalServerError  = "C-MED-500"
	CMEDNotImplemented       = "C-MED-501"
	CMEDBadGateway           = "C-MED-502"
	CMEDServiceUnavailable   = "C-MED-503"
	CMEDGatewayTimeout       = "C-MED-504"
	CMEDInsufficientStorage  = "C-MED-507"

	// C-SRC - Search Service
	CSRCBadRequest           = "C-SRC-400"
	CSRCUnauthorized         = "C-SRC-401"
	CSRCForbidden            = "C-SRC-403"
	CSRCNotFound             = "C-SRC-404"
	CSRCNotAcceptable        = "C-SRC-406"
	CSRCRequestTimeout       = "C-SRC-408"
	CSRCConflict             = "C-SRC-409"
	CSRCGone                 = "C-SRC-410"
	CSRCPayloadTooLarge      = "C-SRC-413"
	CSRCUnsupportedMediaType = "C-SRC-415"
	CSRCValidationFailed     = "C-SRC-422"
	CSRCLocked               = "C-SRC-423"
	CSRCPreconditionRequired = "C-SRC-428"
	CSRCTooManyRequests      = "C-SRC-429"
	CSRCInternalServerError  = "C-SRC-500"
	CSRCNotImplemented       = "C-SRC-501"
	CSRCBadGateway           = "C-SRC-502"
	CSRCServiceUnavailable   = "C-SRC-503"
	CSRCGatewayTimeout       = "C-SRC-504"
	CSRCInsufficientStorage  = "C-SRC-507"

	// C-NOT - Notification Service
	CNOTBadRequest           = "C-NOT-400"
	CNOTUnauthorized         = "C-NOT-401"
	CNOTForbidden            = "C-NOT-403"
	CNOTNotFound             = "C-NOT-404"
	CNOTNotAcceptable        = "C-NOT-406"
	CNOTRequestTimeout       = "C-NOT-408"
	CNOTConflict             = "C-NOT-409"
	CNOTGone                 = "C-NOT-410"
	CNOTPayloadTooLarge      = "C-NOT-413"
	CNOTUnsupportedMediaType = "C-NOT-415"
	CNOTValidationFailed     = "C-NOT-422"
	CNOTLocked               = "C-NOT-423"
	CNOTPreconditionRequired = "C-NOT-428"
	CNOTTooManyRequests      = "C-NOT-429"
	CNOTInternalServerError  = "C-NOT-500"
	CNOTNotImplemented       = "C-NOT-501"
	CNOTBadGateway           = "C-NOT-502"
	CNOTServiceUnavailable   = "C-NOT-503"
	CNOTGatewayTimeout       = "C-NOT-504"
	CNOTInsufficientStorage  = "C-NOT-507"

	// C-RPT - Report Service
	CRPTBadRequest           = "C-RPT-400"
	CRPTUnauthorized         = "C-RPT-401"
	CRPTForbidden            = "C-RPT-403"
	CRPTNotFound             = "C-RPT-404"
	CRPTNotAcceptable        = "C-RPT-406"
	CRPTRequestTimeout       = "C-RPT-408"
	CRPTConflict             = "C-RPT-409"
	CRPTGone                 = "C-RPT-410"
	CRPTPayloadTooLarge      = "C-RPT-413"
	CRPTUnsupportedMediaType = "C-RPT-415"
	CRPTValidationFailed     = "C-RPT-422"
	CRPTLocked               = "C-RPT-423"
	CRPTPreconditionRequired = "C-RPT-428"
	CRPTTooManyRequests      = "C-RPT-429"
	CRPTInternalServerError  = "C-RPT-500"
	CRPTNotImplemented       = "C-RPT-501"
	CRPTBadGateway           = "C-RPT-502"
	CRPTServiceUnavailable   = "C-RPT-503"
	CRPTGatewayTimeout       = "C-RPT-504"
	CRPTInsufficientStorage  = "C-RPT-507"

	// C-PRD - Product Service
	CPRDBadRequest           = "C-PRD-400"
	CPRDUnauthorized         = "C-PRD-401"
	CPRDForbidden            = "C-PRD-403"
	CPRDNotFound             = "C-PRD-404"
	CPRDNotAcceptable        = "C-PRD-406"
	CPRDRequestTimeout       = "C-PRD-408"
	CPRDConflict             = "C-PRD-409"
	CPRDGone                 = "C-PRD-410"
	CPRDPayloadTooLarge      = "C-PRD-413"
	CPRDUnsupportedMediaType = "C-PRD-415"
	CPRDValidationFailed     = "C-PRD-422"
	CPRDLocked               = "C-PRD-423"
	CPRDPreconditionRequired = "C-PRD-428"
	CPRDTooManyRequests      = "C-PRD-429"
	CPRDInternalServerError  = "C-PRD-500"
	CPRDNotImplemented       = "C-PRD-501"
	CPRDBadGateway           = "C-PRD-502"
	CPRDServiceUnavailable   = "C-PRD-503"
	CPRDGatewayTimeout       = "C-PRD-504"
	CPRDInsufficientStorage  = "C-PRD-507"

	// ---------------------------------------------------------------
	// Zone X - Integration Services
	// X-INT - Integration Service
	// ---------------------------------------------------------------
	XINTBadRequest           = "X-INT-400"
	XINTUnauthorized         = "X-INT-401"
	XINTForbidden            = "X-INT-403"
	XINTNotFound             = "X-INT-404"
	XINTNotAcceptable        = "X-INT-406"
	XINTRequestTimeout       = "X-INT-408"
	XINTConflict             = "X-INT-409"
	XINTGone                 = "X-INT-410"
	XINTPayloadTooLarge      = "X-INT-413"
	XINTUnsupportedMediaType = "X-INT-415"
	XINTValidationFailed     = "X-INT-422"
	XINTLocked               = "X-INT-423"
	XINTPreconditionRequired = "X-INT-428"
	XINTTooManyRequests      = "X-INT-429"
	XINTInternalServerError  = "X-INT-500"
	XINTNotImplemented       = "X-INT-501"
	XINTBadGateway           = "X-INT-502"
	XINTServiceUnavailable   = "X-INT-503"
	XINTGatewayTimeout       = "X-INT-504"
	XINTInsufficientStorage  = "X-INT-507"

	// ---------------------------------------------------------------
	// Zone W - Workers / Background Jobs
	// ---------------------------------------------------------------
	// W-RSS - Realtime Search Sync Worker (go-realtime-search-sync)
	WRSSBadRequest           = "W-RSS-400"
	WRSSUnauthorized         = "W-RSS-401"
	WRSSForbidden            = "W-RSS-403"
	WRSSNotFound             = "W-RSS-404"
	WRSSNotAcceptable        = "W-RSS-406"
	WRSSRequestTimeout       = "W-RSS-408"
	WRSSConflict             = "W-RSS-409"
	WRSSGone                 = "W-RSS-410"
	WRSSPayloadTooLarge      = "W-RSS-413"
	WRSSUnsupportedMediaType = "W-RSS-415"
	WRSSValidationFailed     = "W-RSS-422"
	WRSSLocked               = "W-RSS-423"
	WRSSPreconditionRequired = "W-RSS-428"
	WRSSTooManyRequests      = "W-RSS-429"
	WRSSInternalServerError  = "W-RSS-500"
	WRSSNotImplemented       = "W-RSS-501"
	WRSSBadGateway           = "W-RSS-502"
	WRSSServiceUnavailable   = "W-RSS-503"
	WRSSGatewayTimeout       = "W-RSS-504"
	WRSSInsufficientStorage  = "W-RSS-507"

	// W-PRD - Product Worker (go-worker-product-service)
	WPRDBadRequest           = "W-PRD-400"
	WPRDUnauthorized         = "W-PRD-401"
	WPRDForbidden            = "W-PRD-403"
	WPRDNotFound             = "W-PRD-404"
	WPRDNotAcceptable        = "W-PRD-406"
	WPRDRequestTimeout       = "W-PRD-408"
	WPRDConflict             = "W-PRD-409"
	WPRDGone                 = "W-PRD-410"
	WPRDPayloadTooLarge      = "W-PRD-413"
	WPRDUnsupportedMediaType = "W-PRD-415"
	WPRDValidationFailed     = "W-PRD-422"
	WPRDLocked               = "W-PRD-423"
	WPRDPreconditionRequired = "W-PRD-428"
	WPRDTooManyRequests      = "W-PRD-429"
	WPRDInternalServerError  = "W-PRD-500"
	WPRDNotImplemented       = "W-PRD-501"
	WPRDBadGateway           = "W-PRD-502"
	WPRDServiceUnavailable   = "W-PRD-503"
	WPRDGatewayTimeout       = "W-PRD-504"
	WPRDInsufficientStorage  = "W-PRD-507"

	// W-LST - Listing Worker (go-worker-listing-service)
	WLSTBadRequest           = "W-LST-400"
	WLSTUnauthorized         = "W-LST-401"
	WLSTForbidden            = "W-LST-403"
	WLSTNotFound             = "W-LST-404"
	WLSTNotAcceptable        = "W-LST-406"
	WLSTRequestTimeout       = "W-LST-408"
	WLSTConflict             = "W-LST-409"
	WLSTGone                 = "W-LST-410"
	WLSTPayloadTooLarge      = "W-LST-413"
	WLSTUnsupportedMediaType = "W-LST-415"
	WLSTValidationFailed     = "W-LST-422"
	WLSTLocked               = "W-LST-423"
	WLSTPreconditionRequired = "W-LST-428"
	WLSTTooManyRequests      = "W-LST-429"
	WLSTInternalServerError  = "W-LST-500"
	WLSTNotImplemented       = "W-LST-501"
	WLSTBadGateway           = "W-LST-502"
	WLSTServiceUnavailable   = "W-LST-503"
	WLSTGatewayTimeout       = "W-LST-504"
	WLSTInsufficientStorage  = "W-LST-507"
)

// Helper functions for common errors

// BadRequest creates a 400 Bad Request error
func BadRequest(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// Unauthorized creates a 401 Unauthorized error
func Unauthorized(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// Forbidden creates a 403 Forbidden error
func Forbidden(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// NotFound creates a 404 Not Found error
func NotFound(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// NotAcceptable creates a 406 Not Acceptable error
func NotAcceptable(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// RequestTimeout creates a 408 Request Timeout error
func RequestTimeout(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// Conflict creates a 409 Conflict error
func Conflict(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// Gone creates a 410 Gone error
func Gone(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// PayloadTooLarge creates a 413 Payload Too Large error
func PayloadTooLarge(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// UnsupportedMediaType creates a 415 Unsupported Media Type error
func UnsupportedMediaType(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// ValidationFailed creates a 422 Validation Failed error
func ValidationFailed(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// Locked creates a 423 Locked error
func Locked(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// PreconditionRequired creates a 428 Precondition Required error
func PreconditionRequired(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// TooManyRequests creates a 429 Too Many Requests error
func TooManyRequests(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// InternalServerError creates a 500 Internal Server Error
func InternalServerError(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// NotImplemented creates a 501 Not Implemented error
func NotImplemented(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// BadGateway creates a 502 Bad Gateway error
func BadGateway(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// ServiceUnavailable creates a 503 Service Unavailable error
func ServiceUnavailable(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// GatewayTimeout creates a 504 Gateway Timeout error
func GatewayTimeout(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// InsufficientStorage creates a 507 Insufficient Storage error
func InsufficientStorage(code, message, traceID string, details ...ErrorIssue) *ErrorResponse {
	return Error(code, message, traceID, details...)
}

// CodeForStatus returns the canonical "<prefix>-<status>" error code for the
// given HTTP status, e.g. CodeForStatus("B-NYB", 422) == "B-NYB-422". It lets a
// BFF that resolves a downstream status dynamically emit the code that matches
// the status instead of hardcoding a single fallback. Statuses outside the
// supported 20-code set fall back to "<prefix>-500".
func CodeForStatus(prefix string, status int) string {
	switch status {
	case 400, 401, 403, 404, 406, 408, 409, 410, 413, 415,
		422, 423, 428, 429, 500, 501, 502, 503, 504, 507:
		return fmt.Sprintf("%s-%d", prefix, status)
	default:
		return fmt.Sprintf("%s-500", prefix)
	}
}

// ---------------------------------------------------------------------------
// Error chaining helpers (BFF / proxy use-cases)
// ---------------------------------------------------------------------------

// NewIssue builds an ErrorIssue with a service label, human-readable issue,
// optional raw error string, and the upstream's original error code.
//
//	resp := response.Unauthorized(
//	    response.BBOBUnauthorized,
//	    "Invalid or expired token",
//	    traceID,
//	    response.NewIssue("Backoffice BFF", "keycloak token verify failed",
//	        err.Error(), ""),
//	)
func NewIssue(service, issue, errStr, code string) ErrorIssue {
	return ErrorIssue{
		Service: service,
		Issue:   issue,
		Err:     errStr,
		Code:    code,
	}
}

// WithChain appends propagation hops to an ErrorResponse so the consumer
// can see where the error originated and how it travelled.
//
//	resp := response.InternalServerError(response.BBOBInternalServerError,
//	    "upstream failure", traceID, ...).
//	    WithChain("B-BOB", "C-LST", "C-RPT-500")
//
// Returns the same *ErrorResponse so the call can be chained inline.
func (r *ErrorResponse) WithChain(hops ...string) *ErrorResponse {
	if r == nil || r.Error == nil {
		return r
	}
	r.Error.Chain = append(r.Error.Chain, hops...)
	return r
}

// AppendIssue appends an extra ErrorIssue to the response. Useful when a
// handler wants to add context after receiving a partial error from upstream.
func (r *ErrorResponse) AppendIssue(issue ErrorIssue) *ErrorResponse {
	if r == nil || r.Error == nil {
		return r
	}
	r.Error.Details = append(r.Error.Details, issue)
	return r
}

// WrapUpstream takes an upstream error envelope (typically decoded from a
// downstream service's JSON response) and produces a new ErrorResponse whose
// outer code/message are set by the caller, while the upstream's code,
// message, details, and chain are preserved as the next hop.
//
// Typical BFF usage:
//
//	var upstream response.ErrorResponse
//	_ = json.Unmarshal(body, &upstream)
//	resp := response.WrapUpstream(
//	    response.BBOBInternalServerError,    // outer code (this layer)
//	    "upstream service error",            // outer message (FE-friendly)
//	    "Backoffice BFF",                    // this layer's service label
//	    traceID,
//	    &upstream,
//	)
//	c.JSON(http.StatusInternalServerError, resp)
//
// The resulting payload has:
//
//	error.code            = outerCode                    (e.g. B-BOB-500)
//	error.message         = outerMessage                 (FE-facing)
//	error.chain           = [outerCode, upstream.code, ...upstream.chain]
//	error.details         = upstream.details + one issue
//	                        describing the upstream hop with its
//	                        code/message/raw err string
func WrapUpstream(outerCode, outerMessage, thisService, traceID string, upstream *ErrorResponse) *ErrorResponse {
	resp := Error(outerCode, outerMessage, traceID)
	resp.Error.Chain = []string{outerCode}

	if upstream == nil || upstream.Error == nil {
		return resp
	}

	up := upstream.Error

	// Preserve the upstream's own details first
	if len(up.Details) > 0 {
		resp.Error.Details = append(resp.Error.Details, up.Details...)
	}

	// Add an explicit hop describing what came back from upstream
	resp.Error.Details = append(resp.Error.Details, ErrorIssue{
		Service: serviceFromCode(up.Code),
		Issue:   up.Message,
		Err:     firstNonEmpty(joinIssueErrs(up.Details), up.Message),
		Code:    up.Code,
	})

	// Build the chain: [outer, ...upstream.chain OR upstream.code]
	if len(up.Chain) > 0 {
		resp.Error.Chain = append(resp.Error.Chain, up.Chain...)
	} else if up.Code != "" {
		// Fallback for upstream errors that don't have a chain yet
		resp.Error.Chain = append(resp.Error.Chain, up.Code)
	}

	_ = thisService // reserved for future use
	return resp
}

// serviceFromCode extracts the service prefix from a code like "C-LST-404"
// and returns "C-LST". Returns the input unchanged if it does not look like
// a zone-prefixed code.
func serviceFromCode(code string) string {
	// Expect "{Z}-{SVC}-{HTTP}" — strip the trailing -HTTP.
	last := -1
	for i := len(code) - 1; i >= 0; i-- {
		if code[i] == '-' {
			last = i
			break
		}
	}
	if last <= 0 {
		return code
	}
	return code[:last]
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func joinIssueErrs(details []ErrorIssue) string {
	out := ""
	for _, d := range details {
		if d.Err == "" {
			continue
		}
		if out != "" {
			out += "; "
		}
		out += d.Err
	}
	return out
}
