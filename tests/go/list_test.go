package gotests

import (
	"testing"
	"strings"
	"fmt"
	"regexp"
)

// Copy the functions we need to test from models/blog.go
func buildNestedListsTest(content string) string {
	type frame struct { kind string; indent int }
	var stack []frame
	var out strings.Builder
	lines := strings.Split(content, "\n")
	
	// Debug: print input
	fmt.Printf("Input content:\n%s\n", content)
	
	itemUL := regexp.MustCompile(`^([ \t]*)([-*+])\s+(.+)$`)
	itemOL := regexp.MustCompile(`^([ \t]*)(\d+)\.\s+(.+)$`)

	closeTo := func(targetIndent int) {
		for len(stack) > 0 && stack[len(stack)-1].indent > targetIndent {
			f := stack[len(stack)-1]
			out.WriteString("</" + f.kind + ">\n")
			stack = stack[:len(stack)-1]
		}
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		fmt.Printf("Processing line %d: '%s'\n", i, line)
		
		if m := itemUL.FindStringSubmatch(line); m != nil {
			indent := len(m[1])
			text := m[3]
			fmt.Printf("  UL match: indent=%d, text='%s'\n", indent, text)
			
			closeTo(indent)
			if len(stack) > 0 && stack[len(stack)-1].indent == indent && stack[len(stack)-1].kind != "ul" {
				closeTo(indent-1)
			}
			if len(stack) == 0 || stack[len(stack)-1].kind != "ul" || stack[len(stack)-1].indent != indent {
				out.WriteString("<ul>\n"); stack = append(stack, frame{"ul", indent})
				fmt.Printf("    Created new UL at indent %d\n", indent)
			}
			out.WriteString("<li>" + text + "</li>\n")
			continue
		}
		if m := itemOL.FindStringSubmatch(line); m != nil {
			indent := len(m[1])
			text := m[3]
			fmt.Printf("  OL match: indent=%d, text='%s'\n", indent, text)
			
			closeTo(indent)
			if len(stack) > 0 && stack[len(stack)-1].indent == indent && stack[len(stack)-1].kind != "ol" {
				closeTo(indent-1)
			}
			if len(stack) == 0 || stack[len(stack)-1].kind != "ol" || stack[len(stack)-1].indent != indent {
				out.WriteString("<ol>\n"); stack = append(stack, frame{"ol", indent})
				fmt.Printf("    Created new OL at indent %d\n", indent)
			}
			out.WriteString("<li>" + text + "</li>\n")
			continue
		}
		// non-list line: close any open lists and write line
		if line != "" {
			fmt.Printf("  Non-list line, closing all lists\n")
			closeTo(-1)
		}
		out.WriteString(line)
		if i < len(lines)-1 { out.WriteByte('\n') }
	}
	closeTo(-1)

	result := out.String()
	fmt.Printf("Final result:\n%s\n", result)
	return result
}

func TestListStructure(t *testing.T) {
	// Test the exact content from the database
	content := `- Unordered item A
- Unordered item B
  - Nested item B1

1. Ordered item 1
2. Ordered item 2
  1. Nested 2.1`

	result := buildNestedListsTest(content)
	
	// Check that we get the expected structure
	if !strings.Contains(result, "<ul>") {
		t.Error("Expected unordered list not found")
	}
	if !strings.Contains(result, "<ol>") {
		t.Error("Expected ordered list not found")
	}
	
	// Count the number of <ul> opening tags - should be 2 (one main, one nested)
	ulCount := strings.Count(result, "<ul>")
	if ulCount != 2 {
		t.Errorf("Expected 2 <ul> tags, got %d", ulCount)
	}
	
	// Count the number of <ol> opening tags - should be 2 (one main, one nested)  
	olCount := strings.Count(result, "<ol>")
	if olCount != 2 {
		t.Errorf("Expected 2 <ol> tags, got %d", olCount)
	}
}