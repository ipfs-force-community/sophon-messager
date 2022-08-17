package mysql

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

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

func getStructFieldValue(obj interface{}) []driver.Value {
	rv := reflect.ValueOf(obj)
	rt := reflect.TypeOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
		rt = rt.Elem()
	}
	vals := make([]driver.Value, 0, rv.NumField())
	for i := 0; i < rv.NumField(); i++ {
		_, ok := rv.Field(i).Interface().(time.Time)
		if ok {
			vals = append(vals, anyTime{})
			continue
		}
		tagVal := rt.Field(i).Tag.Get(gormTag)
		isEmbedded := isEmbedded(tagVal)
		if !isEmbedded {
			vals = append(vals, rv.Field(i).Interface())
			continue
		}
		embeddedStruct := rv.Field(i).Elem()
		for j := 0; j < embeddedStruct.NumField(); j++ {
			_, ok := embeddedStruct.Field(j).Interface().(time.Time)
			if ok {
				vals = append(vals, anyTime{})
				continue
			}
			vals = append(vals, embeddedStruct.Field(j).Interface())
		}
	}
	return vals
}

var gormTag = "gorm"

func genUpdateSQL(obj interface{}) string {
	rt := reflect.TypeOf(obj)
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
		rt = rt.Elem()
	}
	tabler, ok := obj.(schema.Tabler)
	if !ok {
		panic("not implement schema.Tabler")
	}
	// id is primary key, eg. UPDATE `table_name` SET `name`=?,`addr`=? WHERE id = ?"
	primaryKey := ""
	filedLen := rv.NumField()
	buf := &bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("UPDATE `%s` SET ", tabler.TableName()))
	for i := 0; i < filedLen; i++ {
		tagVal := rt.Field(i).Tag.Get(gormTag)
		isEmbedded := isEmbedded(tagVal)
		if !isEmbedded {
			columnStr := getColumn(tagVal)
			if isPrimaryKey(tagVal) {
				primaryKey = columnStr
				continue
			}
			if i < filedLen-1 {
				buf.WriteString(fmt.Sprintf("`%s`=?,", columnStr))
			} else {
				buf.WriteString(fmt.Sprintf("`%s`=?", columnStr))
			}
		} else {
			embeddedStruct := rv.Field(i).Elem()
			embeddedStructLen := embeddedStruct.NumField()
			prefix := getEmbeddedPrefix(tagVal)
			for j := 0; j < embeddedStruct.NumField(); j++ {
				tagVal := reflect.TypeOf(embeddedStruct.Interface()).Field(j).Tag.Get(gormTag)
				columnStr := getColumn(tagVal)
				if i < filedLen-1 || j < embeddedStructLen-1 {
					buf.WriteString(fmt.Sprintf("`%s%s`=?,", prefix, columnStr))
				} else {
					buf.WriteString(fmt.Sprintf("`%s%s`=?", prefix, columnStr))
				}
			}
		}
	}
	buf.WriteString(fmt.Sprintf(" WHERE `%s` = ?", primaryKey))
	return buf.String()
}

func isPrimaryKey(str string) bool {
	return strings.Contains(str, "primary_key")
}

func getColumn(str string) string {
	// eg. str=column:id;type:varchar(256);primary_key or str=primary_key;column:id;type:varchar(256)
	for _, one := range strings.Split(str, ";") {
		if strings.Contains(one, "column") {
			return strings.Split(one, ":")[1]
		}
	}
	panic("not found column from " + str)
}

func genInsertSQL(obj interface{}) string {
	rt := reflect.TypeOf(obj)
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
		rt = rt.Elem()
	}
	tabler, ok := obj.(schema.Tabler)
	if !ok {
		panic("not implement schema.Tabler")
	}
	// eg. INSERT INTO `table_name` (`id`,`name`) VALUES (?,?)"
	filedLen := rv.NumField()
	buf := &bytes.Buffer{}
	vals := ""
	buf.WriteString(fmt.Sprintf("INSERT INTO `%s` (", tabler.TableName()))
	for i := 0; i < filedLen; i++ {
		tagVal := rt.Field(i).Tag.Get(gormTag)
		isEmbedded := isEmbedded(tagVal)
		if !isEmbedded {
			columnStr := getColumn(tagVal)
			if i < filedLen-1 {
				buf.WriteString(fmt.Sprintf("`%s`,", columnStr))
				vals += "?,"
			} else {
				buf.WriteString(fmt.Sprintf("`%s`", columnStr))
				vals += "?"
			}
		} else {
			embeddedStruct := rv.Field(i).Elem()
			embeddedStructLen := embeddedStruct.NumField()
			prefix := getEmbeddedPrefix(tagVal)
			for j := 0; j < embeddedStruct.NumField(); j++ {
				tagVal := reflect.TypeOf(embeddedStruct.Interface()).Field(j).Tag.Get(gormTag)
				columnStr := getColumn(tagVal)
				if i < filedLen-1 || j < embeddedStructLen-1 {
					buf.WriteString(fmt.Sprintf("`%s%s`,", prefix, columnStr))
					vals += "?,"
				} else {
					buf.WriteString(fmt.Sprintf("`%s%s`", prefix, columnStr))
					vals += "?"
				}
			}
		}
	}
	buf.WriteString(") VALUES (")
	buf.WriteString(vals)
	buf.WriteString(")")
	return buf.String()
}

func isEmbedded(str string) bool {
	return strings.Contains(str, "embedded") && strings.Contains(str, "embeddedPrefix")
}

func getEmbeddedPrefix(str string) string {
	// eg. str=embedded;embeddedPrefix:receipt_
	for _, one := range strings.Split(str, ";") {
		if strings.Contains(one, "embeddedPrefix") {
			return strings.Split(one, ":")[1]
		}
	}
	panic("not found embedded prefix from " + str)
}
