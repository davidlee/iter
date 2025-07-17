package flotsam

import (
	"strings"
	"testing"
	
	"gopkg.in/yaml.v3"
)

// parseFrontmatter is a helper function to parse YAML frontmatter for testing
func parseFrontmatter(content string) (map[string]interface{}, string, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 || lines[0] != "---" {
		return nil, content, nil // No frontmatter
	}
	
	// Find the closing ---
	endIndex := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			endIndex = i
			break
		}
	}
	
	if endIndex == -1 {
		return nil, content, nil // No closing ---
	}
	
	// Parse YAML frontmatter
	frontmatterLines := lines[1:endIndex]
	frontmatterText := strings.Join(frontmatterLines, "\n")
	
	var frontmatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontmatterText), &frontmatter); err != nil {
		return nil, content, err
	}
	
	// Get body content
	bodyLines := lines[endIndex+1:]
	body := strings.Join(bodyLines, "\n")
	
	return frontmatter, body, nil
}

func TestZKViceInteropFrontmatter(t *testing.T) {
	// Test frontmatter with Vice extensions
	content := `---
id: test
title: Test Note with Vice Extensions
created-at: "2025-07-17 10:00:00"
tags: [test, interop]
vice:
  srs:
    easiness: 2.5
    consecutive_correct: 0
    due: 1640995200
    total_reviews: 0
  context: "default"
  flotsam_type: "idea"
---
# Test Note with Vice Extensions

This is a test note with Vice-specific frontmatter extensions.

The [[test link]] should work in both ZK and Vice.`

	// Parse the content
	frontmatter, body, err := parseFrontmatter(content)
	if err != nil {
		t.Fatalf("Failed to parse note content: %v", err)
	}

	// Verify ZK-standard fields are parsed correctly
	if frontmatter["id"] != "test" {
		t.Errorf("Expected id 'test', got %v", frontmatter["id"])
	}
	
	if frontmatter["title"] != "Test Note with Vice Extensions" {
		t.Errorf("Expected title 'Test Note with Vice Extensions', got %v", frontmatter["title"])
	}
	
	if frontmatter["created-at"] != "2025-07-17 10:00:00" {
		t.Errorf("Expected created-at '2025-07-17 10:00:00', got %v", frontmatter["created-at"])
	}
	
	// Verify tags are parsed correctly
	tags, ok := frontmatter["tags"].([]interface{})
	if !ok {
		t.Errorf("Expected tags to be []interface{}, got %T", frontmatter["tags"])
	} else {
		if len(tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(tags))
		}
		if tags[0] != "test" || tags[1] != "interop" {
			t.Errorf("Expected tags [test, interop], got %v", tags)
		}
	}
	
	// Verify Vice extensions are preserved
	vice, ok := frontmatter["vice"].(map[string]interface{})
	if !ok {
		t.Errorf("Expected vice to be map[string]interface{}, got %T", frontmatter["vice"])
	} else {
		// Check SRS data
		srs, ok := vice["srs"].(map[string]interface{})
		if !ok {
			t.Errorf("Expected vice.srs to be map[string]interface{}, got %T", vice["srs"])
		} else {
			if srs["easiness"] != 2.5 {
				t.Errorf("Expected easiness 2.5, got %v", srs["easiness"])
			}
			if srs["consecutive_correct"] != 0 {
				t.Errorf("Expected consecutive_correct 0, got %v", srs["consecutive_correct"])
			}
		}
		
		// Check context
		if vice["context"] != "default" {
			t.Errorf("Expected context 'default', got %v", vice["context"])
		}
		
		// Check flotsam_type
		if vice["flotsam_type"] != "idea" {
			t.Errorf("Expected flotsam_type 'idea', got %v", vice["flotsam_type"])
		}
	}
	
	// Verify body content
	expectedBody := `# Test Note with Vice Extensions

This is a test note with Vice-specific frontmatter extensions.

The [[test link]] should work in both ZK and Vice.`
	
	if body != expectedBody {
		t.Errorf("Expected body:\n%s\n\nGot:\n%s", expectedBody, body)
	}
	
	// Test link extraction works on the body
	links := ExtractLinks(body)
	if len(links) != 1 {
		t.Errorf("Expected 1 link, got %d", len(links))
	} else {
		link := links[0]
		if link.Href != "test link" {
			t.Errorf("Expected link href 'test link', got %v", link.Href)
		}
		if link.Type != LinkTypeWikiLink {
			t.Errorf("Expected link type WikiLink, got %v", link.Type)
		}
	}
}

func TestZKBackwardCompatibility(t *testing.T) {
	// Test standard ZK frontmatter without Vice extensions
	content := `---
id: jgtt
title: git
created-at: "2025-06-23 14:06:42"
tags: [draft, 'to/review', tech, versioning]
---
# git

[[jujutsu]]`

	// Parse the content
	frontmatter, body, err := parseFrontmatter(content)
	if err != nil {
		t.Fatalf("Failed to parse note content: %v", err)
	}

	// Verify standard ZK fields are parsed correctly
	if frontmatter["id"] != "jgtt" {
		t.Errorf("Expected id 'jgtt', got %v", frontmatter["id"])
	}
	
	if frontmatter["title"] != "git" {
		t.Errorf("Expected title 'git', got %v", frontmatter["title"])
	}
	
	// Verify no Vice extensions are present
	_, hasVice := frontmatter["vice"]
	if hasVice {
		t.Error("Expected no 'vice' key in standard ZK frontmatter")
	}
	
	// Verify link extraction works
	links := ExtractLinks(body)
	if len(links) != 1 {
		t.Errorf("Expected 1 link, got %d", len(links))
	} else {
		link := links[0]
		if link.Href != "jujutsu" {
			t.Errorf("Expected link href 'jujutsu', got %v", link.Href)
		}
	}
}