package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/clstb/yalms/pkg/logseq"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

type LogseqMode string

const (
	ModeGeneral     LogseqMode = "general"
	ModeOntological LogseqMode = "ontological"
)

type MCPServer struct {
	server *server.MCPServer
	client *logseq.Client
	logger *zap.Logger
	mode   LogseqMode
}

func NewMCPServer(client *logseq.Client, logger *zap.Logger, mode LogseqMode) *MCPServer {
	s := server.NewMCPServer("yalms", "0.1.0")
	ms := &MCPServer{
		server: s,
		client: client,
		logger: logger,
		mode:   mode,
	}

	ms.registerTools()
	return ms
}

func (s *MCPServer) Serve() error {
	return server.ServeStdio(s.server)
}

func (s *MCPServer) GetServer() *server.MCPServer {
	return s.server
}

func (s *MCPServer) HandleQuery(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleQuery(ctx, req)
}

func (s *MCPServer) HandleListNamespaces(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleListNamespaces(ctx, req)
}

func (s *MCPServer) HandleGetDailyJournal(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetDailyJournal(ctx, req)
}

func (s *MCPServer) HandleReadGraphInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleReadGraphInfo(ctx, req)
}

func (s *MCPServer) HandleReadPage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleReadPage(ctx, req)
}

func (s *MCPServer) HandleUpdatePage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUpdatePage(ctx, req)
}

func (s *MCPServer) HandleDeletePage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleDeletePage(ctx, req)
}

func (s *MCPServer) HandleCreatePages(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreatePages(ctx, req)
}

func (s *MCPServer) HandleDeletePages(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleDeletePages(ctx, req)
}

func (s *MCPServer) HandleRenamePage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleRenamePage(ctx, req)
}

func (s *MCPServer) HandleReadNamespace(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleReadNamespace(ctx, req)
}

func (s *MCPServer) HandleCreateNamespace(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreateNamespace(ctx, req)
}

func (s *MCPServer) HandleReadBlock(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleReadBlock(ctx, req)
}

func (s *MCPServer) HandleUpdateBlock(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUpdateBlock(ctx, req)
}

func (s *MCPServer) HandleDeleteBlock(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleDeleteBlock(ctx, req)
}

func (s *MCPServer) HandleDeleteBlocks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleDeleteBlocks(ctx, req)
}

func (s *MCPServer) HandleAppendBlock(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleAppendBlock(ctx, req)
}

func (s *MCPServer) HandleAddTag(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleAddTag(ctx, req)
}

func (s *MCPServer) HandleRemoveTag(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleRemoveTag(ctx, req)
}

func (s *MCPServer) HandleRemoveProperty(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleRemoveProperty(ctx, req)
}

func (s *MCPServer) HandleUpsertProperty(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUpsertProperty(ctx, req)
}

func (s *MCPServer) HandleCreateEntity(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreateEntity(ctx, req)
}

func (s *MCPServer) HandleCreateBlock(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreateBlock(ctx, req)
}

func (s *MCPServer) HandleCreateBlockTree(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreateBlockTree(ctx, req)
}

func ToSnakeCase(s string) string {
	return toSnakeCase(s)
}

func (s *MCPServer) registerTools() {
	// Graph Tools
	s.server.AddTool(mcp.NewTool("read_graph_info",
		mcp.WithDescription("Get information about the current graph"),
	), s.handleReadGraphInfo)

	s.server.AddTool(mcp.NewTool("query",
		mcp.WithDescription("Execute an advanced Datalog query against the Logseq database. Recommended for complex data retrieval and filtering. Examples: '[:find (pull ?p [*]) :where [?p :block/name]]' (all pages), '[:find (pull ?b [*]) :where [?b :block/content ?c] [(clojure.string/includes? ?c \"term\")]]' (blocks containing 'term')."),
		mcp.WithString("query", mcp.Required(), mcp.Description("The Datalog query string (e.g., '[:find (pull ?b [*]) :where ...]')")),
	), s.handleQuery)

	s.server.AddTool(mcp.NewTool("list_namespaces",
		mcp.WithDescription("List all existing namespaces/Classes in the graph."),
	), s.handleListNamespaces)

	s.server.AddTool(mcp.NewTool("get_daily_journal",
		mcp.WithDescription("Retrieve today's journal page details."),
	), s.handleGetDailyJournal)

	// Page/Entity Tools
	if s.mode == ModeOntological {
		s.server.AddTool(mcp.NewTool("read_entity",
			mcp.WithDescription("Retrieve structured data for an Instance (Particular). Use this to inspect record Attributes (data) and Relationships (links)."),
			mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID or name of the Instance (Particular)")),
		), s.handleReadPage)

		s.server.AddTool(mcp.NewTool("update_entity",
			mcp.WithDescription("Modify Instance Attributes or Relationships. Ensures data integrity by normalizing property keys to snake_case."),
			mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID or name of the Instance")),
			mcp.WithString("properties", mcp.Required(), mcp.Description("JSON string of updated Attributes (data) or Relationships (page links)")),
		), s.handleUpdatePage)

		s.server.AddTool(mcp.NewTool("delete_entity",
			mcp.WithDescription("Permanently remove an Instance record from the database."),
			mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID or name of the Instance")),
		), s.handleDeletePage)
	}

	if s.mode == ModeGeneral {
		s.server.AddTool(mcp.NewTool("read_page",
			mcp.WithDescription("Get page details. Returns the page properties and metadata."),
			mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID or name of the page")),
		), s.handleReadPage)
	}

	s.server.AddTool(mcp.NewTool("create_entity",
		mcp.WithDescription("Create a new Instance (Particular). Instances represent unique database entries. Classes (Universals) should be added as tags (e.g. #Person). Attributes (data) and Relationships (links) should be added as properties. Always use the returned UUID for subsequent operations."),
		mcp.WithString("name", mcp.Required(), mcp.Description("The specific name of the Instance (e.g. 'The Hobbit', 'Alice Smith')")),
		mcp.WithString("namespace", mcp.Description("The optional Class or category (e.g., 'Person', 'Project').")),
		mcp.WithString("properties", mcp.Description("JSON string of Attributes (e.g. 'published-date: 1937') or Relationships (e.g. 'author: [[J.R.R. Tolkien]]'). Keys will be converted to snake_case in ontological mode.")),
	), s.handleCreateEntity)

	if s.mode == ModeGeneral {
		s.server.AddTool(mcp.NewTool("create_pages",
			mcp.WithDescription("Create multiple pages. Use create_entity for ontological items."),
			mcp.WithString("pages", mcp.Required(), mcp.Description("JSON array of objects with 'name' and optional 'properties'")),
		), s.handleCreatePages)

		s.server.AddTool(mcp.NewTool("update_page",
			mcp.WithDescription("Update page properties. Use this to modify entity attributes."),
			mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID or name of the page")),
			mcp.WithString("properties", mcp.Required(), mcp.Description("JSON string of properties to update")),
		), s.handleUpdatePage)

		s.server.AddTool(mcp.NewTool("delete_page",
			mcp.WithDescription("Permanently delete a page/entity."),
			mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID or name of the page")),
		), s.handleDeletePage)

		s.server.AddTool(mcp.NewTool("delete_pages",
			mcp.WithDescription("Permanently delete multiple pages/entities."),
			mcp.WithString("uuids", mcp.Required(), mcp.Description("JSON array of page UUIDs or names to delete")),
		), s.handleDeletePages)
	}

	s.server.AddTool(mcp.NewTool("rename_page",
		mcp.WithDescription("Rename a page. Note: This may break ontological references if not handled carefully."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID or name of the page")),
		mcp.WithString("new_name", mcp.Required(), mcp.Description("The new name of the page")),
	), s.handleRenamePage)

	// Namespace Tools
	s.server.AddTool(mcp.NewTool("read_namespace",
		mcp.WithDescription("List all Instances within a specific Class or namespace hierarchy."),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The Class or category to list")),
	), s.handleReadNamespace)

	if s.mode == ModeGeneral {
		s.server.AddTool(mcp.NewTool("create_namespace",
			mcp.WithDescription("Create a new namespace or category level. Defines a high-level grouping."),
			mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace name (e.g. 'work/project')")),
		), s.handleCreateNamespace)
	}

	// Block Tools
	if s.mode == ModeOntological {
		s.server.AddTool(mcp.NewTool("read_entry",
			mcp.WithDescription("Read a specific entry (block) within an Instance outline."),
			mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID of the entry")),
		), s.handleReadBlock)

		s.server.AddTool(mcp.NewTool("update_entry",
			mcp.WithDescription("Modify an entry (block). In ontological mode, properties are normalized to snake_case."),
			mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID of the entry")),
			mcp.WithString("content", mcp.Required(), mcp.Description("The updated content")),
			mcp.WithString("properties", mcp.Description("JSON string of updated entry Attributes or Relationships")),
		), s.handleUpdateBlock)

		s.server.AddTool(mcp.NewTool("remove_entry",
			mcp.WithDescription("Permanently remove an entry from an Instance outline."),
			mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID of the entry")),
		), s.handleDeleteBlock)

		s.server.AddTool(mcp.NewTool("append_entry_to_entity",
			mcp.WithDescription("Append a new bullet point to an Instance. Use this to add data entries or notes in a clean outliner format. Do NOT use for bulk data; prefer create_entry_tree for structured trees."),
			mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID or name of the Instance page")),
			mcp.WithString("content", mcp.Required(), mcp.Description("The content of the bullet point")),
		), s.handleAppendBlock)

		s.server.AddTool(mcp.NewTool("create_entry",
			mcp.WithDescription("Insert an entry (block). Properties are normalized to snake_case."),
			mcp.WithString("parent_uuid", mcp.Required(), mcp.Description("The UUID of the parent entry or Instance page")),
			mcp.WithString("content", mcp.Required(), mcp.Description("The content of the entry")),
			mcp.WithString("properties", mcp.Description("JSON string of entry Attributes or Relationships")),
			mcp.WithBoolean("sibling", mcp.Description("Insert as sibling instead of child")),
			mcp.WithBoolean("before", mcp.Description("Insert before the reference entry (only if sibling=true)")),
		), s.handleCreateBlock)

		s.server.AddTool(mcp.NewTool("create_entry_tree",
			mcp.WithDescription("Insert a structured tree of entries. Preferred for complex data structures. This forces an outliner-style hierarchy. Example tree: '[{\"content\": \"Root\", \"children\": [{\"content\": \"Child\"}]}]'"),
			mcp.WithString("parent_uuid", mcp.Required(), mcp.Description("The UUID of the parent entry or page")),
			mcp.WithString("tree", mcp.Required(), mcp.Description("JSON array of BlockContent objects. Use nested 'children' to represent the outline hierarchy.")),
			mcp.WithBoolean("sibling", mcp.Description("Insert as sibling instead of child")),
			mcp.WithBoolean("before", mcp.Description("Insert before the reference entry (only if sibling=true)")),
		), s.handleCreateBlockTree)
	}

	if s.mode == ModeGeneral {
		s.server.AddTool(mcp.NewTool("read_block",
			mcp.WithDescription("Get block details, including content and nested properties."),
			mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID of the block")),
		), s.handleReadBlock)

		s.server.AddTool(mcp.NewTool("append_block",
			mcp.WithDescription("Append a block to the end of a page/entity."),
			mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID or name of the page")),
			mcp.WithString("content", mcp.Required(), mcp.Description("The content of the block")),
		), s.handleAppendBlock)

		s.server.AddTool(mcp.NewTool("update_block",
			mcp.WithDescription("Update existing block content or properties."),
			mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID of the block")),
			mcp.WithString("content", mcp.Required(), mcp.Description("The new content")),
			mcp.WithString("properties", mcp.Description("JSON string of properties")),
		), s.handleUpdateBlock)

		s.server.AddTool(mcp.NewTool("remove_block",
			mcp.WithDescription("Permanently remove a block."),
			mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID of the block")),
		), s.handleDeleteBlock)

		s.server.AddTool(mcp.NewTool("remove_blocks",
			mcp.WithDescription("Permanently remove multiple blocks."),
			mcp.WithString("uuids", mcp.Required(), mcp.Description("JSON array of block UUIDs to delete")),
		), s.handleDeleteBlocks)

		s.server.AddTool(mcp.NewTool("create_block",
			mcp.WithDescription("Insert a block."),
			mcp.WithString("parent_uuid", mcp.Required(), mcp.Description("The UUID of the parent block or page")),
			mcp.WithString("content", mcp.Required(), mcp.Description("The content of the block")),
			mcp.WithString("properties", mcp.Description("JSON string of block-level properties")),
			mcp.WithBoolean("sibling", mcp.Description("Insert as sibling instead of child")),
			mcp.WithBoolean("before", mcp.Description("Insert before the reference block (only if sibling=true)")),
		), s.handleCreateBlock)

		s.server.AddTool(mcp.NewTool("create_block_tree",
			mcp.WithDescription("Insert a structured tree of blocks."),
			mcp.WithString("parent_uuid", mcp.Required(), mcp.Description("The UUID of the parent block")),
			mcp.WithString("tree", mcp.Required(), mcp.Description("JSON array of BlockContent objects. Use nested 'children' to represent the outline hierarchy.")),
			mcp.WithBoolean("sibling", mcp.Description("Insert as sibling instead of child")),
			mcp.WithBoolean("before", mcp.Description("Insert before the reference block (only if sibling=true)")),
		), s.handleCreateBlockTree)
	}

	// Tag/Property Tools
	s.server.AddTool(mcp.NewTool("add_tag",
		mcp.WithDescription("Add a #tag for discoverability (Classes/Universals). If the target is a page and has no entries, a new empty block will be created to hold the tag."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID of the block/entry or page/entity")),
		mcp.WithString("tag", mcp.Required(), mcp.Description("The tag to add (e.g. 'Project' or '#Project')")),
	), s.handleAddTag)

	s.server.AddTool(mcp.NewTool("remove_tag",
		mcp.WithDescription("Remove a discovery tag (Class/Universal)."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID of the block/entry or page/entity")),
		mcp.WithString("tag", mcp.Required(), mcp.Description("The tag to remove (e.g. 'Project' or '#Project')")),
	), s.handleRemoveTag)

	s.server.AddTool(mcp.NewTool("remove_property",
		mcp.WithDescription("Remove a specific property/attribute/relationship."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID of the block/entry or page/entity")),
		mcp.WithString("key", mcp.Required(), mcp.Description("The property key to remove")),
	), s.handleRemoveProperty)

	s.server.AddTool(mcp.NewTool("add_property",
		mcp.WithDescription("Add or update a specific property/attribute (data) or relationship (link)."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("The UUID of the block/entry or page/entity")),
		mcp.WithString("key", mcp.Required(), mcp.Description("The property key to add or update")),
		mcp.WithString("value", mcp.Required(), mcp.Description("The property value (use [[Page Name]] for relationships)")),
	), s.handleUpsertProperty)
}

// Handlers

func parseArguments(req mcp.CallToolRequest, target any) error {
	argBytes, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return err
	}
	return json.Unmarshal(argBytes, target)
}

func (s *MCPServer) handleReadGraphInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleReadGraphInfo", zap.Any("req", req))
	graph, err := s.client.GetGraph()
	if err != nil {
		s.logger.Error("handleReadGraphInfo failed", zap.Error(err))
		return mcp.NewToolResultError("Could not retrieve graph information. Please ensure Logseq is running and the HTTP API is enabled in settings."), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Graph: %s\nPath: %s", graph.Name, graph.Path)), nil
}

func (s *MCPServer) handleQuery(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleQuery", zap.Any("req", req))
	var args struct {
		Query string `json:"query"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.Query == "" {
		return mcp.NewToolResultError("A query string is required. Please provide a valid Datalog query (e.g., '[:find (pull ?p [*]) :where [?p :block/name]]')."), nil
	}
	results, err := s.client.Query(args.Query)
	if err != nil {
		s.logger.Error("handleQuery failed", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("The query failed: %v. Please check your Datalog syntax or ensure the requested entities exist.", err)), nil
	}

	jsonResults, _ := json.MarshalIndent(results, "", "  ")
	return mcp.NewToolResultText(string(jsonResults)), nil
}

func (s *MCPServer) handleListNamespaces(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleListNamespaces", zap.Any("req", req))
	namespaces, err := s.client.ListNamespaces()
	if err != nil {
		s.logger.Error("handleListNamespaces failed", zap.Error(err))
		return mcp.NewToolResultError("Could not list namespaces. This may happen if the graph is empty or the API is unreachable."), nil
	}

	jsonResults, _ := json.MarshalIndent(namespaces, "", "  ")
	return mcp.NewToolResultText(string(jsonResults)), nil
}

func (s *MCPServer) handleGetDailyJournal(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleGetDailyJournal", zap.Any("req", req))
	page, err := s.client.GetDailyJournal()
	if err != nil {
		s.logger.Error("handleGetDailyJournal failed", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Could not retrieve the daily journal page: %v. Please check if Logseq is running.", err)), nil
	}
	if page == nil {
		return mcp.NewToolResultError("No journal page exists for today. You can create one by adding a block or property to it."), nil
	}

	jsonResults, _ := json.MarshalIndent(page, "", "  ")
	return mcp.NewToolResultText(string(jsonResults)), nil
}

func (s *MCPServer) handleReadPage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleReadPage", zap.Any("req", req))
	var args struct {
		UUID string `json:"uuid"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.UUID == "" {
		return mcp.NewToolResultError("A UUID or page name is required. Please provide the unique identifier for the page you wish to read."), nil
	}
	page, err := s.client.GetPage(args.UUID)
	if err != nil {
		s.logger.Error("handleReadPage failed", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Could not retrieve the page: %v. Please ensure the UUID or name is correct and the page exists.", err)), nil
	}
	if page == nil {
		return mcp.NewToolResultError(fmt.Sprintf("Page not found: '%s'. Please double-check the name or UUID. For namespaced pages, use the full path like 'Projects/MyTask'.", args.UUID)), nil
	}

	jsonPage, _ := json.MarshalIndent(page, "", "  ")
	return mcp.NewToolResultText(string(jsonPage)), nil
}

func (s *MCPServer) handleCreatePage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreateEntity(ctx, req)
}

func (s *MCPServer) handleCreateEntity(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleCreateEntity", zap.Any("req", req))
	var args struct {
		Name       string `json:"name"`
		Namespace  string `json:"namespace"`
		Properties string `json:"properties"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.Name == "" {
		return mcp.NewToolResultError("A name is required to create an entity. Please provide a title for the new page."), nil
	}

	var props map[string]any
	if args.Properties != "" {
		if err := json.Unmarshal([]byte(args.Properties), &props); err != nil {
			return mcp.NewToolResultError("The properties provided are not valid JSON. Please check your formatting and try again."), nil
		}
	} else {
		props = make(map[string]any)
	}

	fullName := args.Name
	if args.Namespace != "" {
		fullName = args.Namespace + "/" + args.Name
	}

	if s.mode == ModeOntological {
		// Qualified Name: Construct the title
		if args.Namespace != "" {
			fullName = args.Namespace + "/" + args.Name
		}

		// Snake_Case: Convert all keys in the properties dict to snake_case
		props = toSnakeCaseKeys(props)
	}

	page, err := s.client.CreatePage(fullName, props, nil)
	if err != nil {
		s.logger.Error("handleCreateEntity failed", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create the entity: %v. Please ensure the name is valid and doesn't contain forbidden characters.", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Entity created successfully: %s (UUID: %s). You should use this UUID for any further updates to this entity.", page.Name, page.UUID)), nil
}

func toSnakeCaseKeys(m map[string]any) map[string]any {
	newMap := make(map[string]any)
	for k, v := range m {
		newKey := toSnakeCase(k)
		newMap[newKey] = v
	}
	return newMap
}

func toSnakeCase(s string) string {
	var res strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			res.WriteRune('_')
		}
		res.WriteRune(unicode.ToLower(r))
	}
	return res.String()
}

func (s *MCPServer) handleCreatePages(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleCreatePages", zap.Any("req", req))
	var args struct {
		Pages string `json:"pages"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.Pages == "" {
		return mcp.NewToolResultError("A list of pages (JSON array) is required. Please provide the names and optional properties for the pages you wish to create."), nil
	}

	type PageReq struct {
		Name       string         `json:"name"`
		Properties map[string]any `json:"properties,omitempty"`
	}

	var pageReqs []PageReq
	if err := json.Unmarshal([]byte(args.Pages), &pageReqs); err != nil {
		return mcp.NewToolResultError("The pages list provided is not valid JSON. Please check your formatting and ensure it is a JSON array of page objects."), nil
	}

	count := 0
	var errs []string

	for _, req := range pageReqs {
		if _, err := s.client.CreatePage(req.Name, req.Properties, nil); err != nil {
			s.logger.Error("Failed to create page in handleCreatePages", zap.String("name", req.Name), zap.Error(err))
			errs = append(errs, fmt.Sprintf("%s: %v", req.Name, err))
		} else {
			count++
		}
	}

	if len(errs) > 0 {
		return mcp.NewToolResultError(fmt.Sprintf("Created %d pages, but failed for: %v. Please ensure all page names are valid.", count, errs)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully created %d pages.", count)), nil
}

func (s *MCPServer) handleUpdatePage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleUpdatePage", zap.Any("req", req))
	var args struct {
		UUID       string `json:"uuid"`
		Properties string `json:"properties"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.UUID == "" {
		return mcp.NewToolResultError("A UUID or page name is required. Please provide the unique identifier for the page you wish to update."), nil
	}
	if args.Properties == "" {
		return mcp.NewToolResultError("Updated properties (JSON string) are required. Please provide the attributes you wish to modify."), nil
	}

	var props map[string]any
	if err := json.Unmarshal([]byte(args.Properties), &props); err != nil {
		return mcp.NewToolResultError("The properties provided are not valid JSON. Please check your formatting and try again."), nil
	}

	// First get the page to get its UUID
	page, err := s.client.GetPage(args.UUID)
	if err != nil {
		s.logger.Error("handleUpdatePage failed to get page", zap.String("uuid", args.UUID), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Could not retrieve the page to update: %v. Please ensure the UUID is correct.", err)), nil
	}
	if page == nil {
		return mcp.NewToolResultError(fmt.Sprintf("Page not found: '%s'. Please double-check the name or UUID.", args.UUID)), nil
	}

	updatedPage, err := s.client.UpdatePage(page.UUID, props)
	if err != nil {
		s.logger.Error("handleUpdatePage failed", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update the page: %v. Please ensure the properties are valid for this entity.", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Page updated successfully: %s", updatedPage.UUID)), nil
}

func (s *MCPServer) handleDeletePage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleDeletePage", zap.Any("req", req))
	var args struct {
		UUID string `json:"uuid"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.UUID == "" {
		return mcp.NewToolResultError("A UUID or page name is required. Please provide the identifier for the page you wish to delete."), nil
	}
	if err := s.client.DeletePage(args.UUID); err != nil {
		s.logger.Error("handleDeletePage failed", zap.String("uuid", args.UUID), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete the page: %v. Please ensure the identifier is correct.", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Page successfully deleted: %s", args.UUID)), nil
}

func (s *MCPServer) handleDeletePages(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleDeletePages", zap.Any("req", req))
	var args struct {
		UUIDs string `json:"uuids"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.UUIDs == "" {
		return mcp.NewToolResultError("A list of UUIDs or names is required. Please provide a JSON array of page identifiers to delete."), nil
	}

	var uuids []string
	if err := json.Unmarshal([]byte(args.UUIDs), &uuids); err != nil {
		return mcp.NewToolResultError("The list of identifiers provided is not valid JSON. Please check your formatting and ensure it is a JSON array of strings."), nil
	}

	count := 0
	var errs []string

	for _, uuid := range uuids {
		if err := s.client.DeletePage(uuid); err != nil {
			s.logger.Error("Failed to delete page in handleDeletePages", zap.String("uuid", uuid), zap.Error(err))
			errs = append(errs, fmt.Sprintf("%s: %v", uuid, err))
		} else {
			count++
		}
	}

	if len(errs) > 0 {
		return mcp.NewToolResultError(fmt.Sprintf("Deleted %d pages, but failed for: %v. Please verify the remaining identifiers are correct.", count, errs)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted %d pages.", count)), nil
}

func (s *MCPServer) handleRenamePage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleRenamePage", zap.Any("req", req))
	var args struct {
		UUID    string `json:"uuid"`
		NewName string `json:"new_name"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.UUID == "" {
		return mcp.NewToolResultError("A UUID or current page name is required. Please provide the identifier for the page you wish to rename."), nil
	}
	if args.NewName == "" {
		return mcp.NewToolResultError("A new name is required. Please provide the target title for the page."), nil
	}

	// Resolve UUID if name provided
	page, err := s.client.GetPage(args.UUID)
	if err != nil {
		s.logger.Error("handleRenamePage failed to get page", zap.String("uuid", args.UUID), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Could not retrieve the page to rename: %v. Please ensure the current identifier is correct.", err)), nil
	}
	if page == nil {
		return mcp.NewToolResultError(fmt.Sprintf("Page not found: '%s'. Please double-check the current name or UUID.", args.UUID)), nil
	}

	if err := s.client.RenamePage(page.UUID, args.NewName); err != nil {
		s.logger.Error("handleRenamePage failed", zap.String("uuid", page.UUID), zap.String("new_name", args.NewName), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to rename the page: %v. Please ensure the new name is valid and not already in use.", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Page successfully renamed to: %s", args.NewName)), nil
}

func (s *MCPServer) handleReadNamespace(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleReadNamespace", zap.Any("req", req))
	var args struct {
		Namespace string `json:"namespace"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.Namespace == "" {
		return mcp.NewToolResultError("A namespace name is required. Please provide the category (e.g., 'Projects') you wish to list."), nil
	}

	pages, err := s.client.GetNamespacePages(args.Namespace)
	if err != nil {
		s.logger.Error("handleReadNamespace failed", zap.String("namespace", args.Namespace), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Could not retrieve pages for namespace '%s': %v. Please ensure the namespace exists.", args.Namespace, err)), nil
	}

	jsonPages, _ := json.MarshalIndent(pages, "", "  ")
	return mcp.NewToolResultText(string(jsonPages)), nil
}

func (s *MCPServer) handleCreateNamespace(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleCreateNamespace", zap.Any("req", req))
	var args struct {
		Namespace string `json:"namespace"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.Namespace == "" {
		return mcp.NewToolResultError("A namespace name is required. Please provide the name (e.g., 'work/project') for the new category."), nil
	}

	// Creating a namespace is essentially creating a page with "/" in the name
	page, err := s.client.CreatePage(args.Namespace, nil, nil)
	if err != nil {
		s.logger.Error("handleCreateNamespace failed", zap.String("namespace", args.Namespace), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create the namespace page: %v. Please ensure the name is valid.", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Namespace created successfully: %s (UUID: %s)", page.Name, page.UUID)), nil
}

func (s *MCPServer) handleReadBlock(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleReadBlock", zap.Any("req", req))
	var args struct {
		UUID string `json:"uuid"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.UUID == "" {
		return mcp.NewToolResultError("A block UUID is required. Please provide the unique identifier for the block you wish to read."), nil
	}
	block, err := s.client.GetBlock(args.UUID)
	if err != nil {
		s.logger.Error("handleReadBlock failed", zap.String("uuid", args.UUID), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Could not retrieve the block: %v. Please ensure the UUID is correct.", err)), nil
	}
	if block == nil {
		return mcp.NewToolResultError(fmt.Sprintf("Block not found: '%s'. Please double-check the UUID.", args.UUID)), nil
	}

	jsonBlock, _ := json.MarshalIndent(block, "", "  ")
	return mcp.NewToolResultText(string(jsonBlock)), nil
}

func (s *MCPServer) handleCreateBlock(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleCreateBlock", zap.Any("req", req))
	var args struct {
		ParentUUID string `json:"parent_uuid"`
		Content    string `json:"content"`
		Properties string `json:"properties"`
		Sibling    bool   `json:"sibling"`
		Before     bool   `json:"before"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.ParentUUID == "" {
		return mcp.NewToolResultError("A parent UUID (page or block) is required. Please provide a valid identifier for where the block should be inserted."), nil
	}
	if args.Content == "" {
		return mcp.NewToolResultError("Block content is required. Please provide the text for the new block."), nil
	}

	var props map[string]any
	if args.Properties != "" {
		if err := json.Unmarshal([]byte(args.Properties), &props); err != nil {
			return mcp.NewToolResultError("The properties provided are not valid JSON. Please check your formatting and try again."), nil
		}
	}

	if s.mode == ModeOntological {
		props = toSnakeCaseKeys(props)
	}

	options := make(map[string]any)
	if args.Sibling {
		options["sibling"] = true
		if args.Before {
			options["before"] = true
		}
	}

	block, err := s.client.InsertBlock(args.ParentUUID, args.Content, props, options)
	if err != nil {
		s.logger.Error("handleCreateBlock failed", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to insert the block: %v. Please ensure the parent exists and the content is valid.", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Block inserted successfully: %s. You can use this UUID to reference or update this block later.", block.UUID)), nil
}

func (s *MCPServer) handleCreateBlockTree(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleCreateBlockTree", zap.Any("req", req))
	var args struct {
		ParentUUID string `json:"parent_uuid"`
		Tree       string `json:"tree"`
		Sibling    bool   `json:"sibling"`
		Before     bool   `json:"before"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.ParentUUID == "" {
		return mcp.NewToolResultError("A parent UUID (page or block) is required. Please provide a valid identifier for where the tree should be inserted."), nil
	}
	if args.Tree == "" {
		return mcp.NewToolResultError("A tree structure (JSON array) is required. Please provide a valid list of blocks and their nested children."), nil
	}

	var batch []logseq.BlockContent
	if err := json.Unmarshal([]byte(args.Tree), &batch); err != nil {
		return mcp.NewToolResultError("The tree structure provided is not valid JSON. Please check your formatting and ensure it matches the BlockContent structure."), nil
	}

	if s.mode == ModeOntological {
		var transformTree func([]logseq.BlockContent)
		transformTree = func(blocks []logseq.BlockContent) {
			for i := range blocks {
				blocks[i].Properties = toSnakeCaseKeys(blocks[i].Properties)
				if len(blocks[i].Children) > 0 {
					transformTree(blocks[i].Children)
				}
			}
		}
		transformTree(batch)
	}

	options := make(map[string]any)
	if args.Sibling {
		options["sibling"] = true
		if args.Before {
			options["before"] = true
		}
	}

	blocks, err := s.client.InsertBatchBlock(args.ParentUUID, batch, options)
	if err != nil {
		s.logger.Error("handleCreateBlockTree failed", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to insert the block tree: %v. Please ensure the parent exists and the tree structure is valid.", err)), nil
	}

	// Create a summary of created blocks
	return mcp.NewToolResultText(fmt.Sprintf("Successfully inserted %d blocks into the tree.", len(blocks))), nil
}

func (s *MCPServer) handleAppendBlock(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleAppendBlock", zap.Any("req", req))
	var args struct {
		UUID    string `json:"uuid"`
		Content string `json:"content"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.UUID == "" {
		return mcp.NewToolResultError("A page name or UUID is required. Please provide the identifier for the page where the block should be appended."), nil
	}
	if args.Content == "" {
		return mcp.NewToolResultError("Block content is required. Please provide the text to append."), nil
	}

	block, err := s.client.AppendBlockInPage(args.UUID, args.Content, nil)
	if err != nil {
		s.logger.Error("handleAppendBlock failed", zap.String("uuid", args.UUID), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to append the block: %v. Please ensure the page exists.", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Block successfully appended to '%s'. New block UUID: %s", args.UUID, block.UUID)), nil
}

func (s *MCPServer) handleUpdateBlock(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleUpdateBlock", zap.Any("req", req))
	var args struct {
		UUID       string `json:"uuid"`
		Content    string `json:"content"`
		Properties string `json:"properties"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.UUID == "" {
		return mcp.NewToolResultError("A block UUID is required. Please provide the unique identifier for the block you wish to update."), nil
	}

	var props map[string]any
	if args.Properties != "" {
		if err := json.Unmarshal([]byte(args.Properties), &props); err != nil {
			return mcp.NewToolResultError("The properties provided are not valid JSON. Please check your formatting."), nil
		}
	}

	if s.mode == ModeOntological {
		props = toSnakeCaseKeys(props)
	}

	block, err := s.client.UpdateBlock(args.UUID, args.Content, props)
	if err != nil {
		s.logger.Error("handleUpdateBlock failed", zap.String("uuid", args.UUID), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update the block: %v. Please ensure the UUID is correct and the block still exists.", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Block updated successfully: %s", block.UUID)), nil
}

func (s *MCPServer) handleDeleteBlock(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleDeleteBlock", zap.Any("req", req))
	var args struct {
		UUID string `json:"uuid"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.UUID == "" {
		return mcp.NewToolResultError("A block UUID is required. Please provide the identifier for the block you wish to delete."), nil
	}
	if err := s.client.DeleteBlock(args.UUID); err != nil {
		s.logger.Error("handleDeleteBlock failed", zap.String("uuid", args.UUID), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete the block: %v. Please ensure the UUID is correct.", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Block successfully deleted: %s", args.UUID)), nil
}

func (s *MCPServer) handleDeleteBlocks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleDeleteBlocks", zap.Any("req", req))
	var args struct {
		UUIDs string `json:"uuids"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.UUIDs == "" {
		return mcp.NewToolResultError("A list of UUIDs is required. Please provide a JSON array of block identifiers to delete."), nil
	}

	var uuids []string
	if err := json.Unmarshal([]byte(args.UUIDs), &uuids); err != nil {
		return mcp.NewToolResultError("The list of UUIDs provided is not valid JSON. Please check your formatting and ensure it is a JSON array of strings."), nil
	}

	count := 0
	var errs []string

	for _, uuid := range uuids {
		if err := s.client.DeleteBlock(uuid); err != nil {
			s.logger.Error("Failed to delete block in handleDeleteBlocks", zap.String("uuid", uuid), zap.Error(err))
			errs = append(errs, fmt.Sprintf("%s: %v", uuid, err))
		} else {
			count++
		}
	}

	if len(errs) > 0 {
		return mcp.NewToolResultError(fmt.Sprintf("Deleted %d blocks, but failed for: %v. Please verify the remaining UUIDs are correct.", count, errs)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted %d blocks.", count)), nil
}

func (s *MCPServer) handleAddTag(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleAddTag", zap.Any("req", req))
	var args struct {
		UUID string `json:"uuid"`
		Tag  string `json:"tag"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.UUID == "" {
		return mcp.NewToolResultError("A UUID or page name is required. Please provide the identifier for the entity you wish to tag."), nil
	}
	if args.Tag == "" {
		return mcp.NewToolResultError("A tag is required. Please provide the text for the tag you wish to add."), nil
	}

	if err := s.client.AddTag(args.UUID, args.Tag); err != nil {
		s.logger.Error("handleAddTag failed", zap.String("uuid", args.UUID), zap.String("tag", args.Tag), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add the tag: %v. Please ensure the target exists and the tag format is valid.", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Tag '%s' successfully added to %s.", args.Tag, args.UUID)), nil
}

func (s *MCPServer) handleRemoveTag(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleRemoveTag", zap.Any("req", req))
	var args struct {
		UUID string `json:"uuid"`
		Tag  string `json:"tag"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.UUID == "" {
		return mcp.NewToolResultError("A UUID or page name is required. Please provide the identifier for the entity from which to remove the tag."), nil
	}
	if args.Tag == "" {
		return mcp.NewToolResultError("A tag is required. Please provide the text for the tag you wish to remove."), nil
	}

	if err := s.client.RemoveTag(args.UUID, args.Tag); err != nil {
		s.logger.Error("handleRemoveTag failed", zap.String("uuid", args.UUID), zap.String("tag", args.Tag), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to remove the tag: %v. Please ensure the entity exists and contains the specified tag.", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Tag '%s' successfully removed from %s.", args.Tag, args.UUID)), nil
}

func (s *MCPServer) handleRemoveProperty(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleRemoveProperty", zap.Any("req", req))
	var args struct {
		UUID string `json:"uuid"`
		Key  string `json:"key"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.UUID == "" {
		return mcp.NewToolResultError("A UUID or page name is required. Please provide the identifier for the entity from which to remove the property."), nil
	}
	if args.Key == "" {
		return mcp.NewToolResultError("A property key is required. Please provide the name of the attribute you wish to remove."), nil
	}

	if err := s.client.RemoveProperty(args.UUID, args.Key); err != nil {
		s.logger.Error("handleRemoveProperty failed", zap.String("uuid", args.UUID), zap.String("key", args.Key), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to remove the property: %v. Please ensure the entity exists and contains the specified attribute.", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Property '%s' successfully removed from %s.", args.Key, args.UUID)), nil
}

func (s *MCPServer) handleUpsertProperty(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("handleUpsertProperty", zap.Any("req", req))
	var args struct {
		UUID  string `json:"uuid"`
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := parseArguments(req, &args); err != nil {
		return mcp.NewToolResultError("Invalid arguments provided. Please check the tool definition and try again."), nil
	}
	if args.UUID == "" {
		return mcp.NewToolResultError("A UUID or page name is required. Please provide the identifier for the entity to which to add/update the property."), nil
	}
	if args.Key == "" {
		return mcp.NewToolResultError("A property key is required. Please provide the name of the attribute you wish to add/update."), nil
	}

	key := args.Key
	if s.mode == ModeOntological {
		key = ToSnakeCase(key)
	}

	if err := s.client.UpsertProperty(args.UUID, key, args.Value); err != nil {
		s.logger.Error("handleUpsertProperty failed", zap.String("uuid", args.UUID), zap.String("key", key), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add/update the property: %v. Please ensure the entity exists.", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Property '%s' successfully added/updated on %s.", key, args.UUID)), nil
}
