package sqlite

import (
	"database/sql"
	"fmt"
)

type tableCreator func(tx *sql.Tx) error

func initialize(tx *sql.Tx) (bool, error) {
	tables, err := getTableCreators(tx)

	if err != nil {
		return false, err
	}

	if len(tables) == 0 {
		return false, nil
	}

	for _, table := range tables {
		if err = table(tx); err != nil {
			break
		}
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func getTableCreators(tx *sql.Tx) (map[string]tableCreator, error) {
	rows, err := tx.Query("SELECT name FROM sqlite_master WHERE type='table'")

	if err != nil {
		return nil, err
	}

	tables := make(map[string]tableCreator)
	tables[targetTableName] = createTargetsTable
	tables[subscriberTableName] = createSubscribersTable
	tables[targetSubscriberTableName] = createTargetSubscriberTable

	for rows.Next() {
		var name string

		err = rows.Scan(&name)

		if err != nil {
			break
		}

		switch name {
		case targetTableName:
			delete(tables, targetTableName)
		case subscriberTableName:
			delete(tables, subscriberTableName)
		case targetSubscriberTableName:
			delete(tables, targetSubscriberTableName)
		}
	}

	if err != nil {
		return nil, err
	}

	return tables, nil
}

func execQueries(tx *sql.Tx, queries []string) error {
	var err error

	for _, query := range queries {
		if _, err = tx.Exec(query); err != nil {
			break
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func createTargetsTable(tx *sql.Tx) error {
	return execQueries(tx, []string{
		fmt.Sprintf(
			"CREATE TABLE %s("+
				"id INTEGER NOT NULL PRIMARY KEY,"+
				"key TEXT NOT NULL,"+
				"name TEXT NOT NULL,"+
				"kind TEXT NOT NULL,"+
				"enabled INTEGER NOT NULL"+
				");",
			targetTableName,
		),
		fmt.Sprintf(
			"CREATE UNIQUE INDEX target_key_idx on %s(key);",
			targetTableName,
		),
		fmt.Sprintf(
			"CREATE UNIQUE INDEX target_name_idx on %s(name);",
			targetTableName,
		),
	})
}

func createSubscribersTable(tx *sql.Tx) error {
	return execQueries(tx, []string{
		fmt.Sprintf(
			"CREATE TABLE %s("+
				"id INTEGER NOT NULL PRIMARY KEY,"+
				"name TEXT NOT NULL,"+
				"event TEXT NOT NULL,"+
				"method TEXT NOT NULL,"+
				"url TEXT NOT NULL,"+
				"headers TEXT,"+
				"data TEXT"+
				");",
			subscriberTableName,
		),
		fmt.Sprintf(
			"CREATE UNIQUE INDEX subscriber_name_idx on %s(name);",
			subscriberTableName,
		),
	})
}

func createTargetSubscriberTable(tx *sql.Tx) error {
	return execQueries(tx, []string{
		fmt.Sprintf(
			"CREATE TABLE %s("+
				"event TEXT NOT NULL,"+
				"enabled INTEGER NOT NULL,"+
				"target_id INTEGER REFERENCES %s(id) ON DELETE CASCADE,"+
				"subscriber_id INTEGER REFERENCES %s(id) ON DELETE CASCADE"+
				");",
			targetSubscriberTableName,
			targetTableName,
			subscriberTableName,
		),
		fmt.Sprintf(
			"CREATE INDEX target_subscriber_idx on %s(target_id, subscriber_id);",
			targetSubscriberTableName,
		),
	})
}