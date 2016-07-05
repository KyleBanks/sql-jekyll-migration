# sql-jekyll-migration

This tool is a Go program to migrate tables in a SQL database to Jekyll collections, by creating a collection page in a Jekyll project for each row in the target tables.

Note: Currently the tool only supports Postgres, but adding support for MySQL, SQLite, other others should be rather straightforward.

## Requirements

Before running, you'll need to set a few environment variables to tell the program where the Postgres database is. For example:

```
export DBHOST="localhost"
export DBNAME="blog"
export DBUSER="username"
export DBPASS="password"
```

## Running

Once the environment variables are set, we can run the program like so:

```
go run migrate.go <TableName> <CollectionPath> <FileNameKey> <ContentKey> <DateKey> <FrontMatter>
```

- **TableName:** the name of the table to migrate in the database. *ex. "posts"*
- **CollectionPath:** the path to the collection directory in your Jekyll project. *ex. "~/Code/kylewbanks.com/_posts"*
- **FileNameKey:** the name of the database column that contains the value to be used as the filename for each row. The name will be lower-cased, non-alphanumeric characters removed, and spaces wil be replaced with hyphens. *ex. "title"*
- **ContentKey:** the name of the database column that contains the content of the collection file. If you pass a hyphen (`-`) instead of a column name, no content will be added to the file.
- **DateKey:** the name of the database column that contains the date to be used as the prefix for the filename. If you pass a hyphen (`-`), no date will be prefixed.
- **FrontMatter:** contains a colon (`:`) delimited list of key-value pairs, where the key is the database column name, and the value is the Jekyll front matter name.
    - For example, given the following FrontMatter:
        ```
        post_title=title:post_color=color
        ```
        We can expect the following Jekyll front matter to be generated at the top of each file in the new collection:
        ```
        title: <post_title value>
        color: <post_color value>
        ```
    - **An Important Note:** Whatever values are used for the FileNameKey, ContentKey, and DateKey (except `-`), they MUST also appear as keys in the FrontMatter.
    - If a hyphen (`-`) is used as the key, the value will not be added to the front-matter. This is useful, for example, because you must specify the ContentKey in FrontMatter, but you probably don't actually want it added to the front-matter YAML in the generated file.

## Example

For an example of how to use the tool to perform a migration, I've added [migrate_kylewbanks.sh](./example/migrate_kylewbanks.sh) to the `example/` directory that shows how I used the tool to migrate the posts and projects from Postgres to Jekyll for my [blog](http://kylewbanks.com).

## License

```
The MIT License (MIT)

Copyright (c) 2016 Kyle Banks

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

```
