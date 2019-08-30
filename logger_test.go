package gormzerolog_test

import (
	"fmt"
	"github.com/Ahmet-Kaplan/gorm-zerolog"
	"github.com/Ahmet-Kaplan/gorm-zerolog/testhelper"
	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog"
	"os"
	"testing"
	"time"
)

var pool *testhelper.DockerPool

func TestMain(m *testing.M) {
	pool = testhelper.MustCreatePool()

	os.Exit(m.Run())
}

func Test_Logger_Postgres(t *testing.T) {
	z := zerolog.New(os.Stdout)
	zerologLogger := gormzerolog.New(&z,
		gormzerolog.WithLevel(zerolog.DebugLevel),
	/*		gormzerolog.WithRecordToFields(func(r gormzerolog.Record) gormzerolog.RecordToFields {
			return gormzerolog.WithRecordToFields()
				zap.String("caller", r.Source),
				zap.Float32("duration_ms", float32(r.Duration.Nanoseconds()/1000)/1000),
				zap.String("query", r.SQL),
				zap.Int64("rows_affected", r.RowsAffected),
			}
		}), */
	)

	defer func() {
		err := zerologLogger.Event.Send
		if err != nil {
			panic(err)
		}
	}()

	conn := pool.MustCreateDB(testhelper.DialectPostgres)
	defer conn.MustClose()

	now := time.Now()
	gorm.NowFunc = func() time.Time { return now }

	db, err := gorm.Open(conn.Dialect, conn.URL)
	if err != nil {
		panic(err)
	}

	type Post struct {
		Title, Body string
		CreatedAt   time.Time
	}
	db.AutoMigrate(&Post{})

	cases := []struct {
		run    func() error
		sql    string
		values []string
	}{
		{
			run: func() error { return db.Create(&Post{Title: "awesome"}).Error },
			sql: fmt.Sprintf(
				"INSERT INTO %q (%q,%q,%q) VALUES (%v,%v,%v) RETURNING %q.*",
				"posts", "title", "body", "created_at", "awesome", "", now.String(),
				"posts",
			),
			values: []string{"awesome", "", now.String()},
		},
		{
			run:    func() error { return db.Model(&Post{}).Find(&[]*Post{}).Error },
			sql:    "SELECT * FROM \"posts\"  ",
			values: []string{},
		},
		{
			run: func() error {
				return db.Where(&Post{Title: "awesome", Body: "This is awesome post !"}).First(&Post{}).Error
			},
			sql: fmt.Sprintf(
				"SELECT * FROM %q  WHERE (%q = %v) AND (%q = %v) LIMIT 1",
				"posts", "title", "awesome", "body", "This is awesome post !",
			),
			values: []string{"awesome", "This is awesome post !"},
		},
	}

	db.SetLogger(zerologLogger)
	db.LogMode(true)

	for _, c := range cases {
		err := c.run()
		if err != nil && err != gorm.ErrRecordNotFound {
			t.Fatalf("Unexpected error: %v", err)
		}

		// TODO: Must get from log entries
		entries := map[string]interface{}{}

		if got, want := len(entries), 1; got != want {
			t.Errorf("Logger logged %d items, want %d items", got, want)
		}

		fieldByName := entries

		if got, want := fieldByName["sql"].(string), c.sql; got != want {
			t.Errorf("Logged sql was %q, want %q", got, want)
		}

		if got, want := len(fieldByName["values"].([]interface{})), len(c.values); got != want {
			t.Errorf("Logged values has %d items, want %d items", got, want)
		}

		for i, want := range c.values {
			got := fieldByName["values"].([]interface{})[i].(string)
			if got != want {
				t.Errorf("Logged values at %d was %v, want %v", i, got, want)
			}
		}
	}
}
