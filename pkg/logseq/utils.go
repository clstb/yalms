package logseq

import (
	"regexp"
	"strings"
)

// ExtractLinks finds all [[Page Name]] references in content
// Returns a slice of unique page names found
func ExtractLinks(content string) []string {
	return extractLinks(content)
}

func extractLinks(content string) []string {
	// Regex to find [[...]]
	// Handles namespaces like [[A/B]]
	re := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	matches := re.FindAllStringSubmatch(content, -1)
	
	unique := make(map[string]bool)
	var links []string
	
	for _, match := range matches {
		if len(match) > 1 {
			name := strings.TrimSpace(match[1])
			if name != "" && !unique[name] {
				unique[name] = true
				links = append(links, name)
			}
		}
	}
	return links
}

// extractBlockRefs finds all ((uuid)) references in content
// Returns a slice of unique UUIDs found
func extractBlockRefs(content string) []string {
	// Regex to find ((...))
	re := regexp.MustCompile(`\(\(([^)]+)\)\)`)
	matches := re.FindAllStringSubmatch(content, -1)

	unique := make(map[string]bool)
	var refs []string

	for _, match := range matches {
		if len(match) > 1 {
			uuid := strings.TrimSpace(match[1])
			if uuid != "" && !unique[uuid] {
				unique[uuid] = true
				refs = append(refs, uuid)
			}
		}
	}
	return refs
}

// IsJournalName checks if a page name looks like a Logseq journal date
func IsJournalName(name string) bool {
	// YYYY-MM-DD
	re1 := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if re1.MatchString(name) {
		return true
	}
	// YYYY_MM_DD
	re2 := regexp.MustCompile(`^\d{4}_\d{2}_\d{2}$`)
	if re2.MatchString(name) {
		return true
	}
	// Common text formats like "Jan 18th, 2026" or "2026/01/18"
	// Logseq usually uses YYYY-MM-DD internally but let's be safe
	re3 := regexp.MustCompile(`^\d{4}/\d{2}/\d{2}$`)
	if re3.MatchString(name) {
		return true
	}

	return false
}
