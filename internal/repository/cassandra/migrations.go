package cassandra

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cassandra"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(hosts []string, keyspace string) error {
	connectionUrl := fmt.Sprintf(
		"cassandra://%s/%s?disable_host_lookup=true&consistency=quorum",
		hosts[0],
		keyspace,
	)
	fmt.Println(connectionUrl)
	wd, _ := os.Getwd()
	fmt.Println("Working dir:", wd)
	m, err := migrate.New("file://migrations", connectionUrl)
	if err != nil {
		return fmt.Errorf("failed to init migrator: %w", err)
	}
	defer m.Close()

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
