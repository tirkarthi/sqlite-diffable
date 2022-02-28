### sqlite-diffable

A golang port of [sqlite-diffable](https://github.com/simonw/sqlite-diffable/)


### Usage

```
./sqlite-diffable dump --help
Dump sqlite database metadata and table

Usage:
  sqlite-diffable dump [tables] [flags]

Flags:
      --all             Dump all tables
  -h, --help            help for dump
  -o, --output string   Output directory
  -p, --path string     Path to sqlite database

```

Dump all tables along with their metdata as JSON

```
sqlite-diffable dump --path dev.db --output /tmp/output --all
```

Dump only given tables as JSON

```
sqlite-diffable dump --path dev.db --output /tmp/output Post
```

### Format

Post.ndjson

```
["ckzxtl2as0000atje6pdq15o7","2022-02-22T07:42:19.108Z","2022-02-22T07:42:19.108Z","Hi from Prisma!",true,"Prisma is a database toolkit and makes databases easy.","Prisma is a database toolkit and makes databases easy. Long Description"]
["ckzxtl3s90000hfjedovzyl6i","2022-02-22T07:42:21.033Z","2022-02-22T07:42:21.033Z","Hi from Prisma!",true,"Prisma is a database toolkit and makes databases easy.","Prisma is a database toolkit and makes databases easy. Long Description"]
```

Post.metadata.json

```
{
    "columns": [
        "id",
        "createdAt",
        "updatedAt",
        "title",
        "published",
        "desc",
        "long_desc"
    ],
    "name": "Post",
    "schema": "CREATE TABLE \"Post\" (\n    \"id\" TEXT NOT NULL PRIMARY KEY,\n    \"createdAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,\n    \"updatedAt\" DATETIME NOT NULL,\n    \"title\" TEXT NOT NULL,\n    \"published\" BOOLEAN NOT NULL,\n    \"desc\" TEXT\n, \"long_desc\" TEXT)"
}
```
