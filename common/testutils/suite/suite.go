package suite

import (
	"context"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/gorm"

	pg "github.com/Nurlan270/task-tracker-go/common/testutils/database/postgres"
)

type CommonSuite struct {
	suite.Suite

	//	Database related
	db     *gorm.DB
	pgcCfg *pg.ContainerConfig
	pgc    *postgres.PostgresContainer
	models []any
}

func (s *CommonSuite) SetupCommonSuite(ctx context.Context) {
	//	Start postgres container
	s.pgc = pg.NewDBContainer(ctx, s.T(), s.pgcCfg)

	//	Get DB instance
	s.db = pg.NewDB(ctx, s.T(), s.pgc, s.pgcCfg)
}

func (s *CommonSuite) SetupCommonTest() {
	pg.MigrateDB(s.T(), s.DB(), s.models)
}

func (s *CommonSuite) TearDownCommonTest() {
	pg.TruncateDB(s.T(), s.DB(), s.models)
}

func (s *CommonSuite) SetPostgresContainerConfig(cfg *pg.ContainerConfig) {
	s.pgcCfg = cfg
}

func (s *CommonSuite) SetMigrationModels(models []any) {
	s.models = models
}

func (s *CommonSuite) DB() *gorm.DB {
	return s.db
}
