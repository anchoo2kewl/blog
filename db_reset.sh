psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_USER\?sslmode=disable -c "drop database $PG_DB"
psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_USER\?sslmode=disable -c "create database $PG_DB"

migrate -source file://migrations -database postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable up

psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c "INSERT INTO ROLES (role_name) values ('Commenter')"
psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c "INSERT INTO ROLES (role_name) values ('Administrator')"
psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c "INSERT INTO ROLES (role_name) values ('Editor')"
psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c "INSERT INTO ROLES (role_name) values ('Viewer')"

# go_port="USER INPUT"
read -p "Start Go server and enter port: " go_port

curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $API_TOKEN" -d '{
  "email": "anshuman@biswas.me",
  "username": "anshuman",
  "password": "123456qwertyu",
  "role": 1
}' http://localhost:$go_port/api/users

curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $API_TOKEN" -d '{
  "email": "anchoo2kewl@gmail.com",
  "username": "anchoo2kewl",
  "password": "123456qwertyu",
  "role": 2
}' http://localhost:$go_port/api/users

psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable -c "INSERT INTO CATEGORIES (category_name) values ('Demo Category')"

curl -X POST -H "Authorization: Bearer $API_TOKEN" -H "Content-Type: application/json" -d '{
  "userID": 2,
  "categoryID": 1,
  "title": "Fictitious Blog Post",
  "slug": "fictitious-blog-post",
  "content": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed nec enim a elit pretium suscipit. Nullam sit amet mauris nisi. Sed at placerat urna. Vivamus nec lectus ac orci dictum ullamcorper vitae eget lacus. Proin pretium turpis sit amet quam egestas, at molestie sapien tempor. Ut euismod odio in risus eleifend, at hendrerit lectus vehicula. Sed eget justo vel felis sollicitudin tincidunt. Nullam ut quam id eros mattis feugiat. Donec vulputate arcu vel nulla accumsan, et dignissim lorem malesuada. Donec quis justo ex. Phasellus scelerisque nunc id tellus sollicitudin, nec convallis ex varius. Suspendisse malesuada odio vel tortor laoreet vestibulum. Nulla facilisi.",
  "isPublished": true,
  "featuredImageURL": "image.jpg"
}' http://localhost:$go_port/api/posts