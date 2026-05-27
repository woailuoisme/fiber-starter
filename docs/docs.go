package docs

import _ "embed"

// OpenAPISpec contains the raw bytes of the generated Swagger/OpenAPI 2.0 specification.
//
//go:embed openapi.json
var OpenAPISpec []byte
