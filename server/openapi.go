package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/gin-gonic/gin"
)

// buildOpenAPIRouter loads an OpenAPI spec and returns a router for use in
// both request and response validation. Returns nil if specPath is empty or invalid.
func buildOpenAPIRouter(specPath string) routers.Router {
	if specPath == "" {
		return nil
	}
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		log.Printf("openapi: failed to load spec %q: %v", specPath, err)
		return nil
	}
	if err := doc.Validate(loader.Context); err != nil {
		log.Printf("openapi: spec %q is invalid: %v", specPath, err)
		return nil
	}
	r, err := legacy.NewRouter(doc)
	if err != nil {
		log.Printf("openapi: failed to build router from %q: %v", specPath, err)
		return nil
	}
	return r
}

// openAPIRequestMiddleware validates incoming requests against a pre-built OpenAPI router.
// Non-strict: always serves the mock but adds X-Specter-Validation-Error header.
// Strict: returns 400 and aborts.
func openAPIRequestMiddleware(router routers.Router, strict bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/__specter/") {
			c.Next()
			return
		}

		// Read body for validation without consuming it
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		route, pathParams, err := router.FindRoute(c.Request)
		if err != nil {
			// Route not in spec — not an error, spec may be partial
			c.Next()
			return
		}

		input := &openapi3filter.RequestValidationInput{
			Request:    c.Request,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
			},
		}
		// Restore body after FindRoute may have consumed it
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if err := openapi3filter.ValidateRequest(c.Request.Context(), input); err != nil {
			msg := err.Error()
			log.Printf("openapi validation: %s %s — %s", c.Request.Method, c.Request.URL.Path, msg)
			if strict {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "request validation failed", "detail": msg})
				return
			}
			c.Header("X-Specter-Validation-Error", msg)
		}

		// Always restore body for the route handler
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		c.Next()
	}
}

// validateOpenAPIResponse validates a response body against the OpenAPI spec.
// Returns nil if the route is not in the spec or validation passes.
func validateOpenAPIResponse(req *http.Request, router routers.Router, status int, ct string, body any) error {
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		return nil // route not in spec — skip
	}
	reqInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		Options: &openapi3filter.Options{
			ExcludeRequestBody: true,
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	}
	if ct == "" {
		ct = "application/json"
	}
	var bodyBytes []byte
	if strings.Contains(ct, "application/json") {
		bodyBytes, _ = json.Marshal(body)
	} else if s, ok := body.(string); ok {
		bodyBytes = []byte(s)
	}
	resInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: reqInput,
		Status:                 status,
		Header:                 http.Header{"Content-Type": []string{ct}},
		Body:                   io.NopCloser(bytes.NewReader(bodyBytes)),
		Options: &openapi3filter.Options{
			ExcludeRequestBody: true,
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	}
	return openapi3filter.ValidateResponse(req.Context(), resInput)
}

// respondValidated validates the response against the OpenAPI spec (if configured)
// then writes it. In strict mode a schema violation replaces the response with 500.
func respondValidated(c *gin.Context, router routers.Router, strict bool, status int, ct string, body any) {
	if router != nil {
		if err := validateOpenAPIResponse(c.Request, router, status, ct, body); err != nil {
			msg := err.Error()
			log.Printf("[specter] openapi response validation: %s %s — %s", c.Request.Method, c.Request.URL.Path, msg)
			if strict {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "response violates OpenAPI schema", "detail": msg})
				return
			}
			c.Header("X-Specter-Response-Validation-Error", msg)
		}
	}
	respond(c, status, ct, body)
}
