-- Add featured boolean column to posts table
ALTER TABLE posts ADD COLUMN featured BOOLEAN DEFAULT false;

-- Add index on featured column for better query performance
CREATE INDEX idx_posts_featured ON posts(featured);

-- Update existing posts with featured images to be marked as featured
UPDATE posts SET featured = true WHERE featured_image_url IS NOT NULL AND featured_image_url != '';