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

        // Format dates for display
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

        // Remove read-more marker then render Markdown.
        // We render Markdown unconditionally—blackfriday passes raw HTML through
        // and converts Markdown primitives (headings, lists, fences) correctly.
        content := replaceMoreTag(post.Content)
        htmlOut := renderMarkdown(content)
        htmlOut = replaceBlockquoteTag(replacelistTag(htmlOut))
        post.ContentHTML = template.HTML(htmlOut)
	}

	if err != nil {
		return nil, fmt.Errorf("Post could not be fetched: %w", err)
	} else {
		fmt.Println("Posts fetched successfully!")
	}

	fmt.Println("Blog Post:", post)

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
	// replace <ul> tag with <ul class="list-disc pl-4">
	const listTag = "<ul>"
	const listClass = "list-disc pl-4"
	if idx := strings.Index(content, listTag); idx != -1 {
		beforeListTag := content[:idx]
		afterListTag := content[idx+len(listTag):]
		return beforeListTag + "<ul class=\"" + listClass + "\">" + afterListTag
	}
	// replace <ol> tag with <ol class="list-decimal pl-4">
	const olTag = "<ol>"
	const olClass = "list-decimal pl-4"
	if idx := strings.Index(content, olTag); idx != -1 {
		beforeOlTag := content[:idx]
		afterOlTag := content[idx+len(olTag):]
		return beforeOlTag + "<ol class=\"" + olClass + "\">" + afterOlTag
	}
	// replace <li> tag with <li class="mb-2">
	const liTag = "<li>"
	const liClass = "mb-2"
	if idx := strings.Index(content, liTag); idx != -1 {
		beforeLiTag := content[:idx]
		afterLiTag := content[idx+len(liTag):]
		return beforeLiTag + "<li class=\"" + liClass + "\">" + afterLiTag
	}
	return content
}

func replaceBlockquoteTag(content string) string {
	// replace <blockquote> tag with <blockquote class="border-l-4 border-primary-500 pl-4 mb-4">
	const blockquoteTag = "<blockquote>"
	const blockquoteClass = "p-4 my-4 border-s-4 border-gray-300 bg-gray-50 dark:border-gray-500 dark:bg-gray-800"
	if idx := strings.Index(content, blockquoteTag); idx != -1 {
		beforeBlockquoteTag := content[:idx]
		afterBlockquoteTag := content[idx+len(blockquoteTag):]
		return beforeBlockquoteTag + "<blockquote class=\"" + blockquoteClass + "\">" + afterBlockquoteTag
	}
	return content
}

// Function to render markdown content
func renderMarkdown(content string) string {
    output := blackfriday.Run([]byte(content))
    // fmt.Println(string(output))
    return string(output)
}

// convertFences converts ```lang\ncode``` fences into HTML blocks for Prism
func convertFences(s string) string {
    re := regexp.MustCompile("(?s)```([a-zA-Z0-9_-]*)\\s*(.*?)```")
    return re.ReplaceAllStringFunc(s, func(m string) string {
        sm := re.FindStringSubmatch(m)
        if len(sm) < 3 { return m }
        lang := strings.TrimSpace(sm[1])
        code := sm[2]
        return fmt.Sprintf(`<pre><code class="language-%s">%s</code></pre>`, lang, html.EscapeString(code))
    })
}
