api that allow users to upsert quotes
tech: go , github.com/mattn/go-sqlite3 github.com/labstack/echo/v4
users will need to authenticate via email password login

db
table user
col uuid id
col string email
col string username
col string hashed_password
col boolean authenticated
col datetime created_at

table group
col string id
col string group_id (not unique)
col string name
col string member_id
col datetime created_at
col datetime updated_at

table quote
col uuid id
col string text
col string author
col string uploader_id
col string group_id
col datetime created_at
col datetime updated_at

routes

POST /quote
  query params: ?author ?page=1 ? limit=10
  body: { text: string, author: string }
GET /quote/random
  query params: ?author
GET /quote

PATCH /quote/:id
  body: { text: string, author: string }
