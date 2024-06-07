psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c "INSERT INTO ROLES (role_name) values ('Commenter')"
psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c "INSERT INTO ROLES (role_name) values ('Administrator')"
psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c "INSERT INTO ROLES (role_name) values ('Editor')"
psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c "INSERT INTO ROLES (role_name) values ('Viewer')"

#Check
psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c 'SELECT * FROM roles'

curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $API_TOKEN" -d '{
  "email": "anshuman@biswas.me",
  "username": "anshuman",
  "password": "123456qwertyu",
  "role": 1
}' http://localhost:$APP_PORT/api/users

curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $API_TOKEN" -d '{
  "email": "anchoo2kewl@gmail.com",
  "username": "anchoo2kewl",
  "password": "123456qwertyu",
  "role": 2
}' http://localhost:$APP_PORT/api/users

curl -X GET -H "Authorization: Bearer $API_TOKEN" http://localhost:$APP_PORT/api/users

curl -X GET -H "Authorization: Bearer $API_TOKEN" http://localhost:$APP_PORT/api/posts

psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c "INSERT INTO CATEGORIES (category_name) values ('Demo Category')"

psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c 'SELECT * FROM categories'

curl -X POST -H "Authorization: Bearer $API_TOKEN" -H "Content-Type: application/json" -d '{
  "userID": 2,
  "categoryID": 1,
  "title": "Fictitious Blog Post",
  "slug": "fictitious-blog-post",
  "content": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed nec enim a elit pretium suscipit. Nullam sit amet mauris nisi. Sed at placerat urna. Vivamus nec lectus ac orci dictum ullamcorper vitae eget lacus. Proin pretium turpis sit amet quam egestas, at molestie sapien tempor. Ut euismod odio in risus eleifend, at hendrerit lectus vehicula. Sed eget justo vel felis sollicitudin tincidunt. Nullam ut quam id eros mattis feugiat. Donec vulputate arcu vel nulla accumsan, et dignissim lorem malesuada. Donec quis justo ex. Phasellus scelerisque nunc id tellus sollicitudin, nec convallis ex varius. Suspendisse malesuada odio vel tortor laoreet vestibulum. Nulla facilisi.<!--more-->Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed nec enim a elit pretium suscipit. Nullam sit amet mauris nisi. Sed at placerat urna. Vivamus nec lectus ac orci dictum ullamcorper vitae eget lacus. Proin pretium turpis sit amet quam egestas, at molestie sapien tempor. Ut euismod odio in risus eleifend, at hendrerit lectus vehicula. Sed eget justo vel felis sollicitudin tincidunt. Nullam ut quam id eros mattis feugiat. Donec vulputate arcu vel nulla accumsan, et dignissim lorem malesuada. Donec quis justo ex. Phasellus scelerisque nunc id tellus sollicitudin, nec convallis ex varius. Suspendisse malesuada odio vel tortor laoreet vestibulum. Nulla facilisi.",
  "isPublished": true,
  "featuredImageURL": "image.jpg"
}' http://localhost:$APP_PORT/api/posts

curl -X POST -H "Authorization: Bearer $API_TOKEN" -H "Content-Type: application/json" -d '{
  "userID": 2,
  "categoryID": 1,
  "title": "Exploring the Cosmos",
  "slug": "exploring-the-cosmos",
  "content": "Curabitur dictum malesuada massa, eu tempor leo sagittis nec. Aliquam erat volutpat. Praesent eu augue non purus fermentum consequat. Quisque non nisl vitae enim pretium ultricies. Vivamus dictum nisi ac lectus interdum, ac ultricies turpis malesuada. Sed vestibulum dolor in massa tempor, vel tempor magna suscipit. Nullam ut quam id eros mattis feugiat. Donec vulputate arcu vel nulla accumsan, et dignissim lorem malesuada. Donec quis justo ex.",
  "isPublished": true,
  "featuredImageURL": "cosmos.jpg"
}' http://localhost:$APP_PORT/api/posts

# Sometimes, need the history toggle for zsh
set +H

psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c "update posts set content='# Lorem Ipsum\
\
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed nec enim a elit pretium suscipit. Nullam sit amet mauris nisi. Sed at placerat urna. Vivamus nec lectus ac orci dictum ullamcorper vitae eget lacus. Proin pretium turpis sit amet quam egestas, at molestie sapien tempor. Ut euismod odio in risus eleifend, at hendrerit lectus vehicula. Sed eget justo vel felis sollicitudin tincidunt. Nullam ut quam id eros mattis feugiat. Donec vulputate arcu vel nulla accumsan, et dignissim lorem malesuada. Donec quis justo ex.\
\
<!--more-->\
\
## Second Section\
\
Suspendisse malesuada odio vel tortor laoreet vestibulum. Nulla facilisi. \
\
### Subsection\
\
- Item 1\
- Item 2\
- Item 3\
\
#### Sub-subsection\
\
**Bold Text** *Italic Text* [Link](https://example.com)\
\
##### Sub-sub-subsection\
\
> Blockquote\
\
###### Sub-sub-sub-subsection' where post_id=1;"

curl -X POST -H "Authorization: Bearer $API_TOKEN" -H "Content-Type: application/json" -d '{
  "userID": 2,
  "categoryID": 1,
  "title": "Journey Through Time",
  "slug": "journey-through-time",
  "content": "Sed varius, purus eu posuere convallis, nisi velit tincidunt libero, nec venenatis lorem orci id massa. Integer a ipsum id augue gravida tincidunt. Aenean tincidunt urna et quam tincidunt, sit amet venenatis nisl tristique. Integer tincidunt dapibus purus. Suspendisse malesuada odio vel tortor laoreet vestibulum. Nulla facilisi. Proin pretium turpis sit amet quam egestas, at molestie sapien tempor. Ut euismod odio in risus eleifend.",
  "isPublished": true,
  "featuredImageURL": "time.jpg"
}' http://localhost:$APP_PORT/api/posts


curl -X POST -H "Authorization: Bearer $API_TOKEN" -H "Content-Type: application/json" -d '{
  "userID": 2,
  "categoryID": 1,
  "title": "The Art of Coding",
  "slug": "the-art-of-coding",
  "content": "Donec ullamcorper, turpis sit amet viverra facilisis, lorem ex vulputate purus, sed lacinia orci enim sit amet eros. Maecenas sit amet vehicula mi, sit amet fermentum leo. Nam nec sapien purus. Aenean sed est augue. Suspendisse sit amet dictum nunc, vitae pharetra velit. Praesent eu augue non purus fermentum consequat. Quisque non nisl vitae enim pretium ultricies.",
  "isPublished": true,
  "featuredImageURL": "coding.jpg"
}' http://localhost:$APP_PORT/api/posts

curl -X POST -H "Authorization: Bearer $API_TOKEN" -H "Content-Type: application/json" -d '{
  "userID": 2,
  "categoryID": 1,
  "title": "Adventures in AI",
  "slug": "adventures-in-ai",
  "content": "In at felis viverra, pellentesque lorem eget, laoreet dui. Integer non mauris sem. Integer sit amet nulla at dolor laoreet eleifend. Curabitur sit amet viverra nisl. Donec a lacus nec ligula malesuada luctus. Nulla vel ante a lacus commodo efficitur. Maecenas sit amet vehicula mi, sit amet fermentum leo. Nam nec sapien purus. Aenean sed est augue. Suspendisse sit amet dictum nunc, vitae pharetra velit.",
  "isPublished": true,
  "featuredImageURL": "ai.jpg"
}' http://localhost:$APP_PORT/api/posts

psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c 'SELECT * FROM posts'

# psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c 'TRUNCATE TABLE ROLES RESTART IDENTITY CASCADE'
 # psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c 'TRUNCATE TABLE USERS RESTART IDENTITY CASCADE'
 # psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c 'TRUNCATE TABLE posts RESTART IDENTITY CASCADE'
 # psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c 'TRUNCATE TABLE categories RESTART IDENTITY CASCADE'