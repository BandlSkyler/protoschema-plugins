// Copyright 2024-2026 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package jsonschema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/BandlSkyler/protoschema-plugins/gen/proto/buf/protoschema/test/v1"
	"github.com/BandlSkyler/protoschema-plugins/internal/protoschema/golden"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
)

func TestJSONSchemaGolden(t *testing.T) {
	t.Parallel()
	dirPath := filepath.FromSlash("../../testdata/jsonschema")
	testDescs, err := golden.GetTestDescriptors("../../testdata")
	require.NoError(t, err)
	generator := NewGenerator()
	for _, testDesc := range testDescs {
		err = generator.Add(testDesc)
		require.NoError(t, err)
	}

	schemas := generator.Generate()
	require.NoError(t, err)
	for _, jsonSchema := range schemas {
		// Serialize the JSON
		data, err := json.MarshalIndent(jsonSchema, "", "  ")
		require.NoError(t, err)

		identifier, ok := jsonSchema["$id"].(string)
		require.True(t, ok)
		require.NotEmpty(t, identifier)

		filePath := filepath.Join(dirPath, identifier)
		err = golden.CheckGolden(filePath, string(data)+"\n")
		require.NoError(t, err)
	}
}

func TestTitle(t *testing.T) {
	t.Parallel()
	require.Equal(t, "Foo", nameToTitle("Foo"))
	require.Equal(t, "Foo Bar", nameToTitle("FooBar"))
	require.Equal(t, "foo Bar", nameToTitle("fooBar"))
	require.Equal(t, "Foo Bar Baz", nameToTitle("FooBarBaz"))
	require.Equal(t, "FOO Bar", nameToTitle("FOOBar"))
	require.Equal(t, "U Int64 Value", nameToTitle("UInt64Value"))
	require.Equal(t, "Uint64 Value", nameToTitle("Uint64Value"))
	require.Equal(t, "FOO", nameToTitle("FOO"))
}

func TestConstraints(t *testing.T) {
	t.Parallel()
	schemaPath := filepath.FromSlash("../../testdata/jsonschema/buf.protoschema.test.v1.ConstraintTests.schema.json")
	bundledSchemaPath := filepath.FromSlash("../../testdata/jsonschema/buf.protoschema.test.v1.ConstraintTests.schema.bundle.json")
	testPath := filepath.FromSlash("../../testdata/jsonschema-doc/test.ConstraintTests.yaml")
	expectedPath := filepath.FromSlash("../../testdata/jsonschema-doc/test.ConstraintTests.txt")
	expectedBundledPath := filepath.FromSlash("../../testdata/jsonschema-doc/test.ConstraintTests.bundle.txt")
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(schemaPath)
	require.NoError(t, err)
	bundledSchema, err := compiler.Compile(bundledSchemaPath)
	require.NoError(t, err)

	yamlData, err := os.ReadFile(testPath)
	require.NoError(t, err)
	var jsonData map[string]any
	err = yaml.Unmarshal(yamlData, &jsonData)
	require.NoError(t, err)

	assertValidation(t, schema, jsonData, expectedPath)
	assertValidation(t, bundledSchema, jsonData, expectedBundledPath)
}

func assertValidation(t *testing.T, schema *jsonschema.Schema, jsonData map[string]any, expectedPath string) {
	t.Helper()
	expectedData, err := os.ReadFile(expectedPath)
	require.NoError(t, err)
	err = schema.Validate(jsonData)
	require.Error(t, err)
	errStr := err.Error()
	// Remove the first line of the error message, which contains the path to the schema file.
	if pos := strings.Index(errStr, "\n"); pos != -1 {
		errStr = errStr[pos+1:]
	}
	expectedStr := string(expectedData)
	expectedStr = strings.TrimSpace(expectedStr)
	require.Equal(t, expectedStr, errStr, errStr)
}

func TestSchemaDraft07(t *testing.T) {
	t.Parallel()

	const (
		draft07SchemaURL = "http://json-schema.org/draft-07/schema#"
		productName      = "buf.protoschema.test.v1.Product"
		locationName     = "buf.protoschema.test.v1.Product.Location"
	)
	productDesc := findTestDescriptor(t, productName)

	// Bundled draft-07 schema uses the `definitions` container and the
	// `#/definitions/` reference prefix.
	bundleGen := NewGenerator(WithSchemaDraft(SchemaDraft07), WithBundle())
	require.NoError(t, bundleGen.Add(productDesc))
	bundle := bundleGen.Generate()[productDesc.FullName()]
	require.Equal(t, draft07SchemaURL, bundle["$schema"])
	_, hasDefs := bundle["$defs"]
	require.False(t, hasDefs)
	require.Equal(t, "#/definitions/"+productName+".schema.json", bundle["$ref"])
	definitions, ok := bundle["definitions"].(map[string]any)
	require.True(t, ok)
	locationSchema, ok := definitions[locationName+".schema.json"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, draft07SchemaURL, locationSchema["$schema"])

	// The draft-07 bundle compiles under a draft-07 validator and validates
	// instances. Round-trip through JSON so the document uses standard JSON value
	// types (e.g. []any) expected by the compiler.
	data, err := json.Marshal(bundle)
	require.NoError(t, err)
	var stdDoc any
	require.NoError(t, json.Unmarshal(data, &stdDoc))
	compiler := jsonschema.NewCompiler()
	require.NoError(t, compiler.AddResource("mem://draft07.json", stdDoc))
	compiled, err := compiler.Compile("mem://draft07.json")
	require.NoError(t, err)
	validProduct := map[string]any{
		"product_id":   int64(1),
		"product_name": "widget",
		"location":     map[string]any{"lat": 12.0, "long": 34.0},
	}
	require.NoError(t, compiled.Validate(validProduct))
	// Missing required fields fails validation.
	require.Error(t, compiled.Validate(map[string]any{"product_name": "widget"}))

	// Non-bundled draft-07 schema carries the draft-07 $schema.
	singleGen := NewGenerator(WithSchemaDraft(SchemaDraft07))
	require.NoError(t, singleGen.Add(productDesc))
	require.Equal(t, draft07SchemaURL, singleGen.Generate()[productDesc.FullName()]["$schema"])

	// Regression: the default generator still produces draft 2020-12.
	defaultGen := NewGenerator(WithBundle())
	require.NoError(t, defaultGen.Add(productDesc))
	defaultBundle := defaultGen.Generate()[productDesc.FullName()]
	require.Equal(t, "https://json-schema.org/draft/2020-12/schema", defaultBundle["$schema"])
	_, hasDefsDefault := defaultBundle["$defs"]
	require.True(t, hasDefsDefault)
	_, hasDefinitionsDefault := defaultBundle["definitions"]
	require.False(t, hasDefinitionsDefault)
}

func TestNonRequiredByDefault(t *testing.T) {
	t.Parallel()

	const (
		productName  = "buf.protoschema.test.v1.Product"
		locationName = "buf.protoschema.test.v1.Product.Location"
	)
	productDesc := findTestDescriptor(t, productName)

	// With the option enabled, strict mode only requires fields explicitly
	// marked (buf.validate.field).required = true. Non-optional scalar fields
	// like price/lat/long are no longer required.
	optGen := NewGenerator(WithStrict(), WithNonRequiredByDefault())
	require.NoError(t, optGen.Add(productDesc))
	schemas := optGen.Generate()
	require.Equal(t, []string{"product_id", "product_name", "location"}, schemas[productDesc.FullName()]["required"])
	_, hasLocationRequired := schemas[locationName]["required"]
	require.False(t, hasLocationRequired, "non-optional scalar fields should not be required by default")

	// Without the option, strict mode forces implicit-default fields (price,
	// lat, long) to be required, as a regression control.
	strictGen := NewGenerator(WithStrict())
	require.NoError(t, strictGen.Add(productDesc))
	strictSchemas := strictGen.Generate()
	require.Equal(t, []string{"product_id", "product_name", "price", "location"}, strictSchemas[productDesc.FullName()]["required"])
	require.Equal(t, []string{"lat", "long"}, strictSchemas[locationName]["required"])
}

func findTestDescriptor(t *testing.T, fqn string) protoreflect.MessageDescriptor {
	t.Helper()
	testDescs, err := golden.GetTestDescriptors("../../testdata")
	require.NoError(t, err)
	for _, desc := range testDescs {
		if string(desc.FullName()) == fqn {
			return desc
		}
	}
	t.Fatalf("test descriptor %q not found", fqn)
	return nil
}
