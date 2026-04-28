package cassandra

import (
	"MusicStreamingHistoryService/internal/config"
	"fmt"
	"time"

	"github.com/gocql/gocql"
)

func NewSession(cfg config.CassandraConfig) (*gocql.Session, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 5 * time.Second
	cluster.ConnectTimeout = 10 * time.Second

	if cfg.Username != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create Cassandra session: %w", err)
	}

	return session, nil
}

func EnsureKeyspaceIsCreated(cfg config.CassandraConfig) error {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 5 * time.Second

	if cfg.Username != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to connect to Cassandra: %w", err)
	}
	defer session.Close()

	query := fmt.Sprintf(`
		CREATE KEYSPACE IF NOT EXISTS %s
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`,
		cfg.Keyspace,
	)

	if err = session.Query(query).Exec(); err != nil {
		return fmt.Errorf("failed to create keyspace: %w", err)
	}

	return nil
}
