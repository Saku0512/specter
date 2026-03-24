package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"strings"

	"github.com/Saku0512/specter/config"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/gin-gonic/gin"
)

// buildOpenAPIRouter loads an OpenAPI spec and returns both the doc and a router.
// Returns (nil, nil) if specPath is empty or invalid.
func buildOpenAPIRouter(specPath string) (*openapi3.T, routers.Router) {
	if specPath == "" {
		return nil, nil
	}
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		log.Printf("openapi: failed to load spec %q: %v", specPath, err)
		return nil, nil
	}
	if err := doc.Validate(loader.Context); err != nil {
		log.Printf("openapi: spec %q is invalid: %v", specPath, err)
		return nil, nil
	}
	r, err := legacy.NewRouter(doc)
	if err != nil {
		log.Printf("openapi: failed to build router from %q: %v", specPath, err)
		return nil, nil
	}
	return doc, r
}

// schemaTypeName returns the first type string from an OpenAPI schema (e.g. "string", "object").
func schemaTypeName(s *openapi3.Schema) string {
	if s.Type == nil {
		return ""
	}
	for _, t := range *s.Type {
		return t
	}
	return ""
}

// randomValueFromSchema generates a random value conforming to an OpenAPI SchemaRef.
func randomValueFromSchema(ref *openapi3.SchemaRef) any {
	if ref == nil || ref.Value == nil {
		return nil
	}
	s := ref.Value

	// Use example if available
	if s.Example != nil {
		return s.Example
	}
	// Use first enum value
	if len(s.Enum) > 0 {
		return s.Enum[rand.IntN(len(s.Enum))]
	}

	switch schemaTypeName(s) {
	case "boolean":
		return gofakeit.Bool()
	case "integer":
		min, max := int64(1), int64(100)
		if s.Min != nil {
			min = int64(*s.Min)
		}
		if s.Max != nil {
			max = int64(*s.Max)
		}
		return min + rand.Int64N(max-min+1)
	case "number":
		return gofakeit.Float64Range(1, 100)
	case "array":
		n := 2
		if s.MinItems > 0 {
			n = int(s.MinItems)
		}
		arr := make([]any, n)
		for i := range arr {
			arr[i] = randomValueFromSchema(s.Items)
		}
		return arr
	case "object":
		obj := map[string]any{}
		for name, propRef := range s.Properties {
			obj[name] = randomValueFromSchema(propRef)
		}
		return obj
	default: // string
		if s.Format != "" {
			switch s.Format {
			case "email":
				return gofakeit.Email()
			case "uuid":
				return gofakeit.UUID()
			case "date-time":
				return gofakeit.Date().Format("2006-01-02T15:04:05Z")
			case "date":
				return gofakeit.Date().Format("2006-01-02")
			case "uri", "url":
				return gofakeit.URL()
			}
		}
		if s.Pattern != "" {
			return gofakeit.LetterN(8)
		}
		return gofakeit.Word()
	}
}

// randomResponseBody generates a random response body from the OpenAPI spec for the given request.
// Returns (body, contentType, ok). ok is false if no schema was found.
func randomResponseBody(req *http.Request, router routers.Router) (any, string, bool) {
	if router == nil {
		return nil, "", false
	}
	route, _, err := router.FindRoute(req)
	if err != nil {
		return nil, "", false
	}
	op := route.Operation
	if op == nil {
		return nil, "", false
	}
	// Pick first 2xx response, fall back to "default"
	for _, code := range []string{"200", "201", "202", "204", "default"} {
		resp := op.Responses.Value(code)
		if resp == nil {
			continue
		}
		if resp.Value == nil {
			continue
		}
		for ct, mt := range resp.Value.Content {
			if mt.Schema == nil {
				continue
			}
			return randomValueFromSchema(mt.Schema), ct, true
		}
		// response defined but no content (e.g. 204)
		return nil, "", true
	}
	return nil, "", false
}

// routeHasBody reports whether a config route has an explicit response configured.
func routeHasBody(rt config.Route) bool {
	return rt.Response != nil || rt.File != "" || rt.Script != "" || len(rt.Responses) > 0
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
