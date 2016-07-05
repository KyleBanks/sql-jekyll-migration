#!/bin/bash
#
# Sample used to migrate the kylewbanks.com project and post lists from Postrgres to MySQL.
#
# Usage:
#    ./example/migrate_kylewbanks.sh DBHOST DBNAME DBUSER DBPASS

# Set the required environment variables from the input provided
export DBHOST="$1"
export DBNAME="$2"
export DBUSER="$3"
export DBPASS="$4"

# Migrate Projects
go run migrate.go project ~/Code/kylewbanks.com/_projects title - - id=sortId:image_url=imageUrl:title=title:description_html=descriptionHtml:link_html=linkHtml:css=css:link_url=linkUrl

# Migrate Posts
go run migrate.go post ~/Code/kylewbanks.com/_posts title body created_at title=title:preview=preview:created_at=-:body=-:created_at=-:url=permalink:layout=layout