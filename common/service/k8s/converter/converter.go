package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Converter struct {
	OutputDir  string
	Prefix     string
	StandAlone bool
	Expanded   bool
	Kubernetes bool
	Strict     bool
}

type OpenAPISchema struct {
	Swagger     string                            `json:"swagger" yaml:"swagger"`
	OpenAPI     string                            `json:"openapi" yaml:"openapi"`
	Definitions map[string]interface{}            `json:"definitions" yaml:"definitions"`
	Components  map[string]map[string]interface{} `json:"components" yaml:"components"`
}

func NewConverter(outputDir, prefix string, standAlone, expanded, kubernetes, strict bool) *Converter {
	return &Converter{
		OutputDir:  outputDir,
		Prefix:     prefix,
		StandAlone: standAlone,
		Expanded:   expanded,
		Kubernetes: kubernetes,
		Strict:     strict,
	}
}

func (c *Converter) Convert(schemaURL string) error {
	// Download and parse schema
	schema, err := c.loadSchema(schemaURL)
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Process schema based on version
	if schema.Swagger != "" {
		return c.processSwagger(schema)
	} else if schema.OpenAPI != "" {
		return c.processOpenAPI(schema)
	}

	return fmt.Errorf("invalid schema format - neither swagger nor openapi version found")
}

func (c *Converter) ConvertData(data []byte) error {
	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	var schema OpenAPISchema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}
	if schema.Swagger != "" {
		return c.processSwagger(&schema)
	} else if schema.OpenAPI != "" {
		return c.processOpenAPI(&schema)
	}
	return fmt.Errorf("invalid schema format - neither swagger nor openapi version found")
}

func (c *Converter) loadSchema(schemaURL string) (*OpenAPISchema, error) {
	var reader io.ReadCloser

	if strings.HasPrefix(schemaURL, "http://") || strings.HasPrefix(schemaURL, "https://") {
		resp, err := http.Get(schemaURL)
		if err != nil {
			return nil, fmt.Errorf("failed to download schema: %w", err)
		}
		reader = resp.Body
	} else {
		// Handle file paths
		filePath := strings.TrimPrefix(schemaURL, "file://")
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open schema file: %w", err)
		}
		reader = file
	}
	defer reader.Close()

	// Parse as YAML (which can also parse JSON)
	var schema OpenAPISchema
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	return &schema, nil
}

func (c *Converter) processSwagger(schema *OpenAPISchema) error {
	// Write shared definitions
	if err := c.writeSharedDefinitions(schema.Definitions); err != nil {
		return fmt.Errorf("failed to write shared definitions: %w", err)
	}

	// Process Kubernetes extensions if enabled
	if c.Kubernetes {
		if err := c.processKubernetesExtensions(schema.Definitions); err != nil {
			return fmt.Errorf("failed to process kubernetes extensions: %w", err)
		}
	}

	// Generate individual schemas
	var types []string
	for title, definition := range schema.Definitions {
		kind := strings.ToLower(strings.Split(title, ".")[len(strings.Split(title, "."))-1])
		fullName := kind

		if c.Kubernetes && c.Expanded {
			parts := strings.Split(title, ".")
			if len(parts) >= 3 {
				group := parts[len(parts)-3]
				apiVersion := parts[len(parts)-2]
				if group == "core" || group == "api" {
					fullName = fmt.Sprintf("%s-%s", kind, apiVersion)
				} else {
					fullName = fmt.Sprintf("%s-%s-%s", kind, group, apiVersion)
				}
			}
		}

		// Convert definition to schema
		schemaDef := definition.(map[string]interface{})
		schemaDef["$schema"] = "http://json-schema.org/schema#"
		schemaDef["type"] = "object"

		if c.Strict {
			schemaDef["additionalProperties"] = false
		}

		// Write individual schema file
		if err := c.writeSchemaFile(fullName, schemaDef); err != nil {
			return fmt.Errorf("failed to write schema for %s: %w", fullName, err)
		}

		types = append(types, title)
	}

	// Write all.json with references to all types
	allRefs := make([]map[string]string, 0, len(types))
	for _, title := range types {
		allRefs = append(allRefs, map[string]string{
			"$ref": fmt.Sprintf("%s#/definitions/%s", c.Prefix, title),
		})
	}

	allSchema := map[string]interface{}{
		"oneOf": allRefs,
	}

	if err := c.writeSchemaFile("all", allSchema); err != nil {
		return fmt.Errorf("failed to write all.json: %w", err)
	}

	return nil
}

func (c *Converter) writeSharedDefinitions(definitions map[string]interface{}) error {
	if c.Kubernetes {
		// Add Kubernetes-specific definitions
		definitions["io.k8s.apimachinery.pkg.util.intstr.IntOrString"] = map[string]interface{}{
			"oneOf": []map[string]string{
				{"type": "string"},
				{"type": "integer"},
			},
		}
		definitions["io.k8s.apimachinery.pkg.api.resource.Quantity"] = map[string]interface{}{
			"oneOf": []map[string]string{
				{"type": "string"},
				{"type": "number"},
			},
		}
	}

	if c.Strict {
		definitions = c.additionalProperties(definitions)
	}

	sharedDefs := map[string]interface{}{
		"definitions": definitions,
	}

	return c.writeSchemaFile("_definitions", sharedDefs)
}

func (c *Converter) processKubernetesExtensions(definitions map[string]interface{}) error {
	for _, typeDef := range definitions {
		def := typeDef.(map[string]interface{})
		if ext, ok := def["x-kubernetes-group-version-kind"]; ok {
			kubeExts := ext.([]interface{})
			for _, ext := range kubeExts {
				kubeExt := ext.(map[string]interface{})
				if c.Expanded && def["properties"] != nil {
					props := def["properties"].(map[string]interface{})
					if apiVersionProp, ok := props["apiVersion"]; ok {
						apiVersion := ""
						if group, ok := kubeExt["group"].(string); ok && group != "" {
							apiVersion = fmt.Sprintf("%s/%s", group, kubeExt["version"])
						} else {
							apiVersion = kubeExt["version"].(string)
						}
						c.appendNoDuplicates(apiVersionProp, "enum", apiVersion)
					}
				}
				if def["properties"] != nil {
					props := def["properties"].(map[string]interface{})
					if kindProp, ok := props["kind"]; ok {
						c.appendNoDuplicates(kindProp, "enum", kubeExt["kind"])
					}
				}
			}
		}
	}
	return nil
}

func (c *Converter) additionalProperties(definitions map[string]interface{}) map[string]interface{} {
	for _, def := range definitions {
		if defMap, ok := def.(map[string]interface{}); ok {
			defMap["additionalProperties"] = false
		}
	}
	return definitions
}

func (c *Converter) appendNoDuplicates(prop interface{}, key string, value interface{}) {
	if propMap, ok := prop.(map[string]interface{}); ok {
		if existing, ok := propMap[key]; ok {
			if enum, ok := existing.([]interface{}); ok {
				// Check if value already exists
				for _, v := range enum {
					if v == value {
						return
					}
				}
				propMap[key] = append(enum, value)
			}
		} else {
			propMap[key] = []interface{}{value}
		}
	}
}

func (c *Converter) processOpenAPI(schema *OpenAPISchema) error {
	// Write shared schemas
	if err := c.writeSharedDefinitions(schema.Components["schemas"]); err != nil {
		return fmt.Errorf("failed to write shared schemas: %w", err)
	}

	// Process Kubernetes extensions if enabled
	if c.Kubernetes {
		if err := c.processKubernetesExtensions(schema.Components["schemas"]); err != nil {
			return fmt.Errorf("failed to process kubernetes extensions: %w", err)
		}
	}

	// Generate individual schemas
	var types []string
	for title, definition := range schema.Components["schemas"] {
		kind := strings.ToLower(strings.Split(title, ".")[len(strings.Split(title, "."))-1])
		fullName := kind

		if c.Kubernetes && c.Expanded {
			parts := strings.Split(title, ".")
			if len(parts) >= 3 {
				group := parts[len(parts)-3]
				apiVersion := parts[len(parts)-2]
				if group == "core" || group == "api" {
					fullName = fmt.Sprintf("%s-%s", kind, apiVersion)
				} else {
					fullName = fmt.Sprintf("%s-%s-%s", kind, group, apiVersion)
				}
			}
		}

		// Convert definition to schema
		schemaDef := definition.(map[string]interface{})
		schemaDef["$schema"] = "http://json-schema.org/schema#"
		schemaDef["type"] = "object"

		if c.Strict {
			schemaDef["additionalProperties"] = false
		}

		// Write individual schema file
		if err := c.writeSchemaFile(fullName, schemaDef); err != nil {
			return fmt.Errorf("failed to write schema for %s: %w", fullName, err)
		}

		types = append(types, title)
	}

	// Write all.json with references to all types
	allRefs := make([]map[string]string, 0, len(types))
	for _, title := range types {
		allRefs = append(allRefs, map[string]string{
			"$ref": title + ".json",
		})
	}

	allSchema := map[string]interface{}{
		"oneOf": allRefs,
	}

	if err := c.writeSchemaFile("all", allSchema); err != nil {
		return fmt.Errorf("failed to write all.json: %w", err)
	}

	return nil
}

func (c *Converter) writeSchemaFile(name string, content interface{}) error {
	filePath := filepath.Join(c.OutputDir, name+".json")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create schema file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(content); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	return nil
}
