package models

import (
    "database/sql"
    "fmt"
    "html/template"
    "regexp"
    "strings"
    "time"
)

type PostsList struct {
	Posts []Post
}

type Post struct {
	ID               int
	UserID           int // Added UserID field
	CategoryID       int
	Title            string
	Content          string
	ContentHTML      template.HTML
	Slug             string
	PublicationDate  string
	LastEditDate     string
	IsPublished      bool
	FeaturedImageURL string
	CreatedAt        string
}

type PostService struct {
    DB *sql.DB
}

// Create will create a new session for the user provided. The session token
// will be returned as the Token field on the Session type, but only the hashed
// session token is stored in the database.
// func (pp *PostService) Create() (*Post, error) {

// }

func (pp *PostService) GetTopPosts() (*PostsList, error) {
	list := PostsList{}

	query := `SELECT * FROM posts LIMIT 5`
	rows, err := pp.DB.Query(query)
	if err != nil {
		return &list, nil
	}

	for rows.Next() {

		var post Post
		err := rows.Scan(&post.ID, &post.UserID, &post.CategoryID, &post.Title, &post.Content, &post.Slug, &post.PublicationDate, &post.LastEditDate, &post.IsPublished, &post.FeaturedImageURL, &post.CreatedAt)
		if err != nil {
			panic(err)
		}

		// Parse and format CreatedAt
		t, err := time.Parse(time.RFC3339, post.CreatedAt)
		if err != nil {
			fmt.Println(err)
		}
		post.CreatedAt = t.Format(time.RFC3339) // Keep original for JavaScript
		post.PublicationDate = t.Format("January 2, 2006") // Readable fallback
		
		// Parse and format PublicationDate if it's different from CreatedAt
		if post.PublicationDate != "" && post.PublicationDate != post.CreatedAt {
			pubDate, pubErr := time.Parse(time.RFC3339, post.PublicationDate)
			if pubErr == nil {
				post.PublicationDate = pubDate.Format("January 2, 2006")
			}
		}

		post.Content = trimContent(post.Content)

		list.Posts = append(list.Posts, post)
	}

	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	} else {
		fmt.Println("Posts fetched successfully!")
	}

	return &list, nil
}

func (pp *PostService) GetAllPosts() (*PostsList, error) {
	list := PostsList{}

	query := `SELECT * FROM posts ORDER BY created_at DESC`
	rows, err := pp.DB.Query(query)
	if err != nil {
		return &list, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.UserID, &post.CategoryID, &post.Title, &post.Content, &post.Slug, &post.PublicationDate, &post.LastEditDate, &post.IsPublished, &post.FeaturedImageURL, &post.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Parse and format CreatedAt
		t, err := time.Parse(time.RFC3339, post.CreatedAt)
		if err != nil {
			fmt.Println(err)
		}
		post.CreatedAt = t.Format(time.RFC3339)
		post.PublicationDate = t.Format("January 2, 2006")
		
		// Parse and format PublicationDate if it's different from CreatedAt
		if post.PublicationDate != "" && post.PublicationDate != post.CreatedAt {
			pubDate, pubErr := time.Parse(time.RFC3339, post.PublicationDate)
			if pubErr == nil {
				post.PublicationDate = pubDate.Format("January 2, 2006")
			}
		}

		post.Content = trimContent(post.Content)
		list.Posts = append(list.Posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("get all posts: %w", err)
	}

	return &list, nil
}

func (pp *PostService) GetPostsByUser(userID int) (*PostsList, error) {
	list := PostsList{}

	query := `SELECT * FROM posts WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := pp.DB.Query(query, userID)
	if err != nil {
		return &list, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.UserID, &post.CategoryID, &post.Title, &post.Content, &post.Slug, &post.PublicationDate, &post.LastEditDate, &post.IsPublished, &post.FeaturedImageURL, &post.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Parse and format CreatedAt
		t, err := time.Parse(time.RFC3339, post.CreatedAt)
		if err != nil {
			fmt.Println(err)
		}
		post.CreatedAt = t.Format(time.RFC3339)
		post.PublicationDate = t.Format("January 2, 2006")
		
		// Parse and format PublicationDate if it's different from CreatedAt
		if post.PublicationDate != "" && post.PublicationDate != post.CreatedAt {
			pubDate, pubErr := time.Parse(time.RFC3339, post.PublicationDate)
			if pubErr == nil {
				post.PublicationDate = pubDate.Format("January 2, 2006")
			}
		}

		post.Content = trimContent(post.Content)
		list.Posts = append(list.Posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("get posts by user: %w", err)
	}

	return &list, nil
}

// Function to trim content up to the <more--> tag
func trimContent(content string) string {
    // Prefer everything before read-more marker; support escaped version too
    if idx := strings.Index(content, "<more-->"); idx != -1 {
        content = content[:idx]
    } else if idx := strings.Index(content, "&lt;more--&gt;"); idx != -1 {
        content = content[:idx]
    }
    // Remove fenced code blocks ```...```
    fence := regexp.MustCompile("(?s)```.*?```")
    content = fence.ReplaceAllString(content, " ")
    // Remove stray backticks
    content = strings.ReplaceAll(content, "```", " ")
    content = strings.ReplaceAll(content, "`", "")
    // Strip HTML tags
    content = stripHTML(content)
    // Collapse whitespace
    words := strings.Fields(content)
    if len(words) == 0 {
        return ""
    }
    // If there was no read-more, fall back to first N words
    N := 40
    if len(words) > N {
        words = words[:N]
    }
    return strings.Join(words, " ")
}

func stripHTML(s string) string {
    var b strings.Builder
    in := false
    for _, r := range s {
        switch r {
        case '<':
            in = true
        case '>':
            in = false
        default:
            if !in {
                b.WriteRune(r)
            }
        }
    }
    return b.String()
}

func (pp *PostService) Create(userID int, categoryID int, title, content string, isPublished bool, featuredImageURL string, slug string) (*Post, error) {
	timefmt := time.Now()
	query := `
		INSERT INTO posts (user_id, category_id, title, content, slug, publication_date, last_edit_date, is_published, featured_image_url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING post_id
	`
	var postID int
	println(userID, categoryID, title, content, isPublished, featuredImageURL)
	err := pp.DB.QueryRow(query, userID, categoryID, title, content, slug, timefmt,
		timefmt, isPublished, featuredImageURL, timefmt).Scan(&postID)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return nil, fmt.Errorf("create post: %w", err)
	}
	fmt.Println("Post created successfully!")
	fmt.Println(postID)

	return &Post{
		ID:               postID,
		UserID:           userID,
		CategoryID:       categoryID,
		Title:            title,
		Content:          content,
		Slug:             slug,
		PublicationDate:  timefmt.Format("January 2, 2006"),
		LastEditDate:     timefmt.Format("January 2, 2006"),
		IsPublished:      isPublished,
		FeaturedImageURL: featuredImageURL,
		CreatedAt:        timefmt.Format("January 2, 2006"),
	}, nil
}

func (pp *PostService) GetByID(id int) (*Post, error) {
    var post Post
    row := pp.DB.QueryRow(`SELECT post_id, user_id, category_id, title, content, slug, publication_date, last_edit_date, is_published, featured_image_url, created_at FROM posts WHERE post_id=$1`, id)
    if err := row.Scan(&post.ID, &post.UserID, &post.CategoryID, &post.Title, &post.Content, &post.Slug, &post.PublicationDate, &post.LastEditDate, &post.IsPublished, &post.FeaturedImageURL, &post.CreatedAt); err != nil {
        return nil, err
    }
    return &post, nil
}

func (pp *PostService) Update(id int, categoryID int, title, content string, isPublished bool, featuredImageURL, slug string) error {
    _, err := pp.DB.Exec(`UPDATE posts SET category_id=$1, title=$2, content=$3, slug=$4, last_edit_date=$5, is_published=$6, featured_image_url=$7 WHERE post_id=$8`,
        categoryID, title, content, slug, time.Now(), isPublished, featuredImageURL, id)
    return err
}
