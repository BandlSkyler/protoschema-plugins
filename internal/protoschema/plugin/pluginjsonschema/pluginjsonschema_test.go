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

package pluginjsonschema

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"slices"
	"testing"

	_ "github.com/BandlSkyler/protoschema-plugins/gen/proto/buf/protoschema/test/v1"
	"github.com/BandlSkyler/protoschema-plugins/internal/protoschema/golden"
	"github.com/BandlSkyler/protoschema-plugins/internal/protoschema/jsonschema"
	"github.com/BandlSkyler/protoschema-plugins/internal/protoschema/testutil"
	"github.com/bufbuild/protoplugin"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestJSONSchemaHandler(t *testing.T) {
	t.Parallel()

	goldenPath := filepath.FromSlash("../../../testdata/jsonschema")
	inputImage := filepath.FromSlash("../../../testdata/codegenrequest/input.json")

	by, err := os.ReadFile(inputImage)
	require.NoError(t, err)
	codeGeneratorRequest, err := testutil.ImageJSONToCodeGeneratorRequest(by)
	require.NoError(t, err)

	request, err := proto.Marshal(codeGeneratorRequest)
	require.NoError(t, err)
	stdin := bytes.NewReader(request)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	err = protoplugin.Run(
		t.Context(),
		protoplugin.Env{
			Stdin:  stdin,
			Stdout: stdout,
			Stderr: stderr,
		},
		protoplugin.HandlerFunc(Handle),
	)
	require.NoError(t, err)
	require.Empty(t, stderr.String())

	response := new(pluginpb.CodeGeneratorResponse)
	err = proto.Unmarshal(stdout.Bytes(), response)
	require.NoError(t, err)

	wantFiles := make([]string, 0, len(response.GetFile()))
	for _, file := range response.GetFile() {
		wantFiles = append(wantFiles, file.GetName())
	}
	slices.Sort(wantFiles)
	require.Equal(t, wantFiles, gatherGoldenFiles(t, goldenPath))

	for _, file := range response.GetFile() {
		filename := path.Join(goldenPath, file.GetName())
		want, err := os.ReadFile(filename)
		require.NoError(t, err)
		require.Equal(t, string(want), file.GetContent())
	}
}

func gatherGoldenFiles(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	var files []string
	for _, entry := range entries {
		if path.Ext(entry.Name()) == ".json" {
			files = append(files, entry.Name())
		}
	}
	slices.Sort(files)
	return files
}

func TestParseOptionsSchemaVersion(t *testing.T) {
	t.Parallel()

	// Unknown version is rejected.
	_, err := parseOptions("schema_version=foo")
	require.Error(t, err)

	// All accepted spellings of both versions parse successfully.
	for _, version := range []string{"draft-07", "07", "draft7", "2020-12", "2020"} {
		opts, err := parseOptions("schema_version=" + version + ",target=proto")
		require.NoError(t, err, version)
		require.Len(t, opts, 1)
	}

	// draft-07 produces draft-07 schemas.
	opts, err := parseOptions("schema_version=draft-07,target=proto")
	require.NoError(t, err)
	gen := jsonschema.NewGenerator(opts[0]...)
	product := findProductDescriptor(t)
	require.NoError(t, gen.Add(product))
	require.Equal(t, "http://json-schema.org/draft-07/schema#", gen.Generate()[product.FullName()]["$schema"])

	// Default (no schema_version) keeps draft 2020-12.
	opts, err = parseOptions("target=proto")
	require.NoError(t, err)
	gen = jsonschema.NewGenerator(opts[0]...)
	require.NoError(t, gen.Add(product))
	require.Equal(t, "https://json-schema.org/draft/2020-12/schema", gen.Generate()[product.FullName()]["$schema"])
}

func TestParseOptionsNonRequiredDefault(t *testing.T) {
	t.Parallel()

	// Unknown values are rejected.
	_, err := parseOptions("non_required_default=foo")
	require.Error(t, err)

	// false is a no-op: strict mode still requires implicit-default fields.
	opts, err := parseOptions("non_required_default=false,target=proto-strict")
	require.NoError(t, err)
	require.Len(t, opts, 1)
	gen := jsonschema.NewGenerator(opts[0]...)
	product := findProductDescriptor(t)
	require.NoError(t, gen.Add(product))
	require.Equal(t, []string{"product_id", "product_name", "price", "location"}, gen.Generate()[product.FullName()]["required"])

	// true makes only explicitly required fields required, even in strict mode.
	opts, err = parseOptions("non_required_default=true,target=proto-strict")
	require.NoError(t, err)
	require.Len(t, opts, 1)
	gen = jsonschema.NewGenerator(opts[0]...)
	require.NoError(t, gen.Add(product))
	require.Equal(t, []string{"product_id", "product_name", "location"}, gen.Generate()[product.FullName()]["required"])
}

func findProductDescriptor(t *testing.T) protoreflect.MessageDescriptor {
	t.Helper()
	testDescs, err := golden.GetTestDescriptors("../../../testdata")
	require.NoError(t, err)
	for _, desc := range testDescs {
		if string(desc.FullName()) == "buf.protoschema.test.v1.Product" {
			return desc
		}
	}
	t.Fatal("test descriptor buf.protoschema.test.v1.Product not found")
	return nil
}
