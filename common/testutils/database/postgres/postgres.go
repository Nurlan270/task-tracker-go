package postgres

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/gorm"

	"github.com/Nurlan270/task-tracker-go/common/database"
)

type ContainerConfig struct {
	Name     string
	User     string
	Password string

	// Img defines docker image of postgres container (app uses "postgres:18-alpine")
	Img string
}

func NewDBContainer(ctx context.Context, t *testing.T, cfg *ContainerConfig) *postgres.PostgresContainer {
	t.Helper()

	pgc, err := postgres.Run(ctx, cfg.Img,
		postgres.WithDatabase(cfg.Name),
		postgres.WithUsername(cfg.User),
		postgres.WithPassword(cfg.Password),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %s", err)
	}

	t.Cleanup(func() {
		t.Logf("terminating postgres container")

		if err = testcontainers.TerminateContainer(pgc); err != nil {
			t.Logf("failed to terminate postgres container: %s", err)
		}
	})

	return pgc
}

func NewDB(
	ctx context.Context,
	t *testing.T,
	pgc *postgres.PostgresContainer,
	cfg *ContainerConfig,
) *gorm.DB {
	t.Helper()

	//	Get Host/Port pair from container
	host, port, err := splitEndpoint(ctx, pgc)
	if err != nil {
		t.Fatalf("failed to split postgres host/port pair: %s", err)
	}

	conn := database.Config{
		Host:     host,
		Port:     port,
		Name:     cfg.Name,
		User:     cfg.User,
		Password: cfg.Password,
	}

	db, err := database.New(conn)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %s", err)
	}

	return db
}

func MigrateDB(t *testing.T, db *gorm.DB, models []any) {
	t.Helper()

	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("failed to migrate models: %s", err)
	}
}

func TruncateDB(t *testing.T, db *gorm.DB, models []any) {
	t.Helper()

	if err := db.Migrator().DropTable(models...); err != nil {
		t.Fatalf("failed to drop models: %s", err)
	}
}

func splitEndpoint(ctx context.Context, pgc *postgres.PostgresContainer) (string, uint16, error) {
	hostPort, err := pgc.Endpoint(ctx, "")
	if err != nil {
		return "", 0, fmt.Errorf("failed to retrieve host/port pair: %s", err)
	}

	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse host/port pair: %s", err)
	}

	portUint64, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return "", 0, fmt.Errorf("failed to convert port (str -> uint64): %s", err)
	}

	return host, uint16(portUint64), nil
}
