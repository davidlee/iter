// Copyright 2024 The zk-org Authors
// Copyright 2024 David Holsgrove
//
// This file contains code copied and adapted from zk (https://github.com/zk-org/zk)
// Original source: internal/adapter/markdown/markdown.go, internal/adapter/markdown/extensions/wikilink.go
// Licensed under GPLv3
//
// SPDX-License-Identifier: GPL-3.0-only

// Package flotsam provides ZK-compatible parsing and linking functionality for flotsam notes.
// This package contains code copied and adapted from zk (https://github.com/zk-org/zk)
// to provide ZK-compatible frontmatter parsing and wiki-link extraction.
package flotsam

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// LinkType represents the kind of link.
type LinkType int

// LinkType constants define the different types of links that can be extracted.
const (
	LinkTypeMarkdown LinkType = iota // Standard markdown [text](url) links
	LinkTypeWikiLink                 // Wiki-style [[target]] links
	LinkTypeImplicit                 // Auto-detected URLs
)

// LinkRelation defines the relationship between a link's source and target.
type LinkRelation string

const (
	// LinkRelationDown defines the target note as a child of the source.
	LinkRelationDown LinkRelation = "down"
	// LinkRelationUp defines the target note as a parent of the source.
	LinkRelationUp LinkRelation = "up"
)

// Link represents a link found in note content.
// Adapted from zk's Link structure for flotsam use.
type Link struct {
	Title        string
	Href         string
	Type         LinkType
	IsExternal   bool
	Rels         []LinkRelation
	Snippet      string
	SnippetStart int
	SnippetEnd   int
}

// WikiLink represents a wiki link found in a Markdown document.
// Copied from zk's WikiLink AST node.
type WikiLink struct {
	ast.Link
}

// wikiLinkParser parses wiki links in markdown content.
// Adapted from zk's wikilink extension.
type wikiLinkParser struct{}

func (p *wikiLinkParser) Trigger() []byte {
	return []byte{'[', '#'}
}

func (p *wikiLinkParser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, _ := block.PeekLine()

	var (
		href  string
		label string
		rel   LinkRelation
	)

	var (
		opened          = false // Found at least [[
		closed          = false // Found at least ]]
		escaping        = false // Found a backslash, next character will be literal
		parsingLabel    = false // Found a | in a Wikilink, now we parse the link's label
		openerCharCount = 0     // Number of [ encountered
		closerCharCount = 0     // Number of ] encountered
		endPos          = 0     // Last position of the link in the line
	)

	appendRune := func(c rune) {
		if parsingLabel {
			label += string(c)
		} else {
			href += string(c)
		}
	}

	for i, char := range string(line) {
		endPos = i

		if closed {
			// Supports trailing hash syntax for Neuron's Folgezettel, e.g. [[id]]#
			if char == '#' {
				rel = LinkRelationDown
			}
			break
		}

		if !opened {
			switch char {
			// Supports leading hash syntax for Neuron's Folgezettel, e.g. #[[id]]
			case '#':
				rel = LinkRelationUp
				continue
			case '[':
				openerCharCount++
				continue
			}

			if openerCharCount < 2 || openerCharCount > 3 {
				return nil
			}
		}
		opened = true

		if !escaping {
			switch char {

			case '|': // [[href | label]]
				parsingLabel = true
				continue

			case '\\':
				escaping = true
				continue

			case ']':
				closerCharCount++
				if closerCharCount == openerCharCount {
					closed = true
					// Neuron's legacy [[[Folgezettel]]].
					if closerCharCount == 3 {
						rel = LinkRelationDown
					}
				}
				continue
			}
		}
		escaping = false

		// Found incomplete number of closing brackets to close the link.
		// We add them to the HREF and reset the count.
		if closerCharCount > 0 {
			for i := 0; i < closerCharCount; i++ {
				appendRune(']')
			}
			closerCharCount = 0
		}
		appendRune(char)
	}

	if !closed || len(href) == 0 {
		return nil
	}

	block.Advance(endPos)

	href = strings.TrimSpace(href)
	label = strings.TrimSpace(label)
	if len(label) == 0 {
		label = href
	}

	link := &WikiLink{Link: *ast.NewLink()}
	link.Destination = []byte(href)
	// Title will be parsed as the link's rel by the Markdown parser.
	link.Title = []byte(rel)
	link.AppendChild(link, ast.NewString([]byte(label)))

	return link
}

// WikiLinkExtension adds support for parsing wiki links.
var WikiLinkExtension = &wikiLinkExt{}

type wikiLinkExt struct{}

func (w *wikiLinkExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&wikiLinkParser{}, 199),
		),
	)
}

// LinkExtractor extracts links from markdown content using goldmark AST parsing.
// This is adapted from zk's parseLinks function.
// AIDEV-NOTE: supports ZK link types: wikilinks [[target]], uplinks #[[target]], downlinks [[target]]#, legacy [[[target]]]
type LinkExtractor struct {
	md goldmark.Markdown
}

// NewLinkExtractor creates a new link extractor with goldmark configured for wiki links.
func NewLinkExtractor() *LinkExtractor {
	return &LinkExtractor{
		md: goldmark.New(
			goldmark.WithExtensions(
				extension.NewLinkify(),
				WikiLinkExtension,
			),
		),
	}
}

// ExtractLinks extracts all links from markdown content using AST parsing.
// This is much more robust than regex-based extraction.
func (le *LinkExtractor) ExtractLinks(content string) ([]Link, error) {
	bytes := []byte(content)

	context := parser.NewContext()
	root := le.md.Parser().Parse(
		text.NewReader(bytes),
		parser.WithContext(context),
	)

	return le.parseLinks(root, bytes)
}

// parseLinks extracts outbound links from the AST.
// Adapted from zk's parseLinks function.
func (le *LinkExtractor) parseLinks(root ast.Node, source []byte) ([]Link, error) {
	links := make([]Link, 0)

	err := ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch link := n.(type) {
			case *ast.Link:
				href, err := url.PathUnescape(string(link.Destination))
				if err != nil {
					href = string(link.Destination) // fallback to original if unescape fails
				}
				if href != "" {
					snippet, snStart, snEnd := le.extractLines(n, source)
					links = append(links, Link{
						Title:        extractLinkText(link, source),
						Href:         href,
						Type:         LinkTypeMarkdown,
						Rels:         parseRels(strings.Fields(string(link.Title))...),
						IsExternal:   IsURL(href),
						Snippet:      snippet,
						SnippetStart: snStart,
						SnippetEnd:   snEnd,
					})
				}

			case *ast.AutoLink:
				if href := string(link.URL(source)); href != "" && link.AutoLinkType == ast.AutoLinkURL {
					snippet, snStart, snEnd := le.extractLines(n, source)
					links = append(links, Link{
						Title:        string(link.Label(source)),
						Href:         href,
						Type:         LinkTypeImplicit,
						Rels:         []LinkRelation{},
						IsExternal:   true,
						Snippet:      snippet,
						SnippetStart: snStart,
						SnippetEnd:   snEnd,
					})
				}

			case *WikiLink:
				href := string(link.Destination)
				if href != "" {
					snippet, snStart, snEnd := le.extractLines(n, source)
					links = append(links, Link{
						Title:        extractLinkText(link, source),
						Href:         href,
						Type:         LinkTypeWikiLink,
						Rels:         parseRels(strings.Fields(string(link.Title))...),
						IsExternal:   IsURL(href),
						Snippet:      snippet,
						SnippetStart: snStart,
						SnippetEnd:   snEnd,
					})
				}
			}
		}
		return ast.WalkContinue, nil
	})
	return links, err
}

// extractLinkText extracts text content from a goldmark AST node by walking its children.
// This replaces the deprecated Text() method.
// AIDEV-NOTE: goldmark Text() deprecated - use manual AST traversal for text extraction
func extractLinkText(node ast.Node, source []byte) string {
	var textBuffer strings.Builder

	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch textNode := n.(type) {
		case *ast.Text:
			textBuffer.Write(textNode.Value(source))
		case *ast.String:
			textBuffer.Write(textNode.Value)
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return ""
	}

	return textBuffer.String()
}

// extractLines extracts the paragraph or line context around a link.
// Copied from zk's extractLines function.
func (le *LinkExtractor) extractLines(n ast.Node, source []byte) (content string, start, end int) {
	if n == nil {
		return
	}
	switch n.Type() {
	case ast.TypeInline:
		return le.extractLines(n.Parent(), source)

	case ast.TypeBlock:
		segs := n.Lines()
		if segs.Len() == 0 {
			return
		}
		start = segs.At(0).Start
		end = segs.At(segs.Len() - 1).Stop
		content = string(source[start:end])
	}

	return
}

// parseRels creates a slice of LinkRelation from a list of strings.
// Copied from zk's LinkRels function.
func parseRels(rel ...string) []LinkRelation {
	rels := []LinkRelation{}
	for _, r := range rel {
		rels = append(rels, LinkRelation(r))
	}
	return rels
}

// Global link extractor instance
var globalExtractor = NewLinkExtractor()

// ExtractLinks is a convenience function that uses the global extractor.
func ExtractLinks(content string) []Link {
	links, err := globalExtractor.ExtractLinks(content)
	if err != nil {
		return []Link{} // Return empty slice on error
	}
	return links
}

// ExtractWikiLinkTargets extracts just the target hrefs from wikilinks.
// This is useful for building backlink indexes.
func ExtractWikiLinkTargets(content string) []string {
	links := ExtractLinks(content)
	targets := make([]string, 0)

	for _, link := range links {
		if link.Type == LinkTypeWikiLink && !link.IsExternal {
			targets = append(targets, link.Href)
		}
	}

	return targets
}

// BuildBacklinkIndex builds a map of note targets to their source notes.
// This is used for context-scoped backlink computation.
func BuildBacklinkIndex(notes map[string]string) map[string][]string {
	backlinks := make(map[string][]string)

	for noteID, content := range notes {
		targets := ExtractWikiLinkTargets(content)

		for _, target := range targets {
			if backlinks[target] == nil {
				backlinks[target] = []string{}
			}
			backlinks[target] = append(backlinks[target], noteID)
		}
	}

	return backlinks
}

// RemoveDuplicateLinks removes duplicate links from a slice.
// Two links are considered duplicates if they have the same href, title, and type.
func RemoveDuplicateLinks(links []Link) []Link {
	seen := make(map[string]bool)
	result := []Link{}

	for _, link := range links {
		key := link.Href + "|" + link.Title + "|" + string(rune(link.Type))
		if !seen[key] {
			seen[key] = true
			result = append(result, link)
		}
	}

	return result
}

// Simple regex fallback for basic cases (keeping for compatibility)
var simpleWikiLinkRegex = regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)

// ExtractWikiLinksSimple provides a simple regex-based fallback for basic wiki link extraction.
func ExtractWikiLinksSimple(content string) []string {
	matches := simpleWikiLinkRegex.FindAllStringSubmatch(content, -1)
	targets := make([]string, 0, len(matches))

	for _, match := range matches {
		if len(match) >= 2 && match[1] != "" {
			targets = append(targets, strings.TrimSpace(match[1]))
		}
	}

	return targets
}
