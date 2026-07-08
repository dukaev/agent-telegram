package operations

// OpenAPI builds an OpenAPI 3.1 document for the REST RPC surface.
func OpenAPI(title, version string) map[string]any {
	paths := map[string]any{
		"/health": map[string]any{
			"get": map[string]any{
				"summary": "Health check",
				"responses": map[string]any{
					"200": response("OK", objectWith(map[string]any{"ok": JSONSchema{"type": "boolean"}})),
				},
			},
		},
		"/methods": map[string]any{
			"get": map[string]any{
				"summary":  "List registered RPC methods",
				"security": []map[string][]string{{"bearerAuth": []string{}}},
				"responses": map[string]any{
					"200": response("Methods", objectWith(map[string]any{
						"ok":      JSONSchema{"type": "boolean"},
						"methods": JSONSchema{"type": "array", "items": JSONSchema{"type": "string"}},
					})),
				},
			},
		},
		"/manifest": map[string]any{
			"get": map[string]any{
				"summary":  "Machine-readable operation manifest",
				"security": []map[string][]string{{"bearerAuth": []string{}}},
				"responses": map[string]any{
					"200": response("Manifest", JSONSchema{
						"type": "object",
						"properties": map[string]any{
							"ok":         JSONSchema{"type": "boolean"},
							"operations": JSONSchema{"type": "array", "items": JSONSchema{"type": "object", "additionalProperties": true}},
							"errorTypes": JSONSchema{"type": "array", "items": JSONSchema{"type": "object", "additionalProperties": true}},
							"skills":     JSONSchema{"type": "array", "items": JSONSchema{"type": "object", "additionalProperties": true}},
						},
					}),
				},
			},
		},
	}

	for _, manifest := range Manifest() {
		paths["/rpc/"+manifest.Method] = map[string]any{
			"post": map[string]any{
				"summary":     manifest.Summary,
				"description": "Safety: " + manifest.Safety,
				"operationId": manifest.Method,
				"tags":        []string{manifest.Category},
				"security":    []map[string][]string{{"bearerAuth": []string{}}},
				"parameters": []map[string]any{
					headerParam("X-Trace-Id", "Optional caller-supplied trace ID for one operation."),
					headerParam("X-Run-Id", "Optional caller-supplied run ID for multi-command agent correlation."),
					boolQueryParam("dryRun", "Validate and preview without executing."),
					boolQueryParam("validateOnly", "Validate params without executing."),
				},
				"requestBody": map[string]any{
					"required": true,
					"content": map[string]any{
						"application/json": map[string]any{"schema": manifest.InputSchema},
					},
				},
				"responses": map[string]any{
					"200": response("Result", objectWith(map[string]any{
						"ok":      JSONSchema{"type": "boolean"},
						"runId":   JSONSchema{"type": "string"},
						"traceId": JSONSchema{"type": "string"},
						"result":  manifest.OutputSchema,
					})),
					"400": response("Bad request", errorSchema()),
					"401": response("Unauthorized", errorSchema()),
					"403": response("Forbidden", errorSchema()),
					"404": response("Not found", errorSchema()),
					"429": response("Rate limited", errorSchema()),
					"500": response("Internal error", errorSchema()),
				},
			},
		}
	}

	return map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":   title,
			"version": version,
		},
		"paths": paths,
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"bearerAuth": map[string]any{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "token",
				},
			},
		},
	}
}

func response(description string, schema any) map[string]any {
	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{"schema": schema},
		},
	}
}

func objectWith(properties map[string]any) JSONSchema {
	return JSONSchema{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
}

func boolQueryParam(name, description string) map[string]any {
	return map[string]any{
		"name":        name,
		"in":          "query",
		"required":    false,
		"description": description,
		"schema":      JSONSchema{"type": "boolean"},
	}
}

func headerParam(name, description string) map[string]any {
	return map[string]any{
		"name":        name,
		"in":          "header",
		"required":    false,
		"description": description,
		"schema":      JSONSchema{"type": "string"},
	}
}

func errorSchema() JSONSchema {
	return objectWith(map[string]any{
		"ok":      JSONSchema{"type": "boolean"},
		"runId":   JSONSchema{"type": "string"},
		"traceId": JSONSchema{"type": "string"},
		"error": objectWith(map[string]any{
			"code":    JSONSchema{"type": "integer"},
			"message": JSONSchema{"type": "string"},
			"data":    JSONSchema{"type": "object", "additionalProperties": true},
		}),
	})
}
