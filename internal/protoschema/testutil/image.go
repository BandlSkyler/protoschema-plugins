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

// Package testutil provides helpers shared by the protoschema plugin tests.
package testutil

import (
	"encoding/json"
	"fmt"

	"github.com/bufbuild/protoplugin/protopluginutil"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// ImageJSONToCodeGeneratorRequest converts the JSON representation of a buf build
// image (e.g. the output of `buf build -o -#format=json`) into a
// CodeGeneratorRequest. It mirrors buf's
// ImageToCodeGeneratorRequest(image, "", nil, false, false), i.e. only non-import
// files are added to FileToGenerate. The isImport flag of each file is read from
// the bufExtension extension field, which the descriptor parse discards.
func ImageJSONToCodeGeneratorRequest(data []byte) (*pluginpb.CodeGeneratorRequest, error) {
	fdset := &descriptorpb.FileDescriptorSet{}
	if err := (&protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, fdset); err != nil {
		return nil, fmt.Errorf("failed to parse image JSON as a file descriptor set: %w", err)
	}
	importFlags, err := fileIsImportFlags(data)
	if err != nil {
		return nil, err
	}
	request := &pluginpb.CodeGeneratorRequest{
		ProtoFile:             make([]*descriptorpb.FileDescriptorProto, 0, len(fdset.GetFile())),
		SourceFileDescriptors: make([]*descriptorpb.FileDescriptorProto, 0),
	}
	for i, file := range fdset.GetFile() {
		if i >= len(importFlags) || !importFlags[i] {
			request.FileToGenerate = append(request.FileToGenerate, file.GetName())
			// Source-retention options for files to generate are provided in
			// SourceFileDescriptors, while the descriptor in ProtoFile has them stripped.
			request.SourceFileDescriptors = append(request.SourceFileDescriptors, file)
			stripped, err := protopluginutil.StripSourceRetentionOptions(file)
			if err != nil {
				return nil, fmt.Errorf("failed to strip source-retention options for file %q: %w", file.GetName(), err)
			}
			file = stripped
		}
		request.ProtoFile = append(request.ProtoFile, file)
	}
	return request, nil
}

// fileIsImportFlags returns, for each file in the image JSON, whether the file
// was an import (read from the bufExtension.isImport field).
func fileIsImportFlags(data []byte) ([]bool, error) {
	var image struct {
		File []struct {
			BufExtension *struct {
				IsImport bool `json:"isImport"`
			} `json:"bufExtension"`
		} `json:"file"`
	}
	if err := json.Unmarshal(data, &image); err != nil {
		return nil, fmt.Errorf("failed to parse image JSON: %w", err)
	}
	importFlags := make([]bool, len(image.File))
	for i, file := range image.File {
		if file.BufExtension != nil {
			importFlags[i] = file.BufExtension.IsImport
		}
	}
	return importFlags, nil
}
