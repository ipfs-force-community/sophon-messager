package mysql

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm/utils/tests"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/filecoin-project/venus-messager/models/repo"
)

type anyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a anyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func wrapper(f func(*testing.T, repo.Repo, sqlmock.Sqlmock), repo repo.Repo, mock sqlmock.Sqlmock) func(t *testing.T) {
	return func(t *testing.T) {
		f(t, repo, mock)
	}
}

func setup(t *testing.T) (repo.Repo, sqlmock.Sqlmock, *sql.DB) {
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	mock.ExpectQuery("SELECT VERSION()").WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(""))

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}))
	assert.NoError(t, err)

	return Repo{DB: gormDB}, mock, sqlDB
}

func closeDB(mock sqlmock.Sqlmock, sqlDB *sql.DB) error {
	mock.ExpectClose()
	return sqlDB.Close()
}

var db, _ = gorm.Open(tests.DummyDialector{}, nil)
var timeT = reflect.TypeOf(time.Time{})

func genInsertSQL(obj interface{}) (string, []driver.Value) {
	tabler, ok := obj.(schema.Tabler)
	if !ok {
		panic("not implement schema.Tabler")
	}

	objSchema, _ := schema.Parse(obj, &sync.Map{}, db.NamingStrategy)
	insertValues := reflect.ValueOf(obj)
	var insertCols []string
	var insertArgs []driver.Value
	for _, dbName := range objSchema.DBNames {
		field := objSchema.LookUpField(dbName)
		if field.FieldType == timeT {
			insertCols = append(insertCols, field.DBName)
			insertArgs = append(insertArgs, anyTime{})
			continue
		}
		fieldV, _ := field.ValueOf(insertValues)
		insertCols = append(insertCols, field.DBName)
		insertArgs = append(insertArgs, fieldV)
	}

	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("INSERT INTO `%s` (", tabler.TableName()))
	for index, col := range insertCols {
		if index == len(insertCols)-1 {
			buf.WriteString(fmt.Sprintf("`%s`", col))
		} else {
			buf.WriteString(fmt.Sprintf("`%s`,", col))
		}
	}
	buf.WriteString(") VALUES (")
	buf.WriteString(strings.TrimRight(strings.Repeat("?,", len(insertCols)), ","))
	buf.WriteString(")")
	return buf.String(), insertArgs
}

func genUpdateSQL(obj interface{}, skipZero bool, where ...string) (string, []driver.Value) {
	tabler, ok := obj.(schema.Tabler)
	if !ok {
		panic("not implement schema.Tabler")
	}

	objSchema, _ := schema.Parse(obj, &sync.Map{}, db.NamingStrategy)
	updatingValue := reflect.ValueOf(obj)
	var updateCols []string
	var updateArgs []driver.Value
	for _, dbName := range objSchema.DBNames {
		field := objSchema.LookUpField(dbName)
		if field.PrimaryKey {
			continue
		}
		if field.FieldType == timeT {
			updateCols = append(updateCols, field.DBName)
			updateArgs = append(updateArgs, anyTime{})
			continue
		}
		if fieldV, isZero := field.ValueOf(updatingValue); !(skipZero && isZero) {
			updateCols = append(updateCols, field.DBName)
			updateArgs = append(updateArgs, fieldV)
		}
	}

	buf := &bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("UPDATE `%s` SET ", tabler.TableName()))

	for index, col := range updateCols {
		if index == len(updateCols)-1 {
			buf.WriteString(fmt.Sprintf("`%s`=?", col))
		} else {
			buf.WriteString(fmt.Sprintf("`%s`=?,", col))
		}
	}

	if len(where) == 0 {
		for _, pri := range objSchema.PrimaryFields {
			where = append(where, pri.DBName)
		}
	}

	for index, wh := range where {
		if index == 0 {
			buf.WriteString(fmt.Sprintf(" WHERE `%s` = ?", wh))
		} else {
			buf.WriteString(fmt.Sprintf(" AND `%s` = ?", wh))
		}
	}
	return buf.String(), updateArgs
}
