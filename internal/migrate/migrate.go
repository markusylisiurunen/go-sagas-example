package migrate

import (
	"context"

	"github.com/golang-migrate/migrate/v4"

	// these seem to be required according to: https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(_ context.Context, folder string, url string) error {
	sourceURL, databaseURL := "file://"+folder, url
	migrator, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return err
	}
	defer migrator.Close()
	if err := migrator.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return err
	}
	return nil
}
