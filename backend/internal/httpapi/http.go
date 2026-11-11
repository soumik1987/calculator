package httpapi

const (
	pathHealth       = "/healthz"
	pathCalculations = "/api/v1/calculations"

	spanOperation = "http.server"
	mediaTypeJSON = "application/json"
	allOrigins    = "*"

	headerContentType           = "Content-Type"
	headerOrigin                = "Origin"
	headerVary                  = "Vary"
	headerAllowOrigin           = "Access-Control-Allow-Origin"
	headerAllowMethods          = "Access-Control-Allow-Methods"
	headerAllowHeaders          = "Access-Control-Allow-Headers"
	headerMaxAge                = "Access-Control-Max-Age"
	headerContentTypeOptions    = "X-Content-Type-Options"
	headerContentSecurityPolicy = "Content-Security-Policy"
	headerCacheControl          = "Cache-Control"
	headerReferrerPolicy        = "Referrer-Policy"
	headerTraceID               = "X-Trace-ID"
	headerExposeHeaders         = "Access-Control-Expose-Headers"
)
