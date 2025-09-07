package models

import (
    "database/sql"
    "fmt"
    "html"
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
    const listTag = "<ul>"
    const listClass = "list-disc pl-4"
    if idx := strings.Index(content, listTag); idx != -1 {
        before := content[:idx]
        after := content[idx+len(listTag):]
        return before + "<ul class=\"" + listClass + "\">" + after
    }
    const olTag = "<ol>"
    const olClass = "list-decimal pl-4"
    if idx := strings.Index(content, olTag); idx != -1 {
        before := content[:idx]
        after := content[idx+len(olTag):]
        return before + "<ol class=\"" + olClass + "\">" + after
    }
    const liTag = "<li>"
    const liClass = "mb-2"
    if idx := strings.Index(content, liTag); idx != -1 {
        before := content[:idx]
        after := content[idx+len(liTag):]
        return before + "<li class=\"" + liClass + "\">" + after
    }
    return content
}

func replaceBlockquoteTag(content string) string {
    const blockquoteTag = "<blockquote>"
    const blockquoteClass = "p-4 my-4 border-s-4 border-gray-300 bg-gray-50 dark:border-gray-500 dark:bg-gray-800"
    if idx := strings.Index(content, blockquoteTag); idx != -1 {
        before := content[:idx]
        after := content[idx+len(blockquoteTag):]
        return before + "<blockquote class=\"" + blockquoteClass + "\">" + after
    }
    return content
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
        code := sm[2]
        return fmt.Sprintf(`<pre><code class="language-%s">%s</code></pre>`, lang, html.EscapeString(code))
    })
}

