package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
)

func ProcessSchemaChanges(ctx context.Context, databaseFiles embed.FS) error {
	initExists, err := doesInitTableExist(ctx)
	if err != nil {
		return fmt.Errorf("error has occurred when checking if database initialization is needed: %v", err)
	}

	if !*initExists {
		log.Println("database is not initialized, attempting to init ...")
		err = createInitTable(ctx)
		if err != nil {
			return fmt.Errorf("error occurred when attempting to create init table: %v", err)
		}
		log.Println("database initialization complete")
	}

	log.Println("checking for migrations ...")
	dirEntries, err := fs.ReadDir(databaseFiles, DatabaseMigrationDirectory)
	if err != nil {
		return fmt.Errorf("an error has occurred when attempting to read database directory. Error: %v", err)
	}
	var migrationFileCandidateFileNames []string
	for _, entry := range dirEntries {
		if !entry.IsDir() {
			migrationFileCandidateFileNames = append(migrationFileCandidateFileNames, entry.Name())
		}
	}

	migrationFiles := filterForMigrationFiles(migrationFileCandidateFileNames)
	var migrationsCompleted []string
	noMigrationsToProcessMessage := "no database migration files to process, skipping migrations ..."
	if len(migrationFiles) == 0 {
		log.Println(noMigrationsToProcessMessage)
		return nil
	} else {
		migrationsCompleted, err = checkForCompletedMigrations(ctx)
		if err != nil {
			return fmt.Errorf("error has occurred when checking for completed migrations: %v", err)
		}
	}

	migrationsNeeded := determineMigrationsNeeded(migrationFiles, migrationsCompleted)
	migrationsNeededSorted := sortMigrationsNeededFiles(migrationsNeeded)
	for _, fileName := range migrationsNeededSorted {
		log.Printf("attempting to perform database migration with %s ...", fileName)

		filePath := fmt.Sprintf("%s/%s", DatabaseMigrationDirectory, fileName)
		err = executeSQLFile(filePath, databaseFiles)
		if err != nil {
			return fmt.Errorf("error occurred when executing sql script: Filename: %s. Error: %v", fileName, err)
		}
		err = recordSuccessfulMigration(ctx, fileName)
		if err != nil {
			return fmt.Errorf("error has occurred when attempting to record a successful migration: %v", err)
		}
	}
	log.Println("finished database schema changes")
	return nil
}

func createInitTable(ctx context.Context) error {
	tx, err := ConnectionPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error, when attempting to start a transaction: %v", err)
	}

	err = func() error {
		_, err = tx.Exec(ctx, "create table init\n(\n    id                  SERIAL not null\n        constraint init_pk\n            primary key,\n    migration_file_name text   not null\n)")
		if err != nil {
			return fmt.Errorf("error, when executing query to create init table: %v", err)
		}

		_, err = tx.Exec(ctx, "comment on table init is 'This table is for tracking which schema migration scripts have been applied already'")
		if err != nil {
			return fmt.Errorf("error, when attempting to add a comment to the init table: %v", err)
		}

		_, err = tx.Exec(ctx, "create unique index init_migration_file_name_uindex\n    on init (migration_file_name)")
		if err != nil {
			return fmt.Errorf("error, when attempting to create a unique index for the init table")
		}
		return nil
	}()
	if err != nil {
		rollBackErr := tx.Rollback(ctx)
		if rollBackErr != nil {
			return fmt.Errorf("error, when attempting to roll back commit: Rollback Error: %v, Original Error: %v", rollBackErr, err)
		}
		return fmt.Errorf("error, when attempting to perform database transaction: %v", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error, when attempting to commit the transaction to the database: %v", err)
	}
	return nil
}

func sortMigrationsNeededFiles(needed []string) []string {
	re := regexp.MustCompile(`^(\d+)`)

	sort.Slice(needed, func(i, j int) bool {
		num1, _ := strconv.Atoi(re.FindStringSubmatch(needed[i])[1])
		num2, _ := strconv.Atoi(re.FindStringSubmatch(needed[j])[1])
		return num1 < num2
	})
	return needed
}

func determineMigrationsNeeded(migrationFiles []string, migrationsCompleted []string) []string {
	var migrationsNeeded []string
	migrationsCompletedMap := make(map[string]bool)
	for _, value := range migrationsCompleted {
		migrationsCompletedMap[value] = true
	}
	for _, value := range migrationFiles {
		if !migrationsCompletedMap[value] {
			migrationsNeeded = append(migrationsNeeded, value)
		}
	}
	return migrationsNeeded
}

func filterForMigrationFiles(candidates []string) []string {
	var migrationFiles []string
	re := regexp.MustCompile(`^\d+`)
	for _, fileName := range candidates {
		if re.MatchString(fileName) {
			migrationFiles = append(migrationFiles, fileName)
		}
	}
	return migrationFiles
}

func recordSuccessfulMigration(ctx context.Context, fileName string) error {
	_, err := ConnectionPool.Exec(
		ctx,
		"INSERT INTO init (migration_file_name)\nVALUES ($1)",
		fileName,
	)
	if err != nil {
		return fmt.Errorf("error occurred when attempting to run sql command: %v", err)
	}
	return nil
}

func checkForCompletedMigrations(ctx context.Context) (results []string, err error) {
	var rows pgx.Rows
	rows, err = ConnectionPool.Query(
		ctx,
		"SELECT migration_file_name\nFROM init",
	)
	defer func() {
		err = rows.Err()
		if err != nil {
			err = fmt.Errorf("error, occurred when reading rows. Error: %v", err)
		}
		rows.Close()
	}()

	if err != nil {
		return nil, fmt.Errorf("error has occurred when attempting to retrieve pending migrations: %v", err)
	}

	for rows.Next() {
		var result string
		err = rows.Scan(
			&result,
		)
		if err != nil {
			return nil, fmt.Errorf("error has occurred when scanning for pending migrations: %v", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func doesInitTableExist(ctx context.Context) (*bool, error) {
	var result bool
	connConfig := ConnectionPool.Config().ConnConfig
	row := ConnectionPool.QueryRow(
		ctx,
		"SELECT EXISTS (\n    SELECT 1\n    FROM pg_tables\n    WHERE tablename = 'init'\n)",
	)
	err := row.Scan(
		&result,
	)
	if err != nil {
		return nil, fmt.Errorf("an error occurred when checking to see if database had been initialized. User: '%s' Error: '%v'", connConfig.User, err)
	}
	return &result, nil
}

func executeSQLFile(filePath string, databaseFiles embed.FS) error {
	content, err := databaseFiles.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	sql := string(content)
	queries := strings.Split(sql, ";")

	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}

		_, err = ConnectionPool.Exec(context.Background(), query)
		if err != nil {
			return fmt.Errorf("error, failed to execute QUERY: %s. ERROR: %v", query, err)
		}
	}

	return nil
}
