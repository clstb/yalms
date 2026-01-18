package logseq

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type Client struct {
	client *resty.Client
	logger *zap.Logger
	token  string
	apiURL string
}

func NewClient(apiURL, token string, logger *zap.Logger) *Client {
	c := resty.New()
	c.SetBaseURL(apiURL)
	c.SetTimeout(10 * time.Second)
	c.SetHeader("Authorization", "Bearer "+token)
	c.SetHeader("Content-Type", "application/json")

	return &Client{
		client: c,
		logger: logger,
		token:  token,
		apiURL: apiURL,
	}
}

// Generic request structure for Logseq API
type apiRequest struct {
	Method string        `json:"method"`
	Args   []interface{} `json:"args"`
}

func (c *Client) Call(method string, args ...any) ([]byte, error) {
	reqBody := apiRequest{
		Method: method,
		Args:   args,
	}

	if c.logger != nil {
		c.logger.Debug("Logseq API Call", zap.String("method", method), zap.Any("args", args))
	}

	resp, err := c.client.R().
		SetBody(reqBody).
		Post("/api")

	if err != nil {
		if c.logger != nil {
			c.logger.Error("Logseq API request failed", zap.String("method", method), zap.Error(err))
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.IsError() {
		if c.logger != nil {
			c.logger.Error("Logseq API error response", zap.String("method", method), zap.Int("status", resp.StatusCode()), zap.String("body", resp.String()))
		}
		return nil, fmt.Errorf("api error: %s (status: %d)", resp.String(), resp.StatusCode())
	}

	// Logseq API often returns errors as JSON with an "error" field even with 200 OK
	var errCheck struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(resp.Body(), &errCheck); err == nil && errCheck.Error != "" {
		if c.logger != nil {
			c.logger.Error("Logseq API business error", zap.String("method", method), zap.String("error", errCheck.Error))
		}
		return nil, fmt.Errorf("api error: %s", errCheck.Error)
	}

	return resp.Body(), nil
}

// Graph Methods

func (c *Client) GetGraph() (*GraphInfo, error) {
	resp, err := c.Call("logseq.App.getCurrentGraph")
	if err != nil {
		return nil, err
	}
	var graph GraphInfo
	if err := json.Unmarshal(resp, &graph); err != nil {
		// Logseq might return null if no graph is open, or different structure
		return nil, fmt.Errorf("failed to parse graph info: %w", err)
	}
	return &graph, nil
}

// Page Methods

func (c *Client) RenamePage(uuid string, newName string) error {
	// 1. Rename to new name directly
	if _, err := c.Call("logseq.Editor.renamePage", uuid, newName); err != nil {
		return fmt.Errorf("failed to rename page: %w", err)
	}
	
	return nil
}

func (c *Client) GetPage(nameOrUUID string) (*Page, error) {
	// 1. Try direct lookup (UUID or Name)
	resp, err := c.Call("logseq.Editor.getPage", nameOrUUID)
	if err != nil {
		return nil, err
	}
	
	// If found, return it
	if string(resp) != "null" && string(resp) != "[]" {
		var page Page
		if err := json.Unmarshal(resp, &page); err == nil {
			return &page, nil
		}
	}
	
	return nil, nil
}

func (c *Client) _RenamePage_Legacy(uuid string, newName string) error {
	// Keep this signature valid for now by not using it directly
	return nil
}

func (c *Client) CreatePage(name string, properties map[string]any, options map[string]any) (*Page, error) {
	// Prepare properties
	if properties == nil {
		properties = make(map[string]any)
	}

	// 1. Check if page exists to avoid overwrite/error
	if existing, err := c.GetPage(name); err == nil && existing != nil {
		// Idempotent: It's the same page. Return it.
		if len(properties) > 0 {
			c.UpdatePage(existing.UUID, properties)
			return c.GetPage(existing.UUID)
		}
		return existing, nil
	}

	// 2. Create Page directly
	args := []any{name, properties}
	if options != nil {
		args = append(args, options)
	}
	
	if c.logger != nil {
		c.logger.Debug("Attempting CreatePage", zap.String("name", name))
	}

	resp, err := c.Call("logseq.Editor.createPage", args...)
	
	var page Page
	success := false
	
	if err == nil {
		if jsonErr := json.Unmarshal(resp, &page); jsonErr == nil {
			// Check if we got what we asked for
			if page.UUID == "" {
				if c.logger != nil {
					c.logger.Warn("CreatePage returned empty UUID", zap.String("requested", name))
				}
				success = false
			} else {
				success = true
			}
		} else {
			if c.logger != nil {
				c.logger.Warn("CreatePage JSON parse error", zap.Error(jsonErr))
			}
		}
	} else {
		if c.logger != nil {
			c.logger.Warn("CreatePage failed", zap.Error(err))
		}
		return nil, err
	}

	if success {
		// Ensure properties (createPage might skip them sometimes)
		if len(properties) > 0 {
			c.UpdatePage(page.UUID, properties)
			if updated, err := c.GetPage(page.UUID); err == nil && updated != nil {
				return updated, nil
			}
		}
		
		return &page, nil
	}

	return nil, fmt.Errorf("failed to create page '%s'", name)
}

func (c *Client) UpdatePage(uuid string, properties map[string]any) (*Page, error) {
	// Auto-create linked pages from properties
	c.EnsureLinkedPages("", properties)

	// Use upsertBlockProperty for each property to ensure they are applied to the page
	// Logseq API: logseq.Editor.upsertBlockProperty(block/page, key, value)
	for k, v := range properties {
		_, err := c.Call("logseq.Editor.upsertBlockProperty", uuid, k, v)
		if err != nil {
			return nil, fmt.Errorf("failed to update property %s: %w", k, err)
		}
	}
	
	// Fetch updated page
	// Wait a tiny bit? No, API should be synchronous enough or consistent.
	return c.GetPage(uuid)
}

func (c *Client) DeletePage(nameOrUUID string) error {
	_, err := c.Call("logseq.Editor.deletePage", nameOrUUID)
	return err
}

func (c *Client) UpsertProperty(uuid string, key string, value any) error {
	_, err := c.Call("logseq.Editor.upsertBlockProperty", uuid, key, value)
	return err
}

func (c *Client) RemoveProperty(uuid string, key string) error {
	_, err := c.Call("logseq.Editor.removeBlockProperty", uuid, key)
	return err
}

// Namespace Methods

func (c *Client) GetNamespacePages(namespace string) ([]Page, error) {
	// Find all pages where the parent is the specified namespace page
	datalog := fmt.Sprintf(`[:find (pull ?p [*]) :where [?p :block/name] [?p :block/parent ?parent] [?parent :block/name "%s"]]`, strings.ToLower(namespace))

	if c.logger != nil {
		c.logger.Debug("GetNamespacePages Query", zap.String("namespace", namespace), zap.String("query", datalog))
	}

	results, err := c.Query(datalog)
	if err != nil {
		return nil, err
	}

	var pages []Page
	if list, ok := results.([]any); ok {
		for _, item := range list {
			pageBytes, _ := json.Marshal(item)
			var p Page
			if err := json.Unmarshal(pageBytes, &p); err == nil && p.UUID != "" {
				pages = append(pages, p)
			}
		}
	}

	return pages, nil
}

// Block Methods

func (c *Client) GetBlock(uuid string) (*Block, error) {
	resp, err := c.Call("logseq.Editor.getBlock", uuid, true) // include children
	if err != nil {
		return nil, err
	}
	if string(resp) == "null" {
		if c.logger != nil {
			c.logger.Debug("GetBlock returned null", zap.String("uuid", uuid))
		}
		return nil, nil
	}
	var block Block
	if err := json.Unmarshal(resp, &block); err != nil {
		return nil, fmt.Errorf("failed to parse block: %w", err)
	}
	// Debug: check properties
	if c.logger != nil {
		// Log properties map keys only to reduce noise? No, full properties.
		// c.logger.Debug("GetBlock properties", zap.String("uuid", uuid), zap.Any("props", block.Properties))
	}
	return &block, nil
}

// Tag Methods (Text-based #Tag)

func (c *Client) getEntityBlock(uuid string) (*Block, error) {
	block, err := c.GetBlock(uuid)
	if err != nil {
		return nil, err
	}
	if block != nil {
		return block, nil
	}

	// Try as page
	page, err := c.GetPage(uuid)
	if err != nil || page == nil {
		return nil, fmt.Errorf("entity not found: %s", uuid)
	}

	// Try to find block by UUID which matches page UUID for page properties block
	block, err = c.GetBlock(page.UUID)
	if err != nil {
		return nil, err
	}
	if block != nil {
		return block, nil
	}

	// Try fetching page blocks tree and take the first block
	blocksRaw, err := c.Call("logseq.Editor.getPageBlocksTree", page.UUID)
	if err == nil && string(blocksRaw) != "null" && string(blocksRaw) != "[]" {
		var bTree []Block
		if err := json.Unmarshal(blocksRaw, &bTree); err == nil && len(bTree) > 0 {
			return &bTree[0], nil
		}
	}

	return nil, fmt.Errorf("failed to find content block for page: %s", uuid)
}

func (c *Client) AddTag(uuid string, tag string) error {
	block, err := c.getEntityBlock(uuid)
	if err != nil {
		// If block not found, try to append an empty block if it's a page
		if strings.Contains(err.Error(), "failed to find content block") {
			block, err = c.AppendBlockInPage(uuid, "", nil)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	cleanTag := strings.TrimPrefix(tag, "#")
	tagStr := "#" + cleanTag
	if strings.Contains(block.Content, tagStr) {
		return nil // Already tagged
	}

	newContent := strings.TrimSpace(block.Content) + " " + tagStr
	_, err = c.UpdateBlock(block.UUID, newContent, nil)
	return err
}

func (c *Client) RemoveTag(uuid string, tag string) error {
	block, err := c.getEntityBlock(uuid)
	if err != nil {
		return err
	}

	cleanTag := strings.TrimPrefix(tag, "#")
	tagStr := "#" + cleanTag

	if !strings.Contains(block.Content, tagStr) {
		return nil
	}

	newContent := strings.ReplaceAll(block.Content, tagStr, "")
	newContent = strings.ReplaceAll(newContent, "  ", " ") // Cleanup spaces
	newContent = strings.TrimSpace(newContent)

	_, err = c.UpdateBlock(block.UUID, newContent, nil)
	return err
}

func (c *Client) EnsureLinkedPages(content string, properties map[string]any) string {
	// 1. From Content
	links := extractLinks(content)
	
	// 2. From Properties (values)
	for _, v := range properties {
		if str, ok := v.(string); ok {
			propLinks := extractLinks(str)
			links = append(links, propLinks...)
		}
	}
	
	// 3. Ensure each page exists (handling namespaces)
	// Deduplicate first
	uniqueLinks := make(map[string]bool)
	var linkList []string
	for _, l := range links {
		if !uniqueLinks[l] {
			uniqueLinks[l] = true
			linkList = append(linkList, l)
		}
	}

	for _, link := range linkList {
		// Handle hierarchy: if "A/B/C", ensure "A", then "A/B", then "A/B/C"
		parts := strings.Split(link, "/")
		currentPath := ""
		
		// If it's a namespaced page, we need its final UUID for replacement
		var finalUUID string
		
		for i, part := range parts {
			if currentPath == "" {
				currentPath = part
			} else {
				currentPath = currentPath + "/" + part
			}
			
			// Check if exists
			if c.logger != nil {
				c.logger.Debug("Checking linked page existence", zap.String("page", currentPath))
			}
			
			page, err := c.GetPage(currentPath)
			if err == nil && page != nil {
				if i == len(parts)-1 {
					finalUUID = page.UUID
				}
				continue // Exists
			}
			
			// Not found -> Create
			if c.logger != nil {
				c.logger.Info("Auto-creating missing linked page/namespace", zap.String("page", currentPath))
			}
			
			created, err := c.CreatePage(currentPath, nil, nil)
			if err != nil {
				if c.logger != nil {
					c.logger.Error("Failed to auto-create linked page", zap.String("page", currentPath), zap.Error(err))
				}
			} else if created != nil && i == len(parts)-1 {
				finalUUID = created.UUID
			}
		}
		
		// 4. Replace namespaced links with UUID refs in content
		// User requirement: "for linking in blocks to pages that are in a namesapce we ALWAYS use a uuid"
		if strings.Contains(link, "/") && finalUUID != "" {
			// Replace [[Name/Space]] with ((UUID))
			// Be careful with replacement to avoid partial matches if possible, but [[...]] is specific.
			// extractLinks returns the name inside [[...]].
			target := "[[" + link + "]]"
			replacement := "((" + finalUUID + "))"
			if c.logger != nil {
				c.logger.Debug("Replacing namespaced link with UUID", zap.String("original", target), zap.String("replacement", replacement))
			}
			content = strings.ReplaceAll(content, target, replacement)
			
			// Update Properties
			for k, v := range properties {
				if str, ok := v.(string); ok && strings.Contains(str, target) {
					properties[k] = strings.ReplaceAll(str, target, replacement)
				}
			}
		}
	}

	// 5. Verify Block Refs ((uuid))
	// We just check if they exist and log warning if not
	refs := extractBlockRefs(content)
	for _, v := range properties {
		if str, ok := v.(string); ok {
			propRefs := extractBlockRefs(str)
			refs = append(refs, propRefs...)
		}
	}

	for _, uuid := range refs {
		block, err := c.GetBlock(uuid)
		if err != nil || block == nil {
			if c.logger != nil {
				c.logger.Warn(" Referenced block not found", zap.String("uuid", uuid))
			}
		}
	}
	
	return content
}

func (c *Client) InsertBlock(parentUUID string, content string, properties map[string]any, options map[string]any) (*Block, error) {
	// Auto-create linked pages before insertion and update content with UUIDs for namespaces
	content = c.EnsureLinkedPages(content, properties)

	args := []any{parentUUID, content}
	if options != nil {
		args = append(args, options)
	}
	
	resp, err := c.Call("logseq.Editor.insertBlock", args...)
	if err != nil {
		return nil, err
	}
	var block Block
	if err := json.Unmarshal(resp, &block); err != nil {
		return nil, fmt.Errorf("failed to parse inserted block: %w", err)
	}

	// Apply properties if provided
	if len(properties) > 0 {
		// We use UpdateBlock to apply properties in batch if possible, or loop upsert
		// UpdateBlock overwrites content, so we pass the same content
		_, err := c.UpdateBlock(block.UUID, content, properties)
		if err != nil {
			if c.logger != nil {
				c.logger.Error("Failed to apply properties to new block", zap.Error(err))
			}
			// Return block anyway, but maybe with error? 
			// Better to error out so user knows properties failed
			return &block, fmt.Errorf("block created but properties failed: %w", err)
		}
		// Refresh block
		return c.GetBlock(block.UUID)
	}

	return &block, nil
}

func (c *Client) InsertBatchBlock(parentUUID string, batch []BlockContent, options map[string]any) ([]Block, error) {
	// Auto-create linked pages for all blocks in batch
	// This might be recursive for children
	var scanBlocks func([]BlockContent)
	scanBlocks = func(blocks []BlockContent) {
		for i := range blocks {
			blocks[i].Content = c.EnsureLinkedPages(blocks[i].Content, blocks[i].Properties)
			if len(blocks[i].Children) > 0 {
				scanBlocks(blocks[i].Children)
			}
		}
	}
	scanBlocks(batch)

	// logseq.Editor.insertBatchBlock(parent, batch, options)
	args := []any{parentUUID, batch}
	if options != nil {
		args = append(args, options)
	}
	resp, err := c.Call("logseq.Editor.insertBatchBlock", args...)
	if err != nil {
		return nil, err
	}
	
	// Returns a list of created blocks (or sometimes null if failed silently, but we check err)
	var blocks []Block
	if err := json.Unmarshal(resp, &blocks); err != nil {
		// Sometimes it might return a list of lists or different structure if only one block?
		// But insertBatchBlock usually returns array of Blocks.
		return nil, fmt.Errorf("failed to parse batch blocks: %w", err)
	}
	
	// If options contain properties for specific blocks in the batch, they should have been in the batch structure.
	// But if we passed global options? `insertBatchBlock` has options for `sibling`, `before`.
	
	return blocks, nil
}

func (c *Client) UpdateBlock(uuid string, content string, properties map[string]any) (*Block, error) {
	// Auto-create linked pages before update and update content with UUIDs for namespaces
	content = c.EnsureLinkedPages(content, properties)

	args := []any{uuid, content}
	if properties != nil {
		args = append(args, properties)
	}
	resp, err := c.Call("logseq.Editor.updateBlock", args...)
	if err != nil {
		return nil, err
	}

	// 1. Try parsing full block response
	var block Block
	if err := json.Unmarshal(resp, &block); err == nil && block.UUID != "" {
		return &block, nil
	}

	// 2. Fallback: If response is just a UUID string or doesn't match Block struct, fetch it
	var respUUID string
	if err := json.Unmarshal(resp, &respUUID); err == nil && respUUID != "" {
		return c.GetBlock(respUUID)
	}

	// 3. Last resort: Return the block with requested UUID and hope it updated (or fetch it)
	return c.GetBlock(uuid)
}

func (c *Client) DeleteBlock(uuid string) error {
	_, err := c.Call("logseq.Editor.removeBlock", uuid)
	return err
}

func (c *Client) AppendBlockInPage(pageName string, content string, options map[string]any) (*Block, error) {
	// Auto-create linked pages and update content
	content = c.EnsureLinkedPages(content, nil)

	args := []any{pageName, content}
	if options != nil {
		args = append(args, options)
	}
	resp, err := c.Call("logseq.Editor.appendBlockInPage", args...)
	if err != nil {
		return nil, err
	}
	var block Block
	if err := json.Unmarshal(resp, &block); err != nil {
		return nil, fmt.Errorf("failed to parse appended block: %w", err)
	}
	return &block, nil
}

// Search
func (c *Client) Query(datalog string) (any, error) {
	resp, err := c.Call("logseq.DB.q", datalog)
	if err != nil {
		return nil, err
	}

	// Fallback to datascriptQuery if q returns empty
	if string(resp) == "[]" || string(resp) == "null" {
		respDS, err := c.Call("logseq.DB.datascriptQuery", datalog)
		if err == nil && string(respDS) != "[]" && string(respDS) != "null" {
			resp = respDS
		}
	}

	var results any
	if err := json.Unmarshal(resp, &results); err != nil {
		return nil, fmt.Errorf("failed to parse query results: %w", err)
	}

	// Flatten if it's list of lists (common for simple finds)
	if list, ok := results.([]any); ok {
		if len(list) > 0 {
			if _, ok := list[0].([]any); ok {
				var flat []any
				for _, item := range list {
					if sublist, ok := item.([]any); ok && len(sublist) > 0 {
						flat = append(flat, sublist[0])
					}
				}
				return flat, nil
			}
		}
	}

	return results, nil
}

func (c *Client) GetDailyJournal() (any, error) {
	// 1. Try logseq.App.getTodayJournalPage first
	// Note: We swallow 500 error here to allow fallback if the method is undefined in this version
	resp, err := c.Call("logseq.App.getTodayJournalPage")
	
	if err == nil {
		if string(resp) != "null" && string(resp) != "[]" {
			var page Page
			if err := json.Unmarshal(resp, &page); err == nil && page.UUID != "" {
				return &page, nil
			}
			// Fallback to raw map if unmarshal fails
			var raw map[string]any
			if err := json.Unmarshal(resp, &raw); err == nil {
				return raw, nil
			}
		}
	}

	// 2. Fallback: Search for pages with specific journal-day match
	// This avoids the 500 "apply" error if getTodayJournalPage is missing
	now := time.Now()
	todayInt := now.Format("20060102")
	datalog := fmt.Sprintf(`[:find (pull ?p [*]) :where [?p :block/journal-day %s]]`, todayInt)

	if c.logger != nil {
		c.logger.Debug("GetDailyJournal Fallback Query", zap.String("query", datalog))
	}

	results, err := c.Query(datalog)
	if err == nil {
		if list, ok := results.([]any); ok && len(list) > 0 {
			pageBytes, _ := json.Marshal(list[0])
			var p Page
			if err := json.Unmarshal(pageBytes, &p); err == nil && p.UUID != "" {
				return &p, nil
			}
			// Fallback to raw map
			var raw map[string]any
			if err := json.Unmarshal(pageBytes, &raw); err == nil {
				return raw, nil
			}
		}
	}

	return nil, nil
}


func (c *Client) ListPages() ([]Page, error) {
	// 1. Try getAllPages (more reliable in some environments)
	resp, err := c.Call("logseq.Editor.getAllPages")
	if err == nil && string(resp) != "null" && string(resp) != "[]" {
		var pages []Page
		if err := json.Unmarshal(resp, &pages); err == nil {
			return pages, nil
		}
	}

	// 2. Fallback to Query to find all pages
	datalog := `[:find (pull ?p [*]) :where [?p :block/name]]`
	
	resp, err = c.Call("logseq.DB.q", datalog)
	if err != nil {
		return nil, err
	}

	var rawResult []interface{}
	if err := json.Unmarshal(resp, &rawResult); err != nil {
		return nil, fmt.Errorf("failed to parse pages: %w", err)
	}

	var pages []Page
	for _, item := range rawResult {
		if list, ok := item.([]interface{}); ok && len(list) > 0 {
			pageBytes, _ := json.Marshal(list[0])
			var p Page
			if err := json.Unmarshal(pageBytes, &p); err == nil {
				pages = append(pages, p)
			}
		} else if itemMap, ok := item.(map[string]interface{}); ok {
			pageBytes, _ := json.Marshal(itemMap)
			var p Page
			if err := json.Unmarshal(pageBytes, &p); err == nil {
				pages = append(pages, p)
			}
		}
	}

	if pages == nil {
		return []Page{}, nil
	}
	return pages, nil
}

func (c *Client) ListNamespaces() ([]string, error) {
	namespaces := make(map[string]bool)

	// 1. Direct parent-child query for namespaces
	// Find all pages that are parents of other pages
	datalog := `[:find ?parentName :where [?p :block/name] [?p :block/parent ?parent] [?parent :block/name ?parentName]]`

	if c.logger != nil {
		c.logger.Debug("ListNamespaces Query", zap.String("query", datalog))
	}

	results, err := c.Query(datalog)
	if err == nil {
		if list, ok := results.([]any); ok {
			for _, res := range list {
				var name string
				if s, ok := res.(string); ok {
					name = s
				} else if sl, ok := res.([]any); ok && len(sl) > 0 {
					if s, ok := sl[0].(string); ok {
						name = s
					}
				}

				if name != "" {
					namespaces[name] = true
				}
			}
		}
	}

	list := []string{}
	for ns := range namespaces {
		list = append(list, ns)
	}
	return list, nil
}



