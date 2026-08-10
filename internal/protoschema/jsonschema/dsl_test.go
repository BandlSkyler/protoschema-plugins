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
	"testing"

	customv1 "github.com/BandlSkyler/protoschema-plugins/gen/proto/buf/protoschema/custom/v1"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/structpb"
)

// buildStrategyDescriptor builds a descriptor for:
//
//	message WinLimitStrategy {
//	  message StrategyBase { bool enabled = 1; }
//	  message WinLimitRule  { string name = 1; }
//	  StrategyBase base = 1;
//	  repeated WinLimitRule rules = 2;
//	}
//
// with the given CustomOptions on the message options.
func buildStrategyDescriptor(t *testing.T, co *customv1.CustomOptions) protoreflect.MessageDescriptor {
	t.Helper()
	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("demo.proto"),
		Package: proto.String("demo"),
		Syntax:  proto.String("proto3"),
	}
	base := &descriptorpb.DescriptorProto{
		Name: proto.String("StrategyBase"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{Name: proto.String("enabled"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_BOOL.Enum(), JsonName: proto.String("enabled")},
		},
	}
	rule := &descriptorpb.DescriptorProto{
		Name: proto.String("WinLimitRule"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{Name: proto.String("name"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("name")},
		},
	}
	msgOpts := &descriptorpb.MessageOptions{}
	if co != nil {
		proto.SetExtension(msgOpts, customv1.E_MessageOptions, co)
	}
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("WinLimitStrategy"),
		NestedType: []*descriptorpb.DescriptorProto{
			base,
			rule,
		},
		Field: []*descriptorpb.FieldDescriptorProto{
			{Name: proto.String("base"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".demo.WinLimitStrategy.StrategyBase"), JsonName: proto.String("base")},
			{Name: proto.String("rules"), Number: proto.Int32(2), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".demo.WinLimitStrategy.WinLimitRule"), Label: descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(), JsonName: proto.String("rules")},
		},
		Options: msgOpts,
	}
	fd.MessageType = []*descriptorpb.DescriptorProto{msg}

	files, err := protodesc.NewFiles(&descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{fd}})
	require.NoError(t, err)
	var desc protoreflect.MessageDescriptor
	files.RangeFiles(func(f protoreflect.FileDescriptor) bool {
		desc = f.Messages().ByName("WinLimitStrategy")
		return desc == nil
	})
	require.NotNil(t, desc)
	return desc
}

// generatePlainJSON generates the schema and converts it to a plain map via
// JSON round-trip so custom types (e.g. orderedProperties) are normalized.
func generatePlainJSON(t *testing.T, desc protoreflect.MessageDescriptor) map[string]any {
	t.Helper()
	g := NewGenerator()
	require.NoError(t, g.Add(desc))
	schemas := g.Generate()
	data, err := json.Marshal(schemas[desc.FullName()])
	require.NoError(t, err)
	var root map[string]any
	require.NoError(t, json.Unmarshal(data, &root))
	return root
}

func TestConditionalCompilesIfThenElse(t *testing.T) {
	t.Parallel()
	co := &customv1.CustomOptions{
		Conditional: &customv1.Conditional{
			If: []*customv1.Condition{
				{Field: proto.String("base.enabled"), Equals: &customv1.Condition_EqualsBool{EqualsBool: true}},
			},
			Then: &customv1.Constraints{
				Required: []string{"rules"},
				FieldConstraints: []*customv1.FieldConstraint{
					{Field: proto.String("rules"), MinItems: proto.Uint64(1)},
				},
			},
			Else: &customv1.Constraints{
				Required: []string{"base"},
			},
		},
	}
	schema := generatePlainJSON(t, buildStrategyDescriptor(t, co))

	require.Equal(t, roundTripJSON(t, map[string]any{
		"required": []any{"base"},
		"properties": map[string]any{
			"base": map[string]any{
				"required": []any{"enabled"},
				"properties": map[string]any{
					"enabled": map[string]any{"const": true},
				},
			},
		},
	}), schema["if"])
	require.Equal(t, roundTripJSON(t, map[string]any{
		"required": []any{"rules"},
		"properties": map[string]any{
			"rules": map[string]any{"minItems": 1},
		},
	}), schema["then"])
	require.Equal(t, roundTripJSON(t, map[string]any{"required": []any{"base"}}), schema["else"])
}

func TestAllOfCompiles(t *testing.T) {
	t.Parallel()
	co := &customv1.CustomOptions{
		AllOf: []*customv1.Constraints{
			{Required: []string{"base"}},
			{
				FieldConstraints: []*customv1.FieldConstraint{
					{Field: proto.String("name"), Const: stringValue(t, "x")},
					{Field: proto.String("rules"), MinItems: proto.Uint64(2)},
				},
			},
			{Properties: &structpb.Struct{Fields: map[string]*structpb.Value{
				"x-meta": stringValue(t, "v"),
			}}},
		},
	}
	schema := generatePlainJSON(t, buildStrategyDescriptor(t, co))

	require.Equal(t, roundTripJSON(t, map[string]any{
		"allOf": []any{
			map[string]any{"required": []any{"base"}},
			map[string]any{"properties": map[string]any{
				"name":  map[string]any{"const": "x"},
				"rules": map[string]any{"minItems": 2},
			}},
			map[string]any{"x-meta": "v"},
		},
	})["allOf"], schema["allOf"])
}

func TestDisplayIfCompilesVendorKeyword(t *testing.T) {
	t.Parallel()
	co := &customv1.CustomOptions{
		DisplayIf: &customv1.Condition{
			Field:  proto.String("base.enabled"),
			Equals: &customv1.Condition_EqualsBool{EqualsBool: true},
		},
	}
	schema := generatePlainJSON(t, buildStrategyDescriptor(t, co))
	require.Equal(t, map[string]any{"field": "base.enabled", "equals": true}, schema["x-display-if"])
}

func TestConditionalSemantics(t *testing.T) {
	t.Parallel()
	co := &customv1.CustomOptions{
		Conditional: &customv1.Conditional{
			If: []*customv1.Condition{
				{Field: proto.String("base.enabled"), Equals: &customv1.Condition_EqualsBool{EqualsBool: true}},
			},
			Then: &customv1.Constraints{
				Required: []string{"rules"},
			},
		},
	}
	schema := generatePlainJSON(t, buildStrategyDescriptor(t, co))
	// Isolate the if/then logic by dropping the permissive $refs.
	for _, key := range []string{"base", "rules"} {
		props := schema["properties"].(map[string]any)
		field := props[key].(map[string]any)
		if items, ok := field["items"]; ok {
			delete(items.(map[string]any), "$ref")
		} else {
			delete(field, "$ref")
		}
	}

	c := jsonschema.NewCompiler()
	require.NoError(t, c.AddResource("demo://strategy.json", schema))
	compiled, err := c.Compile("demo://strategy.json")
	require.NoError(t, err)

	cases := []struct {
		name string
		data any
		ok   bool
	}{
		{"enabled=true, no rules -> invalid", map[string]any{"base": map[string]any{"enabled": true}}, false},
		{"enabled=true, one rule -> valid", map[string]any{"base": map[string]any{"enabled": true}, "rules": []any{map[string]any{"name": "a"}}}, true},
		{"enabled=false, no rules -> valid", map[string]any{"base": map[string]any{"enabled": false}}, true},
		{"missing base, no rules -> valid", map[string]any{}, true},
	}
	for _, tc := range cases {
		err := compiled.Validate(tc.data)
		if tc.ok {
			require.NoErrorf(t, err, "%s: expected valid, got %v", tc.name, err)
		} else {
			require.Errorf(t, err, "%s: expected invalid", tc.name)
		}
	}
}

func stringValue(t *testing.T, s string) *structpb.Value {
	t.Helper()
	v, err := structpb.NewValue(s)
	require.NoError(t, err)
	return v
}

// roundTripJSON normalizes a map through JSON so that all maps and slices use
// the same concrete types as a freshly unmarshalled schema.
func roundTripJSON(t *testing.T, v map[string]any) map[string]any {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))
	return out
}
