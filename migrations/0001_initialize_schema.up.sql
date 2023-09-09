-- Create the Roles table
CREATE TABLE IF NOT EXISTS Roles (
    role_id SERIAL PRIMARY KEY,
    role_name VARCHAR(255) NOT NULL
);

-- Create the Users table
CREATE TABLE IF NOT EXISTS Users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    registration_date TIMESTAMP NOT NULL,
    profile_picture_url TEXT,
    role_id INT REFERENCES Roles(role_id)
);

-- Create the Categories table
CREATE TABLE IF NOT EXISTS Categories (
    category_id SERIAL PRIMARY KEY,
    category_name VARCHAR(255) NOT NULL
);

-- Create the Posts table
CREATE TABLE IF NOT EXISTS Posts (
    post_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES Users(user_id),
    category_id INT REFERENCES Categories(category_id),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    publication_date TIMESTAMP NOT NULL,
    last_edit_date TIMESTAMP,
    is_published BOOLEAN NOT NULL,
    featured_image_url TEXT
);

-- Create the Likes table
CREATE TABLE IF NOT EXISTS Likes (
    like_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES Users(user_id),
    post_id INT REFERENCES Posts(post_id),
    like_date TIMESTAMP NOT NULL
);

-- Create the Comments table
CREATE TABLE IF NOT EXISTS Comments (
    comment_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES Users(user_id),
    post_id INT REFERENCES Posts(post_id),
    parent_comment_id INT REFERENCES Comments(comment_id),
    content TEXT NOT NULL,
    comment_date TIMESTAMP NOT NULL
);

-- Create the Tags table
CREATE TABLE IF NOT EXISTS Tags (
    tag_id SERIAL PRIMARY KEY,
    tag_name VARCHAR(255) NOT NULL
);

-- Create the Post_Tags table for the many-to-many relationship
CREATE TABLE IF NOT EXISTS Post_Tags (
    post_tag_id SERIAL PRIMARY KEY,
    post_id INT REFERENCES Posts(post_id),
    tag_id INT REFERENCES Tags(tag_id)
);

-- Create the Drafts table
CREATE TABLE IF NOT EXISTS Drafts (
    draft_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES Users(user_id),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    creation_date TIMESTAMP NOT NULL,
    last_edit_date TIMESTAMP
);
