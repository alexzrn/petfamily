// Package testsupport готовит чистую базу для тестов.
package testsupport

import (
	"database/sql"

	"github.com/alexzrn/petfamily/apps/api/internal/store"
)

// Reset сносит все таблицы в текущей базе и накатывает миграции заново.
// Внешние ключи на время сноса выключаются, иначе порядок удаления имеет значение.
func Reset(db *sql.DB) error {
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return err
	}
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE()")
	if err != nil {
		return err
	}
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			rows.Close()
			return err
		}
		names = append(names, n)
	}
	rows.Close()
	for _, n := range names {
		if _, err := db.Exec("DROP TABLE IF EXISTS `" + n + "`"); err != nil {
			return err
		}
	}
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		return err
	}
	return store.Migrate(db)
}
