-- Create the Post_Categories table for many-to-many relationship between posts and categories
-- This allows posts to have multiple categories (displayed as tags to users)
CREATE TABLE IF NOT EXISTS Post_Categories (
    post_category_id SERIAL PRIMARY KEY,
    post_id INT REFERENCES Posts(post_id) ON DELETE CASCADE,
    category_id INT REFERENCES Categories(category_id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(post_id, category_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_post_categories_post_id ON Post_Categories(post_id);
CREATE INDEX IF NOT EXISTS idx_post_categories_category_id ON Post_Categories(category_id);

-- Insert default categories if they don't exist
-- First check and insert each category individually
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM Categories WHERE category_name = 'Performance') THEN
        INSERT INTO Categories (category_name) VALUES ('Performance');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM Categories WHERE category_name = 'Architecture') THEN
        INSERT INTO Categories (category_name) VALUES ('Architecture');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM Categories WHERE category_name = 'Scale') THEN
        INSERT INTO Categories (category_name) VALUES ('Scale');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM Categories WHERE category_name = 'Cloud Computing') THEN
        INSERT INTO Categories (category_name) VALUES ('Cloud Computing');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM Categories WHERE category_name = 'AI') THEN
        INSERT INTO Categories (category_name) VALUES ('AI');
    END IF;
END $$;

-- For existing posts that have a category_id, migrate them to use the new Post_Categories table
INSERT INTO Post_Categories (post_id, category_id)
SELECT post_id, category_id 
FROM Posts 
WHERE category_id IS NOT NULL
ON CONFLICT (post_id, category_id) DO NOTHING;