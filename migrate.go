package main

import (
	"os"
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
	"strings"
	"io/ioutil"
	"regexp"
	"time"
)

const (
	NumArgs = 6

	// Value to denote that an argument should be ignored
	Ignore = "-"
)

var (
	db *sql.DB

	FileNameSanitizer = regexp.MustCompile("[^A-Za-z0-9\\-]+")
)

func main() {
	// Open database connection
	db = openConnection(os.Getenv("DBHOST"), os.Getenv("DBNAME"), os.Getenv("DBUSER"), os.Getenv("DBPASS"))
	defer db.Close()

	// Parse arguments
	table, collectionPath, fileNameKey, contentKey, dateKey, frontMatter := args()

	// Perform some sanity before we migrate
	querySelect := constructQuerySelect(frontMatter)
	ensureDirectoryExists(collectionPath)
	ensureTableIsAccessible(table, querySelect)

	// Everything looks good, start processing
	performMigration(table, collectionPath, fileNameKey, contentKey, dateKey, frontMatter)
}

// openConnection
func openConnection(host, database, username, password string) *sql.DB {
	fmt.Printf("Opening connection: {Host: %v, DBName: %v, User: %v, Password: %v}\n", host, database, username, password)

	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=disable", username, password, host, database))
	if err != nil {
		fail(err)
	}

	return db
}

// ensureTableIsAccessible performs a simple count on the table to ensure we can access it, and to provide some debug logging.
// If the count succeeds, it will also attempt to select a single row from the table using the `querySelect` just to ensure there is no error.
func ensureTableIsAccessible(table string, querySelect string) {
	// Perform a COUNT
	rows, err := db.Query("SELECT COUNT(*) AS count FROM " + table)
	if err != nil {
		fail(err)
	}
	defer rows.Close()

	for rows.Next() {
		var count int64
		if err := rows.Scan(&count); err != nil {
			fail(err)
		}

		fmt.Printf("Processing %v rows...\n", count)
	}

	// Test the querySelect
	if rows, err := db.Query(querySelect + " FROM " + table + " LIMIT 1"); err != nil {
		fail(err)
	} else {
		rows.Close()
	}
}

// ensureDirectoryExists panics if the directory does not exist.
func ensureDirectoryExists(dir string) {
	if  _, err := os.Stat(dir); err != nil {
		fail(err)
	}
}

// args parses and returns the required arguments for program execution.
func args() (table, collectionPath, fileNameKey, contentKey, dateKey string, frontMatterValues []frontMatter) {
	// Validate argument length
	args := os.Args[1:]
	if len(args) < NumArgs {
		fail(fmt.Errorf("Invalid number of arguments provided, %v required.\n", NumArgs))
	}

	// Retrieve and return the arguments
	table = args[0]
	collectionPath = args[1]
	fileNameKey = args[2]
	contentKey = args[3]
	dateKey = args[4]
	frontMatterStr := args[5]

	// Sanitize args as required
	if !strings.HasSuffix(collectionPath, "/") {
		collectionPath = collectionPath + "/"
	}
	for _, val := range strings.Split(frontMatterStr, ":") {
		frontMatterValues = append(frontMatterValues, frontMatter{
			DbColumn: strings.Split(val, "=")[0],
			JekyllKey: strings.Split(val, "=")[1],
		})
	}

	// Finally, return
	fmt.Printf("Args: {Table: %v, CollectionPath: %v, FileNameKey: %v, ContentKey: %v, DateKey: %v, FrontMatter: %v}\n", table, collectionPath, fileNameKey, contentKey, dateKey, frontMatterValues)
	return table, collectionPath, fileNameKey, contentKey, dateKey, frontMatterValues
}

// constructQuerySelect returns the SELECT portion of the query based on the frontMatter values provided.
func constructQuerySelect(props []frontMatter) (selectQuery string) {
	s := "SELECT"
	for i, p := range props {
		var delimiter string
		if i > 0 {
			delimiter = ","
		}

		s = s + fmt.Sprintf("%v %v", delimiter, p.DbColumn)
	}

	return s
}

// performMigration queries the table and processes each row as a collection item.
func performMigration(table, collectionPath, fileNameKey, contentKey, dateKey string, frontMatter []frontMatter) {
	query := constructQuerySelect(frontMatter) + " FROM " + table
	fmt.Println("Querying: ", query)

	rows, err := db.Query(query)
	if err != nil {
		fail(err)
	}
	defer rows.Close()

	for rows.Next() {
		fileName, contents := processRow(rows, fileNameKey, contentKey, dateKey, frontMatter)
		if len(fileName) == 0 {
			fmt.Println("WARNING: Skipping row because the file name could not be determined: ", contents)
			continue
		}

		fmt.Println(fileName, "writing...")
		err := ioutil.WriteFile(collectionPath + fileName, []byte(contents), 0644)
		if err != nil {
			fail(err)
		}
		fmt.Println(fileName, "complete")
	}
}

// processRow handles a single row from the rows provided, and returns the generated contents of a single collection file,
// as well as the name of the file.
func processRow(rows *sql.Rows, fileNameKey string, contentKey string, dateKey string, frontMatter []frontMatter) (fileName, contents string) {
	// Dynamically load the data into a map
	columns := make([]interface{}, len(frontMatter))
	columnPointers := make([]interface{}, len(frontMatter))
	for i, _ := range columns {
		columnPointers[i] = &columns[i]
	}

	if err := rows.Scan(columnPointers...); err != nil {
		fail(err)
	}

	// Iterate over the map and add to the front-matter string
	var contentBody string
	date := time.Unix(0, 0)
	for i, f := range frontMatter {
		k := f.DbColumn
		val := columns[i]
		if val == nil {
			continue
		}

		// Check for special columns that can contain the filename/content body
		if k == fileNameKey {
			fileName = fmt.Sprintf("%v", val)
			fileName = strings.TrimSpace(fileName)
			if len(fileName) == 0 {
				return "", ""
			}

			fileName = strings.Replace(fileName, " ", "-", -1)
			fileName = strings.ToLower(fileName)
			fileName = FileNameSanitizer.ReplaceAllString(fileName, "")
			fileName = fileName + ".md"
		} else if k == contentKey {
			contentBody = fmt.Sprintf("%v", val)
		} else if k == dateKey {
			date = val.(time.Time)
		}

		// Check if we should ignore this key
		if f.JekyllKey == Ignore {
			continue
		}

		// Determine how to display the value based on the type
		var s string
		switch val.(type) {
		default:
			s = fmt.Sprintf("%v", val)
		case string:
			s = fmt.Sprintf("\"%v\"", strings.Replace(fmt.Sprintf("%v", val), `"`, `\"`, -1))
		}

		contents = fmt.Sprintf("%v%v: %v \n", contents, f.JekyllKey, s)
	}

	contents = fmt.Sprintf("---\n%v---\n%v", contents, contentBody)
	if date.Unix() > 0 {
		fileName = fmt.Sprintf("%v-%v", date.Format("2006-01-02"), fileName)
	}

	return fileName, contents
}

// fail logs an error and panics, stopping program execution.
func fail(err error) {
	fmt.Println("Execution failed: ", err)
	panic(err)
}


// --- START frontMatter

type frontMatter struct {
	DbColumn  string
	JekyllKey string
}

func (m frontMatter) String() string {
	return fmt.Sprintf("{Column: %v, JekyllKey: %v}", m.DbColumn, m.JekyllKey)
}


// --- END frontMatter