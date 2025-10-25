// Copyright 2025.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ssh_config_file

import (
	"reflect"
	"testing"

	"github.com/kevinburke/ssh_config"
)

func TestConvertCLIForwardToConfigFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic local forward",
			input:    "8080:localhost:80",
			expected: "8080 localhost:80",
		},
		{
			name:     "local forward with bind address",
			input:    "127.0.0.1:8080:localhost:80",
			expected: "127.0.0.1:8080 localhost:80",
		},
		{
			name:     "local forward with wildcard bind",
			input:    "*:8080:localhost:80",
			expected: "*:8080 localhost:80",
		},
		{
			name:     "remote forward",
			input:    "8080:localhost:3000",
			expected: "8080 localhost:3000",
		},
		{
			name:     "remote forward with bind address",
			input:    "0.0.0.0:80:localhost:8080",
			expected: "0.0.0.0:80 localhost:8080",
		},
		{
			name:     "forward with IPv6 address",
			input:    "8080:[2001:db8::1]:80",
			expected: "8080 [2001:db8::1]:80",
		},
		{
			name:     "forward with domain",
			input:    "3306:db.example.com:3306",
			expected: "3306 db.example.com:3306",
		},
		{
			name:     "invalid format - only one colon",
			input:    "8080:localhost",
			expected: "8080:localhost", // returned as-is
		},
		{
			name:     "invalid format - no colons",
			input:    "8080",
			expected: "8080", // returned as-is
		},
	}

	r := &Repository{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.convertCLIForwardToConfigFormat(tt.input)
			if result != tt.expected {
				t.Errorf("convertCLIForwardToConfigFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertConfigForwardToCLIFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic local forward",
			input:    "8080 localhost:80",
			expected: "8080:localhost:80",
		},
		{
			name:     "local forward with bind address",
			input:    "127.0.0.1:8080 localhost:80",
			expected: "127.0.0.1:8080:localhost:80",
		},
		{
			name:     "local forward with wildcard bind",
			input:    "*:8080 localhost:80",
			expected: "*:8080:localhost:80",
		},
		{
			name:     "remote forward",
			input:    "8080 localhost:3000",
			expected: "8080:localhost:3000",
		},
		{
			name:     "remote forward with bind address",
			input:    "0.0.0.0:80 localhost:8080",
			expected: "0.0.0.0:80:localhost:8080",
		},
		{
			name:     "forward with IPv6 address",
			input:    "8080 [2001:db8::1]:80",
			expected: "8080:[2001:db8::1]:80",
		},
		{
			name:     "forward with domain",
			input:    "3306 db.example.com:3306",
			expected: "3306:db.example.com:3306",
		},
		{
			name:     "already in CLI format",
			input:    "8080:localhost:80",
			expected: "8080:localhost:80", // returned as-is
		},
		{
			name:     "no space separator",
			input:    "8080",
			expected: "8080", // returned as-is
		},
	}

	r := &Repository{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.convertConfigForwardToCLIFormat(tt.input)
			if result != tt.expected {
				t.Errorf("convertConfigForwardToCLIFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractTagsFromHost(t *testing.T) {
	repo := &Repository{}
	host := &ssh_config.Host{
		Nodes: []ssh_config.Node{
			&ssh_config.Empty{Comment: "tag: foo, bar"},
		},
	}

	got := repo.extractTagsFromHost(host)
	want := []string{"foo", "bar"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractTagsFromHost() = %v, want %v", got, want)
	}

	host.Nodes = []ssh_config.Node{
		&ssh_config.KV{Key: "User", Value: "git", Comment: "tag: deploy"},
	}
	got = repo.extractTagsFromHost(host)
	want = []string{"deploy"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractTagsFromHost() inline = %v, want %v", got, want)
	}
}

func TestSetTagsOnHost(t *testing.T) {
	repo := &Repository{}
	host := &ssh_config.Host{
		Nodes: []ssh_config.Node{
			&ssh_config.KV{Key: "HostName", Value: "example.com"},
		},
	}

	repo.setTagsOnHost(host, []string{" foo ", "bar", "Foo"})
	if len(host.Nodes) != 2 {
		t.Fatalf("expected 2 nodes after inserting tag comment, got %d", len(host.Nodes))
	}
	commentNode, ok := host.Nodes[0].(*ssh_config.Empty)
	if !ok {
		t.Fatalf("expected first node to be comment, got %T", host.Nodes[0])
	}
	if commentNode.Comment != "tag: foo, bar" {
		t.Fatalf("unexpected comment: %q", commentNode.Comment)
	}

	// Removing tags should drop the comment node
	repo.setTagsOnHost(host, nil)
	if len(host.Nodes) != 1 {
		t.Fatalf("expected tag comment removed, remaining nodes %d", len(host.Nodes))
	}
}
