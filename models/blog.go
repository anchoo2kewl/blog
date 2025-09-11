package models

import (
	"database/sql"
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"time"

	"github.com/russross/blackfriday/v2"
)

type BlogService struct {
	DB *sql.DB
}

func (bs *BlogService) GetBlogPostBySlug(slug string) (*Post, error) {
	post := Post{}
	fmt.Println("Fetching blog post with slug:", slug)

	query := `SELECT * FROM posts WHERE slug = $1 LIMIT 1`
	rows, err := bs.DB.Query(query, slug)
	if err != nil {
		return &post, nil
	}

	for rows.Next() {
		err := rows.Scan(&post.ID, &post.UserID, &post.CategoryID, &post.Title, &post.Content, &post.Slug, &post.PublicationDate, &post.LastEditDate, &post.IsPublished, &post.FeaturedImageURL, &post.CreatedAt)
		if err != nil {
			panic(err)
		}

		// Display-friendly dates
		if post.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, post.CreatedAt); err == nil {
				post.CreatedAt = t.Format("January 2, 2006")
			}
		}
		if post.PublicationDate != "" {
			if pt, err := time.Parse(time.RFC3339, post.PublicationDate); err == nil {
				post.PublicationDate = pt.Format("January 2, 2006")
			}
		} else {
			post.PublicationDate = post.CreatedAt
		}
		if post.LastEditDate != "" {
			if lt, err := time.Parse(time.RFC3339, post.LastEditDate); err == nil {
				post.LastEditDate = lt.Format("January 2, 2006")
			}
		}

		// Markdown render (blackfriday passes raw HTML through)
		content := replaceMoreTag(post.Content)

		content = stripStyleSnippets(content)
		content = preprocessLooseMarkdownHTML(content)
		content = normalizeInlinePipeTables(content)
		// Convert Markdown-style fenced blocks to HTML before markdown render,
		// because editor content may mix HTML and backticks.
		content = convertFences(content)
		// Let blackfriday process the complete markdown including lists and links
		htmlOut := renderMarkdown(content)
		htmlOut = replaceBlockquoteTag(replacelistTag(htmlOut))
		// Convert leftover **bold** and *italic* that might still exist inside HTML nodes
		htmlOut = convertInlineEmphasisInHTML(htmlOut)
		// Ensure image galleries have lightbox links even if editor inserted bare <img>
		htmlOut = wrapImageGalleries(htmlOut)
		post.ContentHTML = template.HTML(htmlOut)
	}

	if err != nil {
		return nil, fmt.Errorf("Post could not be fetched: %w", err)
	}
	return &post, nil
}

func replaceMoreTag(content string) string {
	// Remove both literal and escaped markers
	candidates := []string{"<more-->", "&lt;more--&gt;"}
	for _, mk := range candidates {
		if idx := strings.Index(content, mk); idx != -1 {
			before := content[:idx]
			after := content[idx+len(mk):]
			content = before + after
		}
	}
	return content
}

func replacelistTag(content string) string {
	// Add utility classes to all ul/ol/li that don't already have a class
	ulRe := regexp.MustCompile(`(?i)<ul[^>]*>`)
	olRe := regexp.MustCompile(`(?i)<ol[^>]*>`)
	liRe := regexp.MustCompile(`(?i)<li[^>]*>`)

	content = ulRe.ReplaceAllStringFunc(content, func(m string) string {
		if strings.Contains(strings.ToLower(m), "class=") {
			return m
		}
		return strings.Replace(m, "<ul", `<ul class="list-disc pl-2"`, 1)
	})
	content = olRe.ReplaceAllStringFunc(content, func(m string) string {
		if strings.Contains(strings.ToLower(m), "class=") {
			return m
		}
		return strings.Replace(m, "<ol", `<ol class="list-decimal pl-2"`, 1)
	})
	content = liRe.ReplaceAllStringFunc(content, func(m string) string {
		if strings.Contains(strings.ToLower(m), "class=") {
			return m
		}
		return strings.Replace(m, "<li", `<li class="mb-2"`, 1)
	})

	return content
}

func replaceBlockquoteTag(content string) string {
	// Add classes to all blockquotes without a class
	bqRe := regexp.MustCompile(`(?i)<blockquote[^>]*>`)
	return bqRe.ReplaceAllStringFunc(content, func(m string) string {
		if strings.Contains(strings.ToLower(m), "class=") {
			return m
		}
		return strings.Replace(m, "<blockquote", `<blockquote class="p-4 my-4 border-s-4 border-gray-300 bg-gray-50 dark:border-gray-500 dark:bg-gray-800"`, 1)
	})
}

// Remove single-line inline CSS snippets users may paste intending to style code blocks globally.
// We hide those instead of rendering them in the article.
func stripStyleSnippets(content string) string {
	// Protect fenced code and <pre> blocks so example CSS is preserved
	codeRe := regexp.MustCompile("(?is)(```[\\s\\S]*?```|<pre[\\s\\S]*?</pre>)")
	placeholders := []string{}
	content = codeRe.ReplaceAllStringFunc(content, func(m string) string {
		placeholders = append(placeholders, m)
		return fmt.Sprintf("[[[STYLE_PROTECT_%d]]]", len(placeholders)-1)
	})
	// Remove simple one-line CSS pasted outside code blocks
	rePre := regexp.MustCompile(`(?m)^\s*pre\s*code\s*\{[^}]*\}\s*$`)
	content = rePre.ReplaceAllString(content, "")
	reToken := regexp.MustCompile(`(?m)^\s*\.[A-Za-z0-9_-]+(?:\.[A-Za-z0-9_-]+)*\s*\{[^}]*\}\s*$`)
	content = reToken.ReplaceAllString(content, "")
	// Restore code
	for i, m := range placeholders {
		content = strings.ReplaceAll(content, fmt.Sprintf("[[[STYLE_PROTECT_%d]]]", i), m)
	}
	return content
}

// RenderContent exposes the same server-side pipeline used for previews
func RenderContent(content string) string {
	content = replaceMoreTag(content)
	content = stripStyleSnippets(content)
	content = preprocessLooseMarkdownHTML(content)
	content = normalizeInlinePipeTables(content)
	content = convertFences(content)
	htmlOut := renderMarkdown(content)
	htmlOut = replaceBlockquoteTag(replacelistTag(htmlOut))
	htmlOut = convertInlineEmphasisInHTML(htmlOut)
	return htmlOut
}

// Normalize inline pipe tables that were collapsed into one line of text
func normalizeInlinePipeTables(content string) string {
	// Protect code
	codeRe := regexp.MustCompile("(?is)(```[\\s\\S]*?```|<pre[\\s\\S]*?</pre>)")
	placeholders := []string{}
	content = codeRe.ReplaceAllStringFunc(content, func(m string) string {
		placeholders = append(placeholders, m)
		return fmt.Sprintf("[[[CODE_BLOCK_%d]]]", len(placeholders)-1)
	})

	// Split "| |" into row breaks if para contains many pipes
	paraRe := regexp.MustCompile(`(?is)<p>([\s\S]*?\|[\s\S]*?)</p>`)
	content = paraRe.ReplaceAllStringFunc(content, func(p string) string {
		if strings.Count(p, "|") >= 8 || strings.Contains(p, "---") {
			return strings.ReplaceAll(p, "| |", "|\n|")
		}
		return p
	})
	if strings.Count(content, "| |") >= 2 {
		content = strings.ReplaceAll(content, "| |", "|\n|")
	}

	for i, m := range placeholders {
		content = strings.ReplaceAll(content, fmt.Sprintf("[[[CODE_BLOCK_%d]]]", i), m)
	}
	return content
}

// Process multi-line blockquotes, combining consecutive lines starting with > into single blockquotes
func processBlockquotes(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	var inBlockquote bool
	var blockquoteLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Check if line starts with > or &gt;
		if strings.HasPrefix(trimmed, ">") || strings.HasPrefix(trimmed, "&gt;") {
			// Extract content after the > marker
			var text string
			if strings.HasPrefix(trimmed, "&gt;") {
				text = strings.TrimSpace(trimmed[4:]) // Remove "&gt;"
			} else {
				text = strings.TrimSpace(trimmed[1:]) // Remove ">"
			}

			if !inBlockquote {
				// Start new blockquote
				inBlockquote = true
				blockquoteLines = []string{text}
			} else {
				// Continue existing blockquote
				blockquoteLines = append(blockquoteLines, text)
			}
		} else {
			// End of blockquote if we were in one
			if inBlockquote {
				// Join all blockquote lines and wrap in blockquote tags
				blockquoteContent := strings.Join(blockquoteLines, " ")
				result = append(result, "<blockquote><p>"+blockquoteContent+"</p></blockquote>")
				inBlockquote = false
				blockquoteLines = nil
			}
			// Add the current line
			result = append(result, line)
		}
	}

	// Handle case where content ends with a blockquote
	if inBlockquote {
		blockquoteContent := strings.Join(blockquoteLines, " ")
		result = append(result, "<blockquote><p>"+blockquoteContent+"</p></blockquote>")
	}

	return strings.Join(result, "\n")
}

// Convert simple markdown-like markers that users may have typed inside HTML paragraphs
// generated by the WYSIWYG editor (e.g., <p>## Heading</p>, <p>- item</p>).
func preprocessLooseMarkdownHTML(content string) string {
	// --- Normalize whitespace early so indentation is reliable ---
	// Convert common Unicode spaces to ASCII space
	content = strings.NewReplacer(
		"\u00A0", " ", // NBSP
		"\u2002", " ", // en space
		"\u2003", " ", // em space
		"\u2007", " ", // figure space
		"\u202F", " ", // narrow NBSP
	).Replace(content)
	// Normalize entities and line breaks
	content = strings.ReplaceAll(content, "&nbsp;", " ")
	content = strings.ReplaceAll(content, "&#160;", " ")
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "<br>", "\n")
	content = strings.ReplaceAll(content, "<br/>", "\n")
	content = strings.ReplaceAll(content, "<br />", "\n")

	// If a paragraph contains numbered/bulleted lines, unwrap it so Markdown can parse the list
	paraRe := regexp.MustCompile(`(?is)<p>([\s\S]*?)</p>`)          // match any paragraph
	listLine := regexp.MustCompile(`(?m)^\s*(?:[-*+]\s+|\d+\.\s+)`) // start-of-line list markers
	content = paraRe.ReplaceAllStringFunc(content, func(p string) string {
		inner := paraRe.FindStringSubmatch(p)
		if len(inner) != 2 {
			return p
		}
		body := inner[1]
		// If list markers are inline like ": 1. foo 2. bar", insert newlines before each marker
		inlineNum := regexp.MustCompile(`([:;])\s*(\d+\.\s+)`)
		body = inlineNum.ReplaceAllString(body, "$1\n\n$2")
		inlineBullet := regexp.MustCompile(`([:;])\s*([-*+]\s+)`)
		body = inlineBullet.ReplaceAllString(body, "$1\n\n$2")
		if listLine.MatchString(body) {
			return "\n" + body + "\n" // unwrap into raw lines so renderer sees a list
		}
		return p
	})

	// Ensure markdown resumes after HTML blocks by injecting a blank line after closing tags
	// This helps Blackfriday recognize headings that appear immediately after a gallery or other HTML block
	closeBlock := regexp.MustCompile(`(?is)</(div|figure|section|table|blockquote)>\s*`)
	content = closeBlock.ReplaceAllString(content, "</$1>\n\n")

	// Protect code blocks to avoid altering their contents
	preRe := regexp.MustCompile(`(?is)<pre[\s\S]*?</pre>`) // matches <pre>...</pre>
	placeholders := []string{}
	content = preRe.ReplaceAllStringFunc(content, func(m string) string {
		placeholders = append(placeholders, m)
		return fmt.Sprintf("[[[PRE_BLOCK_%d]]]", len(placeholders)-1)
	})

	// Convert top-level markdown lines too (not only inside <p>)
	// Headings
	reTopH3 := regexp.MustCompile(`(?m)^[ \t]*###[ \t]+(.+)$`)
	content = reTopH3.ReplaceAllString(content, `<h3>$1</h3>`)
	reTopH2 := regexp.MustCompile(`(?m)^[ \t]*##[ \t]+(.+)$`)
	content = reTopH2.ReplaceAllString(content, `<h2>$1</h2>`)
	reTopH1 := regexp.MustCompile(`(?m)^[ \t]*#[ \t]+(.+)$`)
	content = reTopH1.ReplaceAllString(content, `<h1>$1</h1>`)

	// Process multi-line blockquotes starting with > or &gt;
	content = processBlockquotes(content)

	// Paragraph-wrapped headings
	reH3 := regexp.MustCompile(`(?is)<p>\s*###\s+(.+?)\s*</p>`)
	content = reH3.ReplaceAllString(content, `<h3>$1</h3>`)
	reH2 := regexp.MustCompile(`(?is)<p>\s*##\s+(.+?)\s*</p>`)
	content = reH2.ReplaceAllString(content, `<h2>$1</h2>`)
	reH1 := regexp.MustCompile(`(?is)<p>\s*#\s+(.+?)\s*</p>`)
	content = reH1.ReplaceAllString(content, `<h1>$1</h1>`)

	// Paragraph-wrapped list markers -> plain lines so nested builder can work
	rePUL := regexp.MustCompile(`(?is)<p>\s*([ \t]*)([-*+])\s+(.+?)\s*</p>`)
	content = rePUL.ReplaceAllString(content, `$1$2 $3`)
	rePOL := regexp.MustCompile(`(?is)<p>\s*([ \t]*)(\d+)\.\s+(.+?)\s*</p>`)
	content = rePOL.ReplaceAllString(content, `$1$2. $3`)

	// NOTE: Do NOT convert top-level list items to custom tags; that kills indentation.
	// We intentionally removed the rules that produced <ul-li> / <ol-li>.

	// Unwrap paragraphs that contain Markdown emphasis so the Markdown renderer can parse them
	rePEmph := regexp.MustCompile(`(?is)<p>\s*([^<>]*?(\*\*.+?\*\*|__.+?__|\*[^*]+?\*|_[^_]+?_)\s*[^<>]*?)\s*</p>`) // no nested tags
	content = rePEmph.ReplaceAllString(content, "\n$1\n")

	// Horizontal rule (top-level and paragraph-wrapped)
	reTopHR := regexp.MustCompile(`(?m)^[ \t]*---[ \t]*$`)
	content = reTopHR.ReplaceAllString(content, `<hr/>`)
	rePHR := regexp.MustCompile(`(?is)<p>\s*---\s*</p>`)
	content = rePHR.ReplaceAllString(content, `<hr/>`)

	// Also handle headings that appear as raw text immediately inside a container
	// e.g., <div>## Heading</div> → <div><h2>Heading</h2></div>
	reInH3 := regexp.MustCompile(`(?is)>(\s*###\s+)(.+?)\s*<`)
	content = reInH3.ReplaceAllString(content, `><h3>$2</h3><`)
	reInH2 := regexp.MustCompile(`(?is)>(\s*##\s+)(.+?)\s*<`)
	content = reInH2.ReplaceAllString(content, `><h2>$2</h2><`)
	reInH1 := regexp.MustCompile(`(?is)>(\s*#\s+)(.+?)\s*<`)
	content = reInH1.ReplaceAllString(content, `><h1>$2</h1><`)

	// Restore code blocks
	for i, m := range placeholders {
		content = strings.ReplaceAll(content, fmt.Sprintf("[[[PRE_BLOCK_%d]]]", i), m)
	}
	return content
}

// Build nested UL/OL lists based on indentation in raw lines.
// Nests when indentation reaches 2 spaces OR 1 tab per level.
func buildNestedLists(content string) string {
	// Protect code blocks
	preRe := regexp.MustCompile(`(?is)<pre[\s\S]*?</pre>`)
	placeholders := []string{}
	content = preRe.ReplaceAllStringFunc(content, func(m string) string {
		placeholders = append(placeholders, m)
		return fmt.Sprintf("[[[PRE_BLOCK_%d]]]", len(placeholders)-1)
	})

	lines := strings.Split(content, "\n")

	// Patterns for list items
	itemUL := regexp.MustCompile(`^([ \t]*)([-*+])\s+(.+)$`)
	itemOL := regexp.MustCompile(`^([ \t]*)(\d+)\.\s+(.+)$`)

	// Convert leading whitespace to a discrete nesting level.
	// Treat 1 tab == 2 spaces; 2 spaces per level.
	levelFromWS := func(ws string) int {
		tabs, spaces := 0, 0
		for i := 0; i < len(ws); i++ {
			if ws[i] == '\t' {
				tabs++
			} else if ws[i] == ' ' {
				spaces++
			}
		}
		units := tabs*2 + spaces // 1 tab == 2 spaces
		return units / 2         // 2 spaces per level
	}

	type frame struct {
		kind  string // "ul" or "ol"
		level int
	}
	var stack []frame
	var out strings.Builder

	// Close lists down to the target level
	closeTo := func(target int) {
		for len(stack) > 0 && stack[len(stack)-1].level > target {
			f := stack[len(stack)-1]
			out.WriteString("</" + f.kind + ">\n")
			stack = stack[:len(stack)-1]
		}
	}

	openList := func(kind string, lvl int) {
		out.WriteString("<" + kind + ">\n")
		stack = append(stack, frame{kind: kind, level: lvl})
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if m := itemUL.FindStringSubmatch(line); m != nil {
			lvl := levelFromWS(m[1])
			text := m[3]

			closeTo(lvl)
			// If same level but different list type, switch the list at this level
			if len(stack) > 0 && stack[len(stack)-1].level == lvl && stack[len(stack)-1].kind != "ul" {
				out.WriteString("</" + stack[len(stack)-1].kind + ">\n")
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 || stack[len(stack)-1].kind != "ul" || stack[len(stack)-1].level != lvl {
				openList("ul", lvl)
			}
			out.WriteString("<li>" + text + "</li>\n")
			continue
		}

		if m := itemOL.FindStringSubmatch(line); m != nil {
			lvl := levelFromWS(m[1])
			text := m[3]

			closeTo(lvl)
			if len(stack) > 0 && stack[len(stack)-1].level == lvl && stack[len(stack)-1].kind != "ol" {
				out.WriteString("</" + stack[len(stack)-1].kind + ">\n")
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 || stack[len(stack)-1].kind != "ol" || stack[len(stack)-1].level != lvl {
				openList("ol", lvl)
			}
			out.WriteString("<li>" + text + "</li>\n")
			continue
		}

		// Non-list line: close any open lists and write the line as-is
		closeTo(-1)
		out.WriteString(line)
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	// Close any remaining open lists
	closeTo(-1)

	// Restore protected <pre> blocks
	result := out.String()
	for i, m := range placeholders {
		result = strings.ReplaceAll(result, fmt.Sprintf("[[[PRE_BLOCK_%d]]]", i), m)
	}
	return result
}

// (custom pipe table conversion removed; rely on Markdown renderer's Tables extension)

// Markdown renderer with common extensions enabled
func renderMarkdown(content string) string {
	exts := blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs | blackfriday.FencedCode | blackfriday.Tables | blackfriday.Strikethrough
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{})
	out := blackfriday.Run([]byte(content), blackfriday.WithExtensions(exts), blackfriday.WithRenderer(renderer))
	return string(out)
}

// convertInlineEmphasisInHTML converts **bold** and *italic* markers that appear inside
// existing HTML elements (e.g., <li>, <p>) into <strong>/<em> so they render properly.
// It only rewrites text between tags (not inside attributes or tags).
func convertInlineEmphasisInHTML(html string) string {
	// Protect code/pre blocks so we don't rewrite markers in code
	codeRe := regexp.MustCompile("(?is)(<pre[\\s\\S]*?</pre>|<code[\\s\\S]*?</code>)")
	placeholders := []string{}
	html = codeRe.ReplaceAllStringFunc(html, func(m string) string {
		placeholders = append(placeholders, m)
		return fmt.Sprintf("[[[EMPH_PROTECT_%d]]]", len(placeholders)-1)
	})

	// Replace emphasis only within text nodes between > and <
	textNodeRe := regexp.MustCompile(">([^<]+)<")
	html = textNodeRe.ReplaceAllStringFunc(html, func(seg string) string {
		// Extract inner text
		inner := seg[1 : len(seg)-1]
		// Strong: **text** or __text__
		inner = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(inner, "<strong>$1</strong>")
		inner = regexp.MustCompile(`__([^_]+)__`).ReplaceAllString(inner, "<strong>$1</strong>")
		// Emphasis: *text* or _text_ (avoid matching already converted strong/em tags)
		inner = regexp.MustCompile(`(^|[^*])\*([^*]+)\*([^*]|$)`).ReplaceAllString(inner, "$1<em>$2</em>$3")
		inner = regexp.MustCompile(`(^|[^_])_([^_]+)_([^_]|$)`).ReplaceAllString(inner, "$1<em>$2</em>$3")
		return ">" + inner + "<"
	})

	// Restore protected code/pre blocks
	for i, m := range placeholders {
		html = strings.ReplaceAll(html, fmt.Sprintf("[[[EMPH_PROTECT_%d]]]", i), m)
	}
	return html
}

// Convert simple markdown-like markers that users may have typed inside HTML paragraphs
// generated by the WYSIWYG editor (e.g., <p>## Heading</p>, <p>- item</p>).
func preprocessLooseMarkdownHTML_dup(content string) string {
	// --- Normalize whitespace early so indentation is reliable ---
	// Convert common Unicode spaces to ASCII space
	content = strings.NewReplacer(
		"\u00A0", " ", // NBSP
		"\u2002", " ", // en space
		"\u2003", " ", // em space
		"\u2007", " ", // figure space
		"\u202F", " ", // narrow NBSP
	).Replace(content)
	// Normalize entities and line breaks
	content = strings.ReplaceAll(content, "&nbsp;", " ")
	content = strings.ReplaceAll(content, "&#160;", " ")
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "<br>", "\n")
	content = strings.ReplaceAll(content, "<br/>", "\n")
	content = strings.ReplaceAll(content, "<br />", "\n")

	// Ensure markdown resumes after HTML blocks by injecting a blank line after closing tags
	// This helps Blackfriday recognize headings that appear immediately after a gallery or other HTML block
	closeBlock := regexp.MustCompile(`(?is)</(div|figure|section|table|blockquote)>\s*`)
	content = closeBlock.ReplaceAllString(content, "</$1>\n\n")

	// Protect code blocks to avoid altering their contents
	preRe := regexp.MustCompile(`(?is)<pre[\s\S]*?</pre>`) // matches <pre>...</pre>
	placeholders := []string{}
	content = preRe.ReplaceAllStringFunc(content, func(m string) string {
		placeholders = append(placeholders, m)
		return fmt.Sprintf("[[[PRE_BLOCK_%d]]]", len(placeholders)-1)
	})

	// Convert top-level markdown lines too (not only inside <p>)
	// Headings
	reTopH3 := regexp.MustCompile(`(?m)^[ \t]*###[ \t]+(.+)$`)
	content = reTopH3.ReplaceAllString(content, `<h3>$1</h3>`)
	reTopH2 := regexp.MustCompile(`(?m)^[ \t]*##[ \t]+(.+)$`)
	content = reTopH2.ReplaceAllString(content, `<h2>$1</h2>`)
	reTopH1 := regexp.MustCompile(`(?m)^[ \t]*#[ \t]+(.+)$`)
	content = reTopH1.ReplaceAllString(content, `<h1>$1</h1>`)

	// Process multi-line blockquotes starting with > or &gt;
	content = processBlockquotes(content)

	// Paragraph-wrapped headings
	reH3 := regexp.MustCompile(`(?is)<p>\s*###\s+(.+?)\s*</p>`)
	content = reH3.ReplaceAllString(content, `<h3>$1</h3>`)
	reH2 := regexp.MustCompile(`(?is)<p>\s*##\s+(.+?)\s*</p>`)
	content = reH2.ReplaceAllString(content, `<h2>$1</h2>`)
	reH1 := regexp.MustCompile(`(?is)<p>\s*#\s+(.+?)\s*</p>`)
	content = reH1.ReplaceAllString(content, `<h1>$1</h1>`)

	// Paragraph-wrapped list markers -> plain lines so nested builder can work
	rePUL := regexp.MustCompile(`(?is)<p>\s*([ \t]*)([-*+])\s+(.+?)\s*</p>`)
	content = rePUL.ReplaceAllString(content, `$1$2 $3`)
	rePOL := regexp.MustCompile(`(?is)<p>\s*([ \t]*)(\d+)\.\s+(.+?)\s*</p>`)
	content = rePOL.ReplaceAllString(content, `$1$2. $3`)

	// NOTE: Do NOT convert top-level list items to custom tags; that kills indentation.
	// We intentionally removed the rules that produced <ul-li> / <ol-li>.

	// Unwrap paragraphs that contain Markdown emphasis so the Markdown renderer can parse them
	rePEmph := regexp.MustCompile(`(?is)<p>\s*([^<>]*?(\*\*.+?\*\*|__.+?__|\*[^*]+?\*|_[^_]+?_)\s*[^<>]*?)\s*</p>`) // no nested tags
	content = rePEmph.ReplaceAllString(content, "\n$1\n")

	// Horizontal rule (top-level and paragraph-wrapped)
	reTopHR := regexp.MustCompile(`(?m)^[ \t]*---[ \t]*$`)
	content = reTopHR.ReplaceAllString(content, `<hr/>`)
	rePHR := regexp.MustCompile(`(?is)<p>\s*---\s*</p>`)
	content = rePHR.ReplaceAllString(content, `<hr/>`)

	// Also handle headings that appear as raw text immediately inside a container
	// e.g., <div>## Heading</div> → <div><h2>Heading</h2></div>
	reInH3 := regexp.MustCompile(`(?is)>(\s*###\s+)(.+?)\s*<`)
	content = reInH3.ReplaceAllString(content, `><h3>$2</h3><`)
	reInH2 := regexp.MustCompile(`(?is)>(\s*##\s+)(.+?)\s*<`)
	content = reInH2.ReplaceAllString(content, `><h2>$2</h2><`)
	reInH1 := regexp.MustCompile(`(?is)>(\s*#\s+)(.+?)\s*<`)
	content = reInH1.ReplaceAllString(content, `><h1>$2</h1><`)

	// Restore code blocks
	for i, m := range placeholders {
		content = strings.ReplaceAll(content, fmt.Sprintf("[[[PRE_BLOCK_%d]]]", i), m)
	}
	return content
}

// Build nested UL/OL lists based on indentation in raw lines.
// Nests when indentation reaches 2 spaces OR 1 tab per level.
func buildNestedLists_dup(content string) string {
	// Protect code blocks
	preRe := regexp.MustCompile(`(?is)<pre[\s\S]*?</pre>`)
	placeholders := []string{}
	content = preRe.ReplaceAllStringFunc(content, func(m string) string {
		placeholders = append(placeholders, m)
		return fmt.Sprintf("[[[PRE_BLOCK_%d]]]", len(placeholders)-1)
	})

	lines := strings.Split(content, "\n")

	// Patterns for list items
	itemUL := regexp.MustCompile(`^([ \t]*)([-*+])\s+(.+)$`)
	itemOL := regexp.MustCompile(`^([ \t]*)(\d+)\.\s+(.+)$`)

	// Convert leading whitespace to a discrete nesting level.
	// Treat 1 tab == 2 spaces; 2 spaces per level.
	levelFromWS := func(ws string) int {
		tabs, spaces := 0, 0
		for i := 0; i < len(ws); i++ {
			if ws[i] == '\t' {
				tabs++
			} else if ws[i] == ' ' {
				spaces++
			}
		}
		units := tabs*2 + spaces // 1 tab == 2 spaces
		return units / 2         // 2 spaces per level
	}

	type frame struct {
		kind  string // "ul" or "ol"
		level int
	}
	var stack []frame
	var out strings.Builder

	// Close lists down to the target level
	closeTo := func(target int) {
		for len(stack) > 0 && stack[len(stack)-1].level > target {
			f := stack[len(stack)-1]
			out.WriteString("</" + f.kind + ">\n")
			stack = stack[:len(stack)-1]
		}
	}

	openList := func(kind string, lvl int) {
		out.WriteString("<" + kind + ">\n")
		stack = append(stack, frame{kind: kind, level: lvl})
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if m := itemUL.FindStringSubmatch(line); m != nil {
			lvl := levelFromWS(m[1])
			text := m[3]

			closeTo(lvl)
			// If same level but different list type, switch the list at this level
			if len(stack) > 0 && stack[len(stack)-1].level == lvl && stack[len(stack)-1].kind != "ul" {
				out.WriteString("</" + stack[len(stack)-1].kind + ">\n")
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 || stack[len(stack)-1].kind != "ul" || stack[len(stack)-1].level != lvl {
				openList("ul", lvl)
			}
			out.WriteString("<li>" + text + "</li>\n")
			continue
		}

		if m := itemOL.FindStringSubmatch(line); m != nil {
			lvl := levelFromWS(m[1])
			text := m[3]

			closeTo(lvl)
			if len(stack) > 0 && stack[len(stack)-1].level == lvl && stack[len(stack)-1].kind != "ol" {
				out.WriteString("</" + stack[len(stack)-1].kind + ">\n")
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 || stack[len(stack)-1].kind != "ol" || stack[len(stack)-1].level != lvl {
				openList("ol", lvl)
			}
			out.WriteString("<li>" + text + "</li>\n")
			continue
		}

		// Non-list line: close any open lists and write the line as-is
		closeTo(-1)
		out.WriteString(line)
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	// Close any remaining open lists
	closeTo(-1)

	// Restore protected <pre> blocks
	result := out.String()
	for i, m := range placeholders {
		result = strings.ReplaceAll(result, fmt.Sprintf("[[[PRE_BLOCK_%d]]]", i), m)
	}
	return result
}

// (custom pipe table conversion removed; rely on Markdown renderer's Tables extension)

// Markdown renderer with common extensions enabled
func renderMarkdown_dup(content string) string {
	exts := blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs | blackfriday.FencedCode | blackfriday.Tables | blackfriday.Strikethrough
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{})
	out := blackfriday.Run([]byte(content), blackfriday.WithExtensions(exts), blackfriday.WithRenderer(renderer))
	return string(out)
}

// Convert ```lang\ncode``` fences (if present pre-HTML) to HTML blocks for Prism
func convertFences(s string) string {
	re := regexp.MustCompile("(?s)```([a-zA-Z0-9_-]*)\\s*(.*?)```")
	return re.ReplaceAllStringFunc(s, func(m string) string {
		sm := re.FindStringSubmatch(m)
		if len(sm) < 3 {
			return m
		}
		lang := strings.TrimSpace(sm[1])
		code := cleanStyleHeader(sm[2])
		return fmt.Sprintf(`<pre><code class="language-%s">%s</code></pre>`, lang, escapeCode(code))
	})
}

// Escape code without converting quotes to numeric entities to avoid double-encoding with Prism
func escapeCode(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// Remove a single CSS line like "pre code{ ... }" at the start of a fenced block
// while keeping the rest (e.g., .token.keyword{ ... })
func cleanStyleHeader(code string) string {
	lines := strings.Split(code, "\n")
	if len(lines) == 0 {
		return code
	}
	first := strings.TrimSpace(lines[0])
	preLine := regexp.MustCompile(`^pre\s+code\s*\{[^}]*\}\s*$`)
	if preLine.MatchString(first) {
		lines = lines[1:]
	}
	return strings.Join(lines, "\n")
}

// wrapImageGalleries finds <div class="image-gallery"> blocks and wraps any
// bare <img ...> elements with <a data-lightbox="article-images">...</a>
// to match preview behavior and enable Lightbox on the blog page without JS.
func wrapImageGalleries(html string) string {
	// Match any div that contains class "image-gallery"
	galleryRe := regexp.MustCompile(`(?is)(<div[^>]*class="[^"]*image-gallery[^"]*"[^>]*>)([\s\S]*?)(</div>)`)

	// Regex to match an <img ...> tag and capture pre/post attributes and src
	imgRe := regexp.MustCompile(`(?is)<img([^>]*?)\s+src="([^"]+)"([^>]*)>`)
	// Regex to extract alt text if present inside the combined attributes
	altRe := regexp.MustCompile(`(?i)alt="([^"]*)"`)

	transformGallery := func(content string) string {
		items := imgRe.ReplaceAllStringFunc(content, func(imgTag string) string {
			// Extract attributes and src
			m := imgRe.FindStringSubmatch(imgTag)
			if len(m) != 4 {
				return imgTag
			}
			preAttrs := m[1]
			src := m[2]
			postAttrs := m[3]
			attrs := preAttrs + " " + postAttrs
			alt := ""
			if am := altRe.FindStringSubmatch(attrs); len(am) == 2 {
				alt = am[1]
			}
			// Build wrapped anchor preserving original <img ...> markup
			anchor := fmt.Sprintf(`<a href="%s" data-lightbox="article-images" rel="lightbox[article-images]" data-title="%s"><img%s src="%s"%s></a>`,
				src, htmlEscapeAttr(alt), preAttrs, src, postAttrs)
			return anchor
		})
		return items
	}

	// Replace each gallery content separately to avoid over-wrapping
	return galleryRe.ReplaceAllStringFunc(html, func(block string) string {
		parts := galleryRe.FindStringSubmatch(block)
		if len(parts) != 4 {
			return block
		}
		open := parts[1]
		inner := parts[2]
		close := parts[3]
		// Keep original container; layout handled by CSS at render time
		return open + transformGallery(inner) + close
	})
}

// htmlEscapeAttr escapes quotes for safe placement in attributes
func htmlEscapeAttr(s string) string {
	s = strings.ReplaceAll(s, `"`, `&quot;`)
	s = strings.ReplaceAll(s, `&`, `&amp;`)
	s = strings.ReplaceAll(s, `<`, `&lt;`)
	s = strings.ReplaceAll(s, `>`, `&gt;`)
	return s
}

// (inline style injector removed; we rely on stylesheet in template)
