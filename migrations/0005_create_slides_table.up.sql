-- Create the Slides table
CREATE TABLE IF NOT EXISTS Slides (
    slide_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES Users(user_id),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    content_file_path TEXT NOT NULL,
    is_published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create the Slide_Categories junction table for many-to-many relationship
CREATE TABLE IF NOT EXISTS Slide_Categories (
    slide_id INT REFERENCES Slides(slide_id) ON DELETE CASCADE,
    category_id INT REFERENCES Categories(category_id) ON DELETE CASCADE,
    PRIMARY KEY (slide_id, category_id)
);

-- Create index for better query performance
CREATE INDEX IF NOT EXISTS idx_slides_user_id ON Slides(user_id);
CREATE INDEX IF NOT EXISTS idx_slides_slug ON Slides(slug);
CREATE INDEX IF NOT EXISTS idx_slides_published ON Slides(is_published);