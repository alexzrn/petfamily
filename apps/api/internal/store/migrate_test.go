package store

import (
	"os"
	"testing"
)

func TestSplitStatements(t *testing.T) {
	body := "-- комментарий\nCREATE TABLE a (id INT);\n\nCREATE TABLE b (id INT);\n"
	got := SplitStatements(body)
	if len(got) != 2 {
		t.Fatalf("операторов %d, ожидали 2: %#v", len(got), got)
	}
	if got[0] != "CREATE TABLE a (id INT)" {
		t.Fatalf("первый оператор %q", got[0])
	}
}

func TestMigrate(t *testing.T) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		t.Skip("MYSQL_DSN не задан")
	}
	db, err := Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// чистая база: сносим всё, что осталось от предыдущих прогонов
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE()")
	if err != nil {
		t.Fatal(err)
	}
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, name)
	}
	rows.Close()
	for _, name := range tables {
		if _, err := db.Exec("DROP TABLE IF EXISTS `" + name + "`"); err != nil {
			t.Fatal(err)
		}
	}

	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	// повторный прогон должен быть безопасным
	if err := Migrate(db); err != nil {
		t.Fatalf("повторная миграция: %v", err)
	}

	want := map[string]bool{"users": false, "sessions": false, "pets": false, "walks": false, "walk_points": false, "meetups": false}
	rows, err = db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE()")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		if _, ok := want[name]; ok {
			want[name] = true
		}
	}
	for name, ok := range want {
		if !ok {
			t.Fatalf("таблица %s не создана", name)
		}
	}
}
