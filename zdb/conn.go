package zdb

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	_ "github.com/lib/pq"
)

const DriverName = "pgx/v5"

func newdb(o *Options, database string) (*gorm.DB, error) {
	db, err := gorm.Open(
		postgres.New(postgres.Config{
			DriverName: DriverName,
			DSN:        o.Dsn(database),
		}),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 使用单数表名
			},
			Logger: NewLogger(o),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("db %s open failed: %v", database, err)
	}
	d, _ := db.DB()
	d.SetMaxIdleConns(o.MaxIdle)
	d.SetMaxOpenConns(o.MaxAlive)
	d.SetConnMaxLifetime(o.MaxAliveLife)
	if o.Debug != nil && *o.Debug {
		db = db.Debug()
	}
	return db, nil
}
