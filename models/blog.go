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
		// Don't preprocess markdown that might interfere with blackfriday
		// content = preprocessLooseMarkdownHTML(content)
		// content = buildNestedLists(content)
		content = normalizeInlinePipeTables(content)
		content = convertPipeTablesSimple(content)
		// Convert Markdown-style fenced blocks to HTML before markdown render,
		// because editor content may mix HTML and backticks.
		content = convertFences(content)
		// Let blackfriday process the complete markdown including lists and links
		htmlOut := renderMarkdown(content)
		htmlOut = replaceBlockquoteTag(replacelistTag(htmlOut))
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
	content = buildNestedLists(content)
	content = normalizeInlinePipeTables(content)
	content = convertPipeTablesSimple(content)
	content = convertFences(content)
	htmlOut := renderMarkdown(content)
	htmlOut = replaceBlockquoteTag(replacelistTag(htmlOut))
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

	// Blockquotes starting with > or &gt;
	reTopBQ := regexp.MustCompile(`(?m)^[ \t]*(&gt;|>)[ \t]+(.+)$`)
	content = reTopBQ.ReplaceAllString(content, `<blockquote><p>$2</p></blockquote>`)

	// Paragraph-wrapped headings
	reH3 := regexp.MustCompile(`(?is)<p>\s*###\s+(.+?)\s*</p>`)
	content = reH3.ReplaceAllString(content, `<h3>$1</h3>`)
	reH2 := regexp.MustCompile(`(?is)<p>\s*##\s+(.+?)\s*</p>`)
	content = reH2.ReplaceAllString(content, `<h2>$1</h2>`)

	// Paragraph-wrapped list markers -> plain lines so nested builder can work
	rePUL := regexp.MustCompile(`(?is)<p>\s*([ \t]*)([-*+])\s+(.+?)\s*</p>`)
	content = rePUL.ReplaceAllString(content, `$1$2 $3`)
	rePOL := regexp.MustCompile(`(?is)<p>\s*([ \t]*)(\d+)\.\s+(.+?)\s*</p>`)
	content = rePOL.ReplaceAllString(content, `$1$2. $3`)

	// NOTE: Do NOT convert top-level list items to custom tags; that kills indentation.
	// We intentionally removed the rules that produced <ul-li> / <ol-li>.

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

// Convert simple pipe table blocks (header | alignment line | rows) into HTML tables.
func convertPipeTablesSimple(content string) string {
	// Protect code blocks
	preRe := regexp.MustCompile(`(?is)<pre[\s\S]*?</pre>`)
	placeholders := []string{}
	content = preRe.ReplaceAllStringFunc(content, func(m string) string {
		placeholders = append(placeholders, m)
		return fmt.Sprintf("[[[PRE_BLOCK_%d]]]", len(placeholders)-1)
	})

	lines := strings.Split(content, "\n")
	var out strings.Builder

	i := 0
	for i < len(lines) {
		// detect header line with pipes
		if strings.Contains(lines[i], "|") {
			if i+1 < len(lines) {
				sep := strings.TrimSpace(lines[i+1])
				if regexp.MustCompile(`^\|?[ \t]*[:\-\| ]+\|?$`).MatchString(sep) {
					// collect rows until a blank line or non-pipe line
					j := i + 2
					for j < len(lines) && strings.Contains(lines[j], "|") && strings.TrimSpace(lines[j]) != "" {
						j++
					}
					block := lines[i:j]
					out.WriteString(pipeBlockToTable(block))
					out.WriteByte('\n')
					i = j
					continue
				}
			}
		}
		out.WriteString(lines[i])
		out.WriteByte('\n')
		i++
	}
	result := out.String()
	for i, m := range placeholders {
		result = strings.ReplaceAll(result, fmt.Sprintf("[[[PRE_BLOCK_%d]]]", i), m)
	}
	return result
}

func pipeBlockToTable(block []string) string {
	mk := func(row string) []string {
		row = strings.TrimSpace(row)
		if strings.HasPrefix(row, "|") {
			row = row[1:]
		}
		if strings.HasSuffix(row, "|") {
			row = row[:len(row)-1]
		}
		parts := strings.Split(row, "|")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}
	headers := mk(block[0])
	aligns := mk(block[1])
	// Build table
	var b strings.Builder
	b.WriteString("<table>\n<thead><tr>")
	for _, h := range headers {
		b.WriteString("<th>" + h + "</th>")
	}
	b.WriteString("</tr></thead>\n<tbody>\n")
	for _, row := range block[2:] {
		cells := mk(row)
		b.WriteString("<tr>")
		for _, c := range cells {
			b.WriteString("<td>" + c + "</td>")
		}
		b.WriteString("</tr>\n")
	}
	b.WriteString("</tbody>\n</table>")
	_ = aligns // alignment not applied in this simple version
	return b.String()
}

// Markdown renderer with common extensions enabled
func renderMarkdown(content string) string {
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
