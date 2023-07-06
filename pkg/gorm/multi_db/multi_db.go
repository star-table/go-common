package multi_db

import (
	"context"

	"github.com/star-table/go-common/go-common/pkg/middleware/meta"
	"gorm.io/gorm"
)

const DefaultDatasource = "default"

type MultiDB struct {
	dbName string
	dbs    map[string]*gorm.DB
}

func NewMultiDB(dbName string, dbs map[string]*gorm.DB) *MultiDB {
	return &MultiDB{
		dbName: dbName,
		dbs:    dbs,
	}
}

func (m *MultiDB) GetDBs() map[string]*gorm.DB {
	newDbs := make(map[string]*gorm.DB, len(m.dbs))
	for s, db := range m.dbs {
		newDbs[s] = db
	}
	return newDbs
}

func (m *MultiDB) GetDBName() string {
	return m.dbName
}

func (m *MultiDB) WithContext(ctx context.Context) *gorm.DB {
	return m.GetDatasource(ctx).WithContext(ctx)
}

func (m *MultiDB) GetDatasource(ctx context.Context) *gorm.DB {
	ch := meta.GetCommonHeaderFromCtx(ctx)

	db := m.dbs[ch.Datasource]
	if db == nil {
		return m.dbs[DefaultDatasource]
	}

	return db
}
