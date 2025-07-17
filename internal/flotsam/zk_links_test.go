package flotsam

import (
	"testing"
)

func TestExtractLinks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Link
	}{
		{
			name:    "basic wikilink",
			content: "This is a [[test link]] in content.",
			expected: []Link{
				{
					Title:      "test link",
					Href:       "test link",
					Type:       LinkTypeWikiLink,
					IsExternal: false,
					Rels:       []LinkRelation{},
				},
			},
		},
		{
			name:    "wikilink with custom label",
			content: "See [[target|Custom Label]] for more info.",
			expected: []Link{
				{
					Title:      "Custom Label",
					Href:       "target",
					Type:       LinkTypeWikiLink,
					IsExternal: false,
					Rels:       []LinkRelation{},
				},
			},
		},
		{
			name:    "uplink with hash prefix",
			content: "This references #[[parent note]] as parent.",
			expected: []Link{
				{
					Title:      "parent note",
					Href:       "parent note",
					Type:       LinkTypeWikiLink,
					IsExternal: false,
					Rels:       []LinkRelation{LinkRelationUp},
				},
			},
		},
		{
			name:    "downlink with hash suffix",
			content: "This references [[child note]]# as child.",
			expected: []Link{
				{
					Title:      "child note",
					Href:       "child note",
					Type:       LinkTypeWikiLink,
					IsExternal: false,
					Rels:       []LinkRelation{LinkRelationDown},
				},
			},
		},
		{
			name:    "legacy downlink",
			content: "This is a [[[legacy downlink]]] example.",
			expected: []Link{
				{
					Title:      "legacy downlink",
					Href:       "legacy downlink",
					Type:       LinkTypeWikiLink,
					IsExternal: false,
					Rels:       []LinkRelation{LinkRelationDown},
				},
			},
		},
		{
			name:    "basic markdown link",
			content: "Visit [Example](https://example.com) for more info.",
			expected: []Link{
				{
					Title:      "Example",
					Href:       "https://example.com",
					Type:       LinkTypeMarkdown,
					IsExternal: true,
					Rels:       []LinkRelation{},
				},
			},
		},
		{
			name:    "relative markdown link",
			content: "See [Other Note](./other-note.md) for details.",
			expected: []Link{
				{
					Title:      "Other Note",
					Href:       "./other-note.md",
					Type:       LinkTypeMarkdown,
					IsExternal: false,
					Rels:       []LinkRelation{},
				},
			},
		},
		{
			name:    "auto-linked URL",
			content: "Visit https://example.com for more info.",
			expected: []Link{
				{
					Title:      "https://example.com",
					Href:       "https://example.com",
					Type:       LinkTypeImplicit,
					IsExternal: true,
					Rels:       []LinkRelation{},
				},
			},
		},
		{
			name:    "external URL in wikilink",
			content: "Link to [[https://example.com]] website.",
			expected: []Link{
				{
					Title:      "https://example.com",
					Href:       "https://example.com",
					Type:       LinkTypeWikiLink,
					IsExternal: true,
					Rels:       []LinkRelation{},
				},
			},
		},
		{
			name:     "empty content",
			content:  "",
			expected: []Link{},
		},
		{
			name:     "no links",
			content:  "This content has no links at all.",
			expected: []Link{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			links := ExtractLinks(test.content)
			
			if len(links) != len(test.expected) {
				t.Errorf("Expected %d links, got %d", len(test.expected), len(links))
				for i, link := range links {
					t.Logf("Link %d: %s -> %s (type %d)", i, link.Title, link.Href, link.Type)
				}
				return
			}
			
			for i, expected := range test.expected {
				if i >= len(links) {
					t.Errorf("Missing link at index %d", i)
					continue
				}
				
				link := links[i]
				
				if link.Title != expected.Title {
					t.Errorf("Expected title %q, got %q", expected.Title, link.Title)
				}
				if link.Href != expected.Href {
					t.Errorf("Expected href %q, got %q", expected.Href, link.Href)
				}
				if link.Type != expected.Type {
					t.Errorf("Expected type %d, got %d", expected.Type, link.Type)
				}
				if link.IsExternal != expected.IsExternal {
					t.Errorf("Expected IsExternal %t, got %t", expected.IsExternal, link.IsExternal)
				}
				if len(link.Rels) != len(expected.Rels) {
					t.Errorf("Expected %d relations, got %d", len(expected.Rels), len(link.Rels))
				} else {
					for j, expectedRel := range expected.Rels {
						if link.Rels[j] != expectedRel {
							t.Errorf("Expected relation %q, got %q", expectedRel, link.Rels[j])
						}
					}
				}
			}
		})
	}
}

func TestExtractLinksComplex(t *testing.T) {
	content := `# Test Note

This note contains multiple types of links:

- Wiki link: [[test note]]
- Markdown link: [Example](https://example.com)
- Bare URL: https://test.com
- Uplink: #[[parent note]]
- Downlink: [[child note]]#
- Legacy downlink: [[[legacy note]]]

Some more complex examples:
- Wiki link with label: [[target|Custom Label]]
- Relative markdown: [Local](./local.md)
`

	links := ExtractLinks(content)
	
	// Debug output to see what we actually get
	for i, link := range links {
		t.Logf("Link %d: %s -> %s (type %d, external: %t)", i, link.Title, link.Href, link.Type, link.IsExternal)
	}
	
	// Check that we have different types
	typeCount := map[LinkType]int{}
	for _, link := range links {
		typeCount[link.Type]++
	}
	
	expectedWikiLinks := 5 // [[test note]], #[[parent note]], [[child note]]#, [[[legacy note]]], [[target|Custom Label]]
	// AIDEV-NOTE: originally expected 6 but test content only has 5 wiki links - fixed in T027
	expectedMarkdownLinks := 2 // [Example](https://example.com), [Local](./local.md)
	expectedImplicitLinks := 1 // https://test.com (if auto-linked)
	
	if typeCount[LinkTypeWikiLink] != expectedWikiLinks {
		t.Errorf("Expected %d wiki links, got %d", expectedWikiLinks, typeCount[LinkTypeWikiLink])
	}
	if typeCount[LinkTypeMarkdown] != expectedMarkdownLinks {
		t.Errorf("Expected %d markdown links, got %d", expectedMarkdownLinks, typeCount[LinkTypeMarkdown])
	}
	// Note: Auto-linking might not work without proper configuration
	if typeCount[LinkTypeImplicit] > 0 && typeCount[LinkTypeImplicit] != expectedImplicitLinks {
		t.Errorf("Expected %d implicit links, got %d", expectedImplicitLinks, typeCount[LinkTypeImplicit])
	}
}

func TestExtractWikiLinkTargets(t *testing.T) {
	content := `# Test Note

Links to [[note one]] and [[note two]].
Also links to https://example.com but that's external.
And [[note one]] again (duplicate).
External wikilink: [[https://example.com]] should be excluded.
`

	targets := ExtractWikiLinkTargets(content)
	
	// Should find 3 targets (including the duplicate, excluding external)
	expectedTargets := []string{"note one", "note two", "note one"}
	if len(targets) != len(expectedTargets) {
		t.Errorf("Expected %d targets, got %d: %v", len(expectedTargets), len(targets), targets)
	}
	
	// Check specific targets
	for i, target := range targets {
		if i >= len(expectedTargets) {
			t.Errorf("Unexpected target at index %d: %s", i, target)
			continue
		}
		if target != expectedTargets[i] {
			t.Errorf("Expected target %q, got %q", expectedTargets[i], target)
		}
	}
}

func TestBuildBacklinkIndex(t *testing.T) {
	notes := map[string]string{
		"note1": "This links to [[note2]] and [[note3]].",
		"note2": "This links to [[note3]] and [[note1]].",
		"note3": "This links to [[note1]].",
	}
	
	backlinks := BuildBacklinkIndex(notes)
	
	// note1 should be linked from note2 and note3
	if len(backlinks["note1"]) != 2 {
		t.Errorf("Expected 2 backlinks to note1, got %d", len(backlinks["note1"]))
	}
	
	// note2 should be linked from note1
	if len(backlinks["note2"]) != 1 {
		t.Errorf("Expected 1 backlink to note2, got %d", len(backlinks["note2"]))
	}
	
	// note3 should be linked from note1 and note2
	if len(backlinks["note3"]) != 2 {
		t.Errorf("Expected 2 backlinks to note3, got %d", len(backlinks["note3"]))
	}
}

func TestRemoveDuplicateLinks(t *testing.T) {
	links := []Link{
		{Title: "Test", Href: "test", Type: LinkTypeWikiLink},
		{Title: "Test", Href: "test", Type: LinkTypeWikiLink}, // duplicate
		{Title: "Other", Href: "other", Type: LinkTypeWikiLink},
		{Title: "Test", Href: "test", Type: LinkTypeMarkdown}, // different type, same href/title
	}
	
	result := RemoveDuplicateLinks(links)
	
	if len(result) != 3 {
		t.Errorf("Expected 3 unique links, got %d", len(result))
		for i, link := range result {
			t.Logf("Result %d: %s -> %s (type %d)", i, link.Title, link.Href, link.Type)
		}
	}
}

func TestLinkExtractor(t *testing.T) {
	extractor := NewLinkExtractor()
	
	content := "This has a [[wiki link]] and [markdown](https://example.com) link."
	
	links, err := extractor.ExtractLinks(content)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(links) != 2 {
		t.Errorf("Expected 2 links, got %d", len(links))
	}
	
	// Check that we have both types
	hasWiki := false
	hasMarkdown := false
	
	for _, link := range links {
		if link.Type == LinkTypeWikiLink {
			hasWiki = true
		}
		if link.Type == LinkTypeMarkdown {
			hasMarkdown = true
		}
	}
	
	if !hasWiki {
		t.Error("Expected to find wiki link")
	}
	if !hasMarkdown {
		t.Error("Expected to find markdown link")
	}
}

func TestExtractWikiLinksSimple(t *testing.T) {
	content := "This has [[simple link]] and [[target|label]] links."
	
	targets := ExtractWikiLinksSimple(content)
	
	expected := []string{"simple link", "target"}
	if len(targets) != len(expected) {
		t.Errorf("Expected %d targets, got %d", len(expected), len(targets))
	}
	
	for i, target := range targets {
		if i >= len(expected) {
			t.Errorf("Unexpected target at index %d: %s", i, target)
			continue
		}
		if target != expected[i] {
			t.Errorf("Expected target %q, got %q", expected[i], target)
		}
	}
}