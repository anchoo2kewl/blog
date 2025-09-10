-- Remove Post_Categories table
DROP TABLE IF EXISTS Post_Categories;

-- Remove default categories (optional - only if you want to completely reverse)
DELETE FROM Categories WHERE category_name IN ('Performance', 'Architecture', 'Scale', 'Cloud Computing', 'AI');