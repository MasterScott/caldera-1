package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"{{[ .Project ]}}/pkg/db"
	{{[- if .Prometheus.Enabled ]}}
	"{{[ .Project ]}}/pkg/metrics"
	{{[- end ]}}
	"{{[ .Project ]}}/pkg/db/migrations"
	{{[- if .Example ]}}
	"{{[ .Project ]}}/pkg/db/provider"
	{{[- end ]}}

	// MySQL driver.
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

const (
	// Driver defines database driver name.
	Driver = "mysql"
)

// MySQL controls mysql driver connection and providers.
type MySQL struct {
	pool *sql.DB
	cfg  *db.Config
	log  *zap.Logger
	{{[- if .Example ]}}
	// Contract providers
	events provider.Events
	{{[- end ]}}
}

// DSN creates dsn type connection.
func DSN(cfg *db.Config) *db.Config {
	if cfg.DSN == "" {
		var properties string

		if len(cfg.Properties) > 0 {
			properties = "?" + strings.Join(cfg.Properties, "&")
		}

		dsn := fmt.Sprintf("%s://%s:%s@%s:%d/%s%s",
			cfg.Driver, cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Name, properties)
		cfg.DSN = dsn
	}

	return cfg
}

// New creates new postgres DB connection.
func New(cfg *db.Config, log *zap.Logger, mig migrations.Migrator) (*MySQL, error) {
	m := &MySQL{
		cfg: cfg,
		log: log,
	}

	var err error

	m.pool, err = db.Connect(cfg)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	name := cfg.Driver

	if err := m.pool.QueryRow("SELECT version()").Scan(&name); err != nil {
		return nil, fmt.Errorf("failed to check the database engine version: %w", err)
	}

	log.Info("DB", zap.String("version", name))

	{{[- if .Example ]}}

	m.events = newEventsProvider(m.pool)
	{{[- end ]}}

	// setup migration connection
	mig.Setup(m.pool)

	return m, mig.Migrate()
}

// Check readiness for database.
func (m MySQL) Check() error {
	return m.pool.Ping()
}

// Shutdown process graceful shutdown for the server.
func (m MySQL) Shutdown(ctx context.Context) error {
	return m.pool.Close()
}

{{[- if .Example ]}}

// EventsProvider returns data store provider for Events.
func (m MySQL) EventsProvider() provider.Events {
	return m.events
}
{{[- end ]}}

{{[- if .Prometheus.Enabled ]}}

// MetricFunc returns a func to monitor connectivity to MySQL.
func (m MySQL) MetricFunc() metrics.MetricFunc {
	return metrics.DBMetricFunc(m.cfg.Host, m.cfg.Name, m.pool.Stats())
}
{{[- end ]}}
