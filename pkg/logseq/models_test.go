package logseq_test

import (
	"encoding/json"
	"testing"

	"github.com/clstb/yalms/pkg/logseq"
)

func TestModels(t *testing.T) {
	// 1. Test BlockContent Marshaling (Recursive)
	content := logseq.BlockContent{
		Content: "Parent",
		Children: []logseq.BlockContent{
			{Content: "Child 1"},
			{Content: "Child 2"},
		},
		Properties: map[string]any{"key": "value"},
	}
	
	bytes, err := json.Marshal(content)
	if err != nil {
		t.Fatalf("Failed to marshal BlockContent: %v", err)
	}
	
	// Verify JSON structure somewhat
	jsonStr := string(bytes)
	if jsonStr == "" {
		t.Fatal("Marshaled JSON is empty")
	}
	
	// 2. Test UnmarshalJSON for Block
	// Case 1: Simple block
	simpleJSON := `{"uuid": "123", "content": "hello"}`
	var b1 logseq.Block
	if err := json.Unmarshal([]byte(simpleJSON), &b1); err != nil {
		t.Fatalf("Failed to unmarshal simple block: %v", err)
	}
	if b1.UUID != "123" || b1.Content != "hello" {
		t.Errorf("Simple block mismatch: %+v", b1)
	}
	
	// Case 2: Block with Children as []interface{} (from API often)
	// Children might be ["uuid", "uuid"] or [{...}, {...}]
	childrenJSON := `{"uuid": "456", "content": "parent", "children": [["uuid:789"], ["uuid:999"]]}`
	var b2 logseq.Block
	if err := json.Unmarshal([]byte(childrenJSON), &b2); err != nil {
		t.Fatalf("Failed to unmarshal block with children: %v", err)
	}
	// Our custom UnmarshalJSON might handle children differently depending on implementation
	// Let's check what it does. It might ignore children if they are not full objects?
	
	// Case 3: Properties parsing
	// Logseq properties often come as map in properties field OR embedded in content?
	// The struct has `Properties map[string]interface{}`.
	// If the JSON has `properties: {...}`, it should work.
	propsJSON := `{"uuid": "p1", "properties": {"status": "done"}}`
	var b3 logseq.Block
	if err := json.Unmarshal([]byte(propsJSON), &b3); err != nil {
		t.Fatalf("Failed to unmarshal block properties: %v", err)
	}
	if b3.Properties["status"] != "done" {
		t.Errorf("Properties mismatch: %+v", b3.Properties)
	}
	
	// Case 4: Page properties
	// Pages also have properties
	pageJSON := `{"name": "test", "properties": {"tags": ["a", "b"]}}`
	var p1 logseq.Page
	if err := json.Unmarshal([]byte(pageJSON), &p1); err != nil {
		t.Fatalf("Failed to unmarshal page: %v", err)
	}
	// Page struct might not have custom Unmarshal?
}
