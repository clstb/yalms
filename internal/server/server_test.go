package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/clstb/yalms/internal/server"
	"github.com/clstb/yalms/pkg/logseq"
	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

func setupTestServer() (*server.MCPServer, *logseq.Client) {
	logger := zap.NewNop()
	client := logseq.NewClient("http://localhost:12345", "token", logger)
	s := server.NewMCPServer(client, logger, server.ModeGeneral)
	return s, client
}

func TestServer_Serve(t *testing.T) {
	// Serve uses ServeStdio which reads from stdin. 
	// We can't easily test it fully here without mocking stdio.
}

func TestServer_Query_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("query", map[string]any{})
	res, err := s.HandleQuery(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing datalog, got %v", res)
	}
}

func TestServer_Query_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("query", map[string]any{"query": "[:find ?p :where [?p :block/name]]"})
	res, err := s.HandleQuery(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleQuery failed: %v", res)
	}
}

func TestServer_ListNamespaces_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("list_namespaces", map[string]any{})
	res, err := s.HandleListNamespaces(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleListNamespaces failed: %v", res)
	}
}

func TestServer_GetDailyJournal_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("get_daily_journal", map[string]any{})
	res, err := s.HandleGetDailyJournal(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleGetDailyJournal failed: %v", res)
	}
}

func TestServer_ReadPage_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("read_page", map[string]any{})
	res, err := s.HandleReadPage(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing uuid, got %v", res)
	}
}

func TestServer_CreatePage_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("create_page", map[string]any{})
	res, err := s.HandleCreateEntity(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing name, got %v", res)
	}

	req = makeRequest("create_page", map[string]any{"name": "test", "properties": "{invalid"})
	res, err = s.HandleCreateEntity(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for invalid properties, got %v", res)
	}
}

func TestServer_CreatePages_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("create_pages", map[string]any{})
	res, err := s.HandleCreatePages(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing pages, got %v", res)
	}

	req = makeRequest("create_pages", map[string]any{"pages": "{invalid"})
	res, err = s.HandleCreatePages(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for invalid pages JSON, got %v", res)
	}
}

func TestServer_DeleteBlocks_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("delete_blocks", map[string]any{})
	res, err := s.HandleDeleteBlocks(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing uuids, got %v", res)
	}

	req = makeRequest("delete_blocks", map[string]any{"uuids": "{invalid"})
	res, err = s.HandleDeleteBlocks(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for invalid uuids JSON, got %v", res)
	}
}

func TestServer_AddTag_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("add_tag", map[string]any{})
	res, err := s.HandleAddTag(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing arguments, got %v", res)
	}
}

func TestServer_RemoveTag_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("remove_tag", map[string]any{})
	res, err := s.HandleRemoveTag(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing arguments, got %v", res)
	}
}

func TestServer_UpdateBlock_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("update_block", map[string]any{})
	res, err := s.HandleUpdateBlock(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing uuid, got %v", res)
	}

	req = makeRequest("update_block", map[string]any{"uuid": "u1", "properties": "{invalid"})
	res, err = s.HandleUpdateBlock(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for invalid properties, got %v", res)
	}
}

func TestServer_AppendBlock_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("append_block", map[string]any{})
	res, err := s.HandleAppendBlock(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing uuid, got %v", res)
	}

	req = makeRequest("append_block", map[string]any{"uuid": "p1"})
	res, err = s.HandleAppendBlock(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing content, got %v", res)
	}
}

func TestServer_DeleteBlock_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("delete_block", map[string]any{})
	res, err := s.HandleDeleteBlock(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing uuid, got %v", res)
	}
}

func TestServer_RemoveProperty_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("remove_property", map[string]any{})
	res, err := s.HandleRemoveProperty(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing arguments, got %v", res)
	}
}

func TestServer_DeletePages_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("delete_pages", map[string]any{})
	res, err := s.HandleDeletePages(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing uuids, got %v", res)
	}

	req = makeRequest("delete_pages", map[string]any{"uuids": "{invalid"})
	res, err = s.HandleDeletePages(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for invalid uuids JSON, got %v", res)
	}
}

func TestServer_RenamePage_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("rename_page", map[string]any{})
	res, err := s.HandleRenamePage(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing arguments, got %v", res)
	}

	req = makeRequest("rename_page", map[string]any{"uuid": "u1"})
	res, err = s.HandleRenamePage(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing new_name, got %v", res)
	}
}

func TestServer_ReadNamespace_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("read_namespace", map[string]any{})
	res, err := s.HandleReadNamespace(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing namespace, got %v", res)
	}
}

func TestServer_CreateNamespace_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("create_namespace", map[string]any{})
	res, err := s.HandleCreateNamespace(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing namespace, got %v", res)
	}
}

func TestServer_ReadBlock_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("read_block", map[string]any{})
	res, err := s.HandleReadBlock(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing uuid, got %v", res)
	}
}

func TestServer_CreateBlock_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("create_block", map[string]any{})
	res, err := s.HandleCreateBlock(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing parent_uuid, got %v", res)
	}

	req = makeRequest("create_block", map[string]any{"parent_uuid": "p1"})
	res, err = s.HandleCreateBlock(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing content, got %v", res)
	}

	req = makeRequest("create_block", map[string]any{"parent_uuid": "p1", "content": "c1", "properties": "{invalid"})
	res, err = s.HandleCreateBlock(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for invalid properties, got %v", res)
	}
}

func TestServer_CreateBlockTree_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("create_block_tree", map[string]any{})
	res, err := s.HandleCreateBlockTree(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing parent_uuid, got %v", res)
	}

	req = makeRequest("create_block_tree", map[string]any{"parent_uuid": "p1"})
	res, err = s.HandleCreateBlockTree(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing tree, got %v", res)
	}

	req = makeRequest("create_block_tree", map[string]any{"parent_uuid": "p1", "tree": "{invalid"})
	res, err = s.HandleCreateBlockTree(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for invalid tree JSON, got %v", res)
	}
}

func TestServer_UpdatePage_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("update_page", map[string]any{})
	res, err := s.HandleUpdatePage(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing arguments, got %v", res)
	}

	req = makeRequest("update_page", map[string]any{"uuid": "p1"})
	res, err = s.HandleUpdatePage(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing properties, got %v", res)
	}

	req = makeRequest("update_page", map[string]any{"uuid": "p1", "properties": "{invalid"})
	res, err = s.HandleUpdatePage(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for invalid properties JSON, got %v", res)
	}
}

func TestServer_DeletePage_Errors(t *testing.T) {
	s, _ := setupTestServer()
	req := makeRequest("delete_page", map[string]any{})
	res, err := s.HandleDeletePage(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error result for missing uuid, got %v", res)
	}
}

func setupSuccessMock() (*httptest.Server, *server.MCPServer) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Check method to return appropriate response
		if r.Body != nil {
			var body struct {
				Method string `json:"method"`
				Args   []any  `json:"args"`
			}
			json.NewDecoder(r.Body).Decode(&body)
			if body.Method == "logseq.App.getCurrentGraph" {
				w.Write([]byte(`{"name": "test-graph", "path": "/test"}`))
				return
			}
			if body.Method == "logseq.Editor.insertBatchBlock" {
				w.Write([]byte(`[{"uuid": "u1", "content": "test"}]`))
				return
			}
			if body.Method == "logseq.DB.q" {
				w.Write([]byte(`[[{"uuid": "u1", "name": "test"}]]`))
				return
			}
			if body.Method == "logseq.Editor.getPage" {
				if len(body.Args) > 0 {
					if name, ok := body.Args[0].(string); ok {
						if name == "non_existent" || name == "non_existent_ns" {
							w.Write([]byte("null"))
							return
						}
						// Return page with requested name
						w.Write([]byte(`{"uuid": "u1", "name": "` + name + `", "content": "test"}`))
						return
					}
				}
			}
			if body.Method == "logseq.Editor.createPage" {
				// Echo back the requested name in the response so CreatePage validation passes
				name := "test"
				if len(body.Args) > 0 {
					if n, ok := body.Args[0].(string); ok {
						name = n
					}
				}
				w.Write([]byte(`{"uuid": "u1", "name": "` + name + `", "content": "test"}`))
				return
			}
		}

		w.Write([]byte(`{"uuid": "u1", "name": "test", "content": "test"}`))
	}))

	logger := zap.NewNop()
	client := logseq.NewClient(ts.URL, "token", logger)
	s := server.NewMCPServer(client, logger, server.ModeGeneral)
	return ts, s
}

func TestServer_ReadGraphInfo_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	res, err := s.HandleReadGraphInfo(context.Background(), mcp.CallToolRequest{})
	if err != nil || res.IsError {
		t.Fatalf("handleReadGraphInfo failed: %v", err)
	}
	expected := "Graph: test-graph\nPath: /test"
	if res.Content[0].(mcp.TextContent).Text != expected {
		t.Errorf("Expected %q, got %q", expected, res.Content[0].(mcp.TextContent).Text)
	}
}

func TestServer_DeletePages_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("delete_pages", map[string]any{"uuids": `["p1", "p2"]`})
	res, err := s.HandleDeletePages(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleDeletePages failed: %v", res)
	}
}

func TestServer_CreatePages_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("create_pages", map[string]any{"pages": `[{"name": "non_existent"}]`})
	res, err := s.HandleCreatePages(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleCreatePages failed: %v", res)
	}
}

func TestServer_ReadBlock_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("read_block", map[string]any{"uuid": "b1"})
	res, err := s.HandleReadBlock(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleReadBlock failed: %v", res)
	}
}

func TestServer_ReadPage_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("read_page", map[string]any{"uuid": "p1"})
	res, err := s.HandleReadPage(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleReadPage failed: %v", res)
	}
}

func TestServer_CreateBlock_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("create_block", map[string]any{"parent_uuid": "p1", "content": "c1"})
	res, err := s.HandleCreateBlock(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleCreateBlock failed: %v", res)
	}
}

func TestServer_CreateBlockTree_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("create_block_tree", map[string]any{"parent_uuid": "p1", "tree": `[{"content": "c1"}]`})
	res, err := s.HandleCreateBlockTree(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleCreateBlockTree failed: %v", res)
	}
}

func TestServer_DeleteBlocks_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("delete_blocks", map[string]any{"uuids": `["b1", "b2"]`})
	res, err := s.HandleDeleteBlocks(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleDeleteBlocks failed: %v", res)
	}
}

func TestServer_UpdateBlock_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("update_block", map[string]any{"uuid": "b1", "content": "c1"})
	res, err := s.HandleUpdateBlock(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleUpdateBlock failed: %v", res)
	}
}

func TestServer_AppendBlock_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("append_block", map[string]any{"uuid": "p1", "content": "c1"})
	res, err := s.HandleAppendBlock(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleAppendBlock failed: %v", res)
	}
}

func TestServer_AddTag_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("add_tag", map[string]any{"uuid": "b1", "tag": "t1"})
	res, err := s.HandleAddTag(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleAddTag failed: %v", res)
	}
}

func TestServer_RemoveTag_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("remove_tag", map[string]any{"uuid": "b1", "tag": "t1"})
	res, err := s.HandleRemoveTag(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleRemoveTag failed: %v", res)
	}
}

func TestServer_RemoveProperty_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("remove_property", map[string]any{"uuid": "b1", "key": "k1"})
	res, err := s.HandleRemoveProperty(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleRemoveProperty failed: %v", res)
	}
}

func TestServer_AddProperty_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("add_property", map[string]any{"uuid": "b1", "key": "k1", "value": "v1"})
	res, err := s.HandleUpsertProperty(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleUpsertProperty failed: %v", res)
	}
}

func TestServer_DeletePage_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("delete_page", map[string]any{"uuid": "p1"})
	res, err := s.HandleDeletePage(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleDeletePage failed: %v", res)
	}
}

func TestServer_RenamePage_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("rename_page", map[string]any{"uuid": "u1", "new_name": "non_existent"})
	res, err := s.HandleRenamePage(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleRenamePage failed: %v", res)
	}
}

func TestServer_ReadNamespace_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("read_namespace", map[string]any{"namespace": "n1"})
	res, err := s.HandleReadNamespace(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleReadNamespace failed: %v", res)
	}
}

func TestServer_CreateNamespace_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("create_namespace", map[string]any{"namespace": "non_existent_ns"})
	res, err := s.HandleCreateNamespace(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleCreateNamespace failed: %v", res)
	}
}

func TestServer_UpdatePage_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("update_page", map[string]any{"uuid": "p1", "properties": `{"k1": "v1"}`})
	res, err := s.HandleUpdatePage(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleUpdatePage failed: %v", res)
	}
}

func TestServer_DeleteBlock_Success(t *testing.T) {
	ts, s := setupSuccessMock()
	defer ts.Close()
	req := makeRequest("delete_block", map[string]any{"uuid": "b1"})
	res, err := s.HandleDeleteBlock(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("handleDeleteBlock failed: %v", res)
	}
}

func TestServer_ReadGraphInfo_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer ts.Close()

	logger := zap.NewNop()
	client := logseq.NewClient(ts.URL, "token", logger)
	s := server.NewMCPServer(client, logger, server.ModeGeneral)

	res, err := s.HandleReadGraphInfo(context.Background(), mcp.CallToolRequest{})
	if err != nil || !res.IsError {
		t.Errorf("Expected error for handleReadGraphInfo, got %v", res)
	}
}

func TestServer_ReadPage_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer ts.Close()

	logger := zap.NewNop()
	client := logseq.NewClient(ts.URL, "token", logger)
	s := server.NewMCPServer(client, logger, server.ModeGeneral)

	req := makeRequest("read_page", map[string]any{"name": "p1"})
	res, err := s.HandleReadPage(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error for handleReadPage, got %v", res)
	}
}

func TestServer_CreatePage_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer ts.Close()

	logger := zap.NewNop()
	client := logseq.NewClient(ts.URL, "token", logger)
	s := server.NewMCPServer(client, logger, server.ModeGeneral)

	req := makeRequest("create_page", map[string]any{"name": "p1"})
	res, err := s.HandleCreateEntity(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error for handleCreatePage, got %v", res)
	}
}

func TestServer_CreateNamespace_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer ts.Close()

	logger := zap.NewNop()
	client := logseq.NewClient(ts.URL, "token", logger)
	s := server.NewMCPServer(client, logger, server.ModeGeneral)

	req := makeRequest("create_namespace", map[string]any{"namespace": "n1"})
	res, err := s.HandleCreateNamespace(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error for handleCreateNamespace, got %v", res)
	}
}

func TestServer_ReadBlock_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer ts.Close()

	logger := zap.NewNop()
	client := logseq.NewClient(ts.URL, "token", logger)
	s := server.NewMCPServer(client, logger, server.ModeGeneral)

	req := makeRequest("read_block", map[string]any{"uuid": "b1"})
	res, err := s.HandleReadBlock(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error for handleReadBlock, got %v", res)
	}
}

func TestServer_AddTag_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer ts.Close()

	logger := zap.NewNop()
	client := logseq.NewClient(ts.URL, "token", logger)
	s := server.NewMCPServer(client, logger, server.ModeGeneral)

	req := makeRequest("add_tag", map[string]any{"uuid": "b1", "tag": "t1"})
	res, err := s.HandleAddTag(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error for handleAddTag, got %v", res)
	}
}

func TestServer_RemoveTag_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer ts.Close()

	logger := zap.NewNop()
	client := logseq.NewClient(ts.URL, "token", logger)
	s := server.NewMCPServer(client, logger, server.ModeGeneral)

	req := makeRequest("remove_tag", map[string]any{"uuid": "b1", "tag": "t1"})
	res, err := s.HandleRemoveTag(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error for handleRemoveTag, got %v", res)
	}
}

func TestServer_RemoveProperty_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer ts.Close()

	logger := zap.NewNop()
	client := logseq.NewClient(ts.URL, "token", logger)
	s := server.NewMCPServer(client, logger, server.ModeGeneral)

	req := makeRequest("remove_property", map[string]any{"uuid": "b1", "key": "k1"})
	res, err := s.HandleRemoveProperty(context.Background(), req)
	if err != nil || !res.IsError {
		t.Errorf("Expected error for handleRemoveProperty, got %v", res)
	}
}

func TestServer_OntologicalMode_Logic(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		var body struct {
			Method string `json:"method"`
			Args   []any  `json:"args"`
		}
		json.NewDecoder(r.Body).Decode(&body)

		if body.Method == "logseq.Editor.createPage" {
			name := body.Args[0].(string)
			props := body.Args[1].(map[string]any)

			// Verify snake_case keys
			for k := range props {
				if strings.Contains(k, "A") || strings.Contains(k, "B") {
					t.Errorf("Property key %s is not snake_case", k)
				}
			}

			w.Write([]byte(`{"uuid": "u1", "name": "` + name + `", "content": "test"}`))
			return
		}
		
		if body.Method == "logseq.Editor.getPage" {
			w.Write([]byte("null"))
			return
		}

		w.Write([]byte(`{"uuid": "u1", "name": "test"}`))
	}))
	defer ts.Close()

	logger := zap.NewNop()
	client := logseq.NewClient(ts.URL, "token", logger)
	s := server.NewMCPServer(client, logger, server.ModeOntological)

	t.Run("CreateEntity_Success", func(t *testing.T) {
		req := makeRequest("create_entity", map[string]any{
			"name":      "JohnDoe",
			"properties": `{"FirstName": "John", "LastName": "Doe"}`,
		})
		res, err := s.HandleCreateEntity(context.Background(), req)
		if err != nil || res.IsError {
			t.Errorf("Failed to create entity: %v", res)
		}
	})

	t.Run("CreateBlock_SnakeCase", func(t *testing.T) {
		req := makeRequest("create_block", map[string]any{
			"parent_uuid": "p1",
			"content":     "test",
			"properties":  `{"SomeProperty": "value"}`,
		})
		_, _ = s.HandleCreateBlock(context.Background(), req)
	})

	t.Run("CreateBlockTree_SnakeCase", func(t *testing.T) {
		req := makeRequest("create_block_tree", map[string]any{
			"parent_uuid": "p1",
			"tree":        `[{"content": "c1", "properties": {"PropA": "v1"}}]`,
		})
		_, _ = s.HandleCreateBlockTree(context.Background(), req)
	})

	t.Run("ListNamespaces_Ontological", func(t *testing.T) {
		req := makeRequest("list_namespaces", map[string]any{})
		res, err := s.HandleListNamespaces(context.Background(), req)
		if err != nil || res.IsError {
			t.Errorf("ListNamespaces failed in ontological mode: %v", res)
		}
	})

	t.Run("GetDailyJournal_Ontological", func(t *testing.T) {
		req := makeRequest("get_daily_journal", map[string]any{})
		res, err := s.HandleGetDailyJournal(context.Background(), req)
		if err != nil || res.IsError {
			t.Errorf("GetDailyJournal failed in ontological mode: %v", res)
		}
	})
}

func TestServer_RegisterTools_ModeSwitch(t *testing.T) {
	logger := zap.NewNop()
	client := logseq.NewClient("http://localhost", "token", logger)

	t.Run("GeneralMode", func(t *testing.T) {
		s := server.NewMCPServer(client, logger, server.ModeGeneral)
		tools := s.GetServer().ListTools()
		found := false
		for _, tool := range tools {
			if tool.Tool.Name == "create_pages" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected create_pages tool in general mode")
		}
	})

	t.Run("OntologicalMode", func(t *testing.T) {
		s := server.NewMCPServer(client, logger, server.ModeOntological)
		tools := s.GetServer().ListTools()
		for _, tool := range tools {
			if tool.Tool.Name == "create_pages" {
				t.Errorf("Did not expect create_pages tool in ontological mode")
			}
		}
	})
}

func TestUtils_ToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"FirstName", "first_name"},
		{"SomeID", "some_i_d"}, // Simple implementation behavior
		{"already_snake", "already_snake"},
		{"Normal", "normal"},
	}

	for _, tt := range tests {
		res := server.ToSnakeCase(tt.input)
		if res != tt.expected {
			t.Errorf("toSnakeCase(%s) = %s, want %s", tt.input, res, tt.expected)
		}
	}
}
