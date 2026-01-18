package logseq_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clstb/yalms/pkg/logseq"
)

func TestModels_UnmarshalError(t *testing.T) {
	var p logseq.Page
	err := json.Unmarshal([]byte(`{"id": "not-an-int"}`), &p)
	if err == nil {
		t.Error("Expected error unmarshaling invalid JSON into Page")
	}
}

func TestClient_CallError(t *testing.T) {
	// Test with invalid URL to trigger request failure
	client := logseq.NewClient("http://invalid-url-12345", "token", nil)
	_, err := client.GetGraph()
	if err == nil {
		t.Error("Expected error from Call with invalid URL")
	}
}

func TestModels_UnmarshalJSON_ExtraFields(t *testing.T) {
	var p logseq.Page
	raw := `{"uuid": "u1", "name": "test/p1", "extra": "val", "journal?": true}`
	err := json.Unmarshal([]byte(raw), &p)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if p.UUID != "u1" {
		t.Errorf("Expected UUID u1, got %s", p.UUID)
	}
	if p.Properties["extra"] != "val" {
		t.Errorf("Expected extra property val, got %v", p.Properties["extra"])
	}
}

func TestClient_GetNamespacePages_Parsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Test different result formats
		if r.URL.Path == "/api" {
			w.Write([]byte(`[[{"uuid": "u1", "name": "test/p1"}], [{"uuid": "u2", "name": "test/p2"}]]`))
		}
	}))
	defer server.Close()

	client := logseq.NewClient(server.URL, "token", nil)
	pages, err := client.GetNamespacePages("test")
	if err != nil {
		t.Fatalf("GetNamespacePages failed: %v", err)
	}
	if len(pages) != 2 {
		t.Errorf("Expected 2 pages, got %d", len(pages))
	}

	// Test directly returning list of maps
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"uuid": "u3", "name": "test/p3"}]`))
	}))
	defer server2.Close()
	client2 := logseq.NewClient(server2.URL, "token", nil)
	pages2, err := client2.GetNamespacePages("test")
	if err != nil || len(pages2) != 1 {
		t.Errorf("Direct map list failed: %v, count: %d", err, len(pages2))
	}
}

func TestClient_AddTag_ComplexFallbacks(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		// 1. GetBlock(uuid) -> null
		// 2. GetPage(uuid) -> page object
		// 3. GetBlock(page.uuid) -> null
		// 4. getPageBlocksTree(page.uuid) -> [block]
		// 5. UpdateBlock(block.uuid, content+tag) -> block
		switch callCount {
		case 1: // GetBlock(uuid)
			w.Write([]byte(`null`))
		case 2: // GetPage(uuid)
			w.Write([]byte(`{"uuid": "u1", "name": "test/p1"}`))
		case 3: // GetBlock(page.uuid)
			w.Write([]byte(`null`))
		case 4: // getPageBlocksTree(page.uuid)
			w.Write([]byte(`[{"uuid": "b1", "content": "initial"}]`))
		case 5: // UpdateBlock
			w.Write([]byte(`{"uuid": "b1", "content": "initial #tag"}`))
		}
	}))
	defer server.Close()

	client := logseq.NewClient(server.URL, "token", nil)
	err := client.AddTag("u1", "tag")
	if err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if callCount != 5 {
		t.Errorf("Expected 5 calls for complex fallback, got %d", callCount)
	}
}

func TestClient_GetGraph_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"name": "test", "path": "/path"}`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	graph, err := client.GetGraph()
	if err != nil || graph.Name != "test" {
		t.Errorf("GetGraph failed: %v, graph: %v", err, graph)
	}
}

func TestClient_RenamePage_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	err := client.RenamePage("u1", "new-name")
	if err != nil {
		t.Errorf("RenamePage failed: %v", err)
	}
}

func TestClient_CreatePage_Success(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 { // GetPage check
			w.Write([]byte(`null`))
		} else { // CreatePage
			w.Write([]byte(`{"uuid": "u1", "name": "test/p1"}`))
		}
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	page, err := client.CreatePage("p1", nil, nil)
	if err != nil || page.UUID != "u1" {
		t.Errorf("CreatePage failed: %v, page: %v", err, page)
	}
}

func TestClient_UpdatePage_Success(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 { // upsertBlockProperty
			w.Write([]byte(`{}`))
		} else { // GetPage
			w.Write([]byte(`{"uuid": "u1", "name": "test/p1"}`))
		}
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	page, err := client.UpdatePage("u1", map[string]any{"k1": "v1"})
	if err != nil || page.UUID != "u1" {
		t.Errorf("UpdatePage failed: %v, page: %v", err, page)
	}
}

func TestClient_DeletePage_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	err := client.DeletePage("p1")
	if err != nil {
		t.Errorf("DeletePage failed: %v", err)
	}
}

func TestClient_RemoveProperty_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	err := client.RemoveProperty("u1", "k1")
	if err != nil {
		t.Errorf("RemoveProperty failed: %v", err)
	}
}

func TestClient_GetBlock_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"uuid": "b1", "content": "c1"}`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	block, err := client.GetBlock("b1")
	if err != nil || block.UUID != "b1" {
		t.Errorf("GetBlock failed: %v, block: %v", err, block)
	}
}

func TestClient_InsertBlock_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"uuid": "b1", "content": "c1"}`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	block, err := client.InsertBlock("p1", "c1", nil, nil)
	if err != nil || block.UUID != "b1" {
		t.Errorf("InsertBlock failed: %v, block: %v", err, block)
	}
}

func TestClient_InsertBatchBlock_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"uuid": "b1", "content": "c1"}]`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	blocks, err := client.InsertBatchBlock("p1", []logseq.BlockContent{{Content: "c1"}}, nil)
	if err != nil || len(blocks) != 1 {
		t.Errorf("InsertBatchBlock failed: %v, count: %d", err, len(blocks))
	}
}

func TestClient_DeleteBlock_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	err := client.DeleteBlock("b1")
	if err != nil {
		t.Errorf("DeleteBlock failed: %v", err)
	}
}

func TestClient_AppendBlockInPage_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"uuid": "b1", "content": "c1"}`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	block, err := client.AppendBlockInPage("p1", "c1", nil)
	if err != nil || block.UUID != "b1" {
		t.Errorf("AppendBlockInPage failed: %v, block: %v", err, block)
	}
}

func TestClient_RemoveTag_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"uuid": "b1", "content": "initial #tag"}`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	err := client.RemoveTag("b1", "tag")
	if err != nil {
		t.Errorf("RemoveTag failed: %v", err)
	}
}

func TestClient_EnsureLinkedPages_Namespaces(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		// 1. GetPage("A") -> null
		// 2. CreatePage("A") -> {uuid: "uA"}
		// 3. GetPage("A/B") -> null
		// 4. CreatePage("A/B") -> {uuid: "uB"}
		switch callCount {
		case 1, 3:
			w.Write([]byte(`null`))
		case 2:
			w.Write([]byte(`{"uuid": "uA", "name": "A"}`))
		case 4:
			w.Write([]byte(`{"uuid": "uB", "name": "A/B"}`))
		}
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	content := "link to [[A/B]]"
	newContent := client.EnsureLinkedPages(content, nil)
	expected := "link to ((uB))"
	if newContent != expected {
		t.Errorf("Expected %s, got %s", expected, newContent)
	}
}

func TestClient_InsertBlock_WithProperties(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		// 1. insertBlock
		// 2. updateBlock (for properties)
		// 3. getBlock (refresh)
		w.Write([]byte(`{"uuid": "b1", "content": "c1"}`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	props := map[string]any{"key": "val"}
	_, err := client.InsertBlock("p1", "c1", props, nil)
	if err != nil {
		t.Errorf("InsertBlock with properties failed: %v", err)
	}
	if callCount < 2 {
		t.Errorf("Expected at least 2 calls for InsertBlock with properties, got %d", callCount)
	}
}

func TestClient_UpdateBlock_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"uuid": "b1", "content": "updated"}`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	block, err := client.UpdateBlock("b1", "updated", nil)
	if err != nil || block.Content != "updated" {
		t.Errorf("UpdateBlock failed: %v", err)
	}
}

func TestModels_EntityRef_UnmarshalJSON(t *testing.T) {
	var er logseq.EntityRef
	// Test ID
	err := json.Unmarshal([]byte(`123`), &er)
	if err != nil || er.ID != 123 {
		t.Errorf("EntityRef ID unmarshal failed: %v, ID: %d", err, er.ID)
	}
	// Test Object
	err = json.Unmarshal([]byte(`{"id": 456, "uuid": "u1"}`), &er)
	if err != nil || er.ID != 456 || er.UUID != "u1" {
		t.Errorf("EntityRef Object unmarshal failed: %v, er: %+v", err, er)
	}
}

func TestClient_GetDailyJournal_Success(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		// 1. getTodayJournalPage
		w.Write([]byte(`{"uuid": "u1", "name": "2026-01-18", "journal?": true}`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	pageAny, err := client.GetDailyJournal()
	if err != nil {
		t.Fatalf("GetDailyJournal failed: %v", err)
	}
	page, ok := pageAny.(*logseq.Page)
	if !ok || page.UUID != "u1" {
		t.Errorf("GetDailyJournal failed: got %v, want UUID u1", pageAny)
	}
}

func TestClient_GetDailyJournal_NoFallback(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 { // getTodayJournalPage
			w.Write([]byte(`null`))
		}
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	page, err := client.GetDailyJournal()
	if err != nil || page != nil {
		t.Errorf("GetDailyJournal should not fall back: err: %v, page: %v", err, page)
	}
}

func TestClient_ListPages_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"uuid": "u1", "name": "test/p1"}, {"uuid": "u2", "name": "test/p2"}]`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	pages, err := client.ListPages()
	if err != nil || len(pages) != 2 {
		t.Errorf("ListPages failed: %v, count: %d", err, len(pages))
	}
}

func TestClient_ListPages_Fallback(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 { // getAllPages
			w.Write([]byte(`null`))
		} else { // logseq.DB.q
			w.Write([]byte(`[[{"uuid": "u3", "name": "test/p3"}]]`))
		}
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	pages, err := client.ListPages()
	if err != nil || len(pages) != 1 {
		t.Errorf("ListPages fallback failed: %v, count: %d", err, len(pages))
	}
}

func TestClient_ListNamespaces_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[["Parent1"], ["Parent2"]]`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	ns, err := client.ListNamespaces()
	if err != nil {
		t.Fatalf("ListNamespaces failed: %v", err)
	}
	if len(ns) != 2 {
		t.Errorf("Expected 2 namespaces, got %v", ns)
	}
}

func TestClient_Call_ErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`error message`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	_, err := client.Call("method")
	if err == nil {
		t.Error("Expected error from Call with 500 status")
	}
}

func TestClient_AddTag_EmptyPage(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		// 1. getBlock(uuid) -> null
		// 2. getPage(uuid) -> page
		// 3. getBlock(page.uuid) -> null
		// 4. getPageBlocksTree(page.uuid) -> null (empty page)
		// 5. appendBlockInPage(page.uuid, "") -> block
		// 6. UpdateBlock(block.uuid, tag) -> block
		switch callCount {
		case 1: // GetBlock(uuid)
			w.Write([]byte(`null`))
		case 2: // GetPage(uuid)
			w.Write([]byte(`{"uuid": "u1", "name": "test/p1"}`))
		case 3: // GetBlock(page.uuid)
			w.Write([]byte(`null`))
		case 4: // getPageBlocksTree(page.uuid)
			w.Write([]byte(`null`))
		case 5: // appendBlockInPage
			w.Write([]byte(`{"uuid": "b1", "content": ""}`))
		case 6: // UpdateBlock
			w.Write([]byte(`{"uuid": "b1", "content": "#tag"}`))
		}
	}))
	defer server.Close()

	client := logseq.NewClient(server.URL, "token", nil)
	err := client.AddTag("u1", "tag")
	if err != nil {
		t.Fatalf("AddTag failed on empty page: %v", err)
	}
	if callCount != 6 {
		t.Errorf("Expected 6 calls, got %d", callCount)
	}
}

func TestClient_Call_BusinessError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "something went wrong"}`))
	}))
	defer ts.Close()
	client := logseq.NewClient(ts.URL, "token", nil)
	_, err := client.Call("method")
	if err == nil || err.Error() != "api error: something went wrong" {
		t.Errorf("Expected business error, got: %v", err)
	}
}
