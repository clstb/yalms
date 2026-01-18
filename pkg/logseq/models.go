package logseq

import "encoding/json"

// Block represents a Logseq block
type Block struct {
	ID         int                    `json:"id"`
	UUID       string                 `json:"uuid"`
	Content    string                 `json:"content"`
	Format     string                 `json:"format"`
	Left       any            `json:"left"` // minimal representation of ID/UUID or just ID
	Parent     EntityRef              `json:"parent"` // Can be map or ID
	Page       EntityRef              `json:"page"`   // Can be map or ID
	Properties map[string]any `json:"properties,omitempty"`
	Children   []any          `json:"children,omitempty"` // Can be blocks or uuids depending on depth
	Refs       []any          `json:"refs,omitempty"`     // References (pages/blocks)
}

// EntityRef represents a reference to a page or block (ID or object with UUID)
type EntityRef struct {
	ID   int    `json:"id"`
	UUID string `json:"uuid"`
}

func (e *EntityRef) UnmarshalJSON(data []byte) error {
	// Try parsing as integer ID first
	var id int
	if err := json.Unmarshal(data, &id); err == nil {
		e.ID = id
		return nil
	}

	// Try parsing as object with ID/UUID
	type Alias EntityRef
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	return json.Unmarshal(data, &aux)
}

// Page represents a Logseq page
type Page struct {
	ID           int                    `json:"id"`
	UUID         string                 `json:"uuid"`
	Name         string                 `json:"name"`
	OriginalName string                 `json:"originalName"`
	Properties   map[string]any `json:"properties,omitempty"` // Explicit properties field if returned
	Journal      bool           `json:"journal?"`

	// Capture all other fields as properties
	RawProperties map[string]any `json:"-"`
}

func (p *Page) UnmarshalJSON(data []byte) error {
	type Alias Page
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Capture explicit properties unmarshaled by json package before we potentially overwrite them
	explicitProps := p.Properties

	// Capture raw map to find extra fields
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Filter out known fields to populate Properties
	p.Properties = make(map[string]any)
	knownFields := map[string]bool{
		"id": true, "uuid": true, "name": true, "originalName": true, "journal?": true,
		"properties": true, "left": true, "parent": true, "page": true, "format": true,
		"children": true, "content": true, "createdAt": true, "updatedAt": true,
		"file": true, "namespace": true,
	}

	for k, v := range raw {
		if !knownFields[k] {
			p.Properties[k] = v
		}
	}

	// Merge explicit properties if they existed
	if explicitProps != nil {
		for k, v := range explicitProps {
			p.Properties[k] = v
		}
	}

	return nil
}

// GraphInfo represents basic graph information
type GraphInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// BlockContent represents a block and its optional children for batch creation
type BlockContent struct {
	Content    string         `json:"content"`
	Properties map[string]any `json:"properties,omitempty"`
	Children   []BlockContent `json:"children,omitempty"`
}
