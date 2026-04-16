package service

import (
	"context"
	"database/sql"
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/validation"
	"log"
	"reflect"
	"strings"
	"time"
)

// DateTimeFormat specifies how to format time.Time fields
type DateTimeFormat string

const (
	DateOnly    DateTimeFormat = "2006.01.02"          // e.g., "2024.12.17"
	TimeOnly    DateTimeFormat = "15:04:05"            // e.g., "14:30:45"
	DateAndTime DateTimeFormat = "2006.01.02 15:04:05" // e.g., "2024.12.17 14:30:45"
)

// Generic service interface
type Service[T any] interface {
	Create(ctx context.Context, entity *T, idField string, tableFields []domain.Fields) ([]domain.FieldError, int64, error)
	GetByID(ctx context.Context, idField string, idValue int64) (*T, error)
	GetAll(ctx context.Context, page int, offset int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]T, error)
	GetAllCustom(ctx context.Context, queryText, whereText string, args []interface{}, limitOffset, orderBy string) (*[]T, error)
	GetTotalRecordsCustom(ctx context.Context, queryText, whereText string, args []interface{}, limitOffset, orderBy string) (int, error)
	GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error)
	Update(ctx context.Context, entity *T, idField string, idValue interface{}, tableFields []domain.Fields) ([]domain.FieldError, error)
	Delete(ctx context.Context, idField string, id int64) error
	MapEntityToValues(entity *T, tableFields []domain.Fields) []domain.Fields
	GetFieldCache() map[string]reflect.StructField
}

type BaseService[T any] struct {
	Repo       repository.BaseRepository[T]
	Validator  validation.Validator[T]
	fieldCache map[string]reflect.StructField
}

// NewBaseService creates a new instance of BaseService.
func NewBaseService[T any](repository repository.BaseRepository[T], validator validation.Validator[T]) *BaseService[T] {
	r := &BaseService[T]{
		Repo:       repository,
		Validator:  validator,
		fieldCache: make(map[string]reflect.StructField),
	}
	// Initialize cache using generic type
	var entity T
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		r.fieldCache[strings.ToLower(field.Name)] = field
	}

	return r
}

func (s *BaseService[T]) GetFieldCache() map[string]reflect.StructField {
	return s.fieldCache
}
func (s *BaseService[T]) SetFieldCache(fieldCache map[string]reflect.StructField) {
	s.fieldCache = fieldCache
}

func (s *BaseService[T]) GetByID(ctx context.Context, idField string, idValue int64) (*T, error) {
	return s.Repo.GetByID(ctx, idField, idValue)
}

func (s *BaseService[T]) GetAll(ctx context.Context, page int, offset int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]T, error) {
	return s.Repo.GetAll(ctx, page, offset, tableFields, idField, searchText, sortBy, sortOrder)
}

func (s *BaseService[T]) GetAllCustom(ctx context.Context, queryText, whereText string, args []interface{}, limitOffset, orderBy string) (*[]T, error) {
	return s.Repo.GetAllCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

func (s *BaseService[T]) GetTotalRecordsCustom(ctx context.Context, queryText, whereText string, args []interface{}, limitOffset, orderBy string) (int, error) {
	return s.Repo.GetTotalRecordsCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

func (s *BaseService[T]) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	return s.Repo.GetTotalRecords(ctx, tableFields, searchText)
}
func (s *BaseService[T]) Create(ctx context.Context, entity *T, idField string, tableFields []domain.Fields) ([]domain.FieldError, int64, error) {
	fieldErrors, err := s.Validator.Validate(entity)

	if err != nil {
		return []domain.FieldError{}, 0, err
	}
	if len(fieldErrors) > 0 {
		return fieldErrors, 0, nil
	}
	lastInsertedID, err := s.Repo.Create(ctx, entity, idField, tableFields)
	log.Println("BaseService Create: lastInsertedID =", lastInsertedID, " error =", err)
	return []domain.FieldError{}, lastInsertedID, err

}

func (s *BaseService[T]) Update(ctx context.Context, entity *T, idField string, idValue interface{}, tableFields []domain.Fields) ([]domain.FieldError, error) {
	fieldErrors, err := s.Validator.Validate(entity)
	if err != nil {
		return fieldErrors, err
	}
	return fieldErrors, s.Repo.Update(ctx, entity, idField, idValue, tableFields)
}

func (s *BaseService[T]) Delete(ctx context.Context, idField string, id int64) error {
	return s.Repo.Delete(ctx, idField, id)
}

// formatTime formats a time.Time with the specified format
func (s *BaseService[T]) formatTime(t time.Time, format DateTimeFormat) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(string(format))
}

func (s *BaseService[T]) formatFieldValue(fieldValue reflect.Value) string {
	typeStr := fieldValue.Type().String()

	// Handle time.Time - default to date only
	if typeStr == "time.Time" {
		t := fieldValue.Interface().(time.Time)
		return s.formatTime(t, DateOnly)
	}

	// Handle sql.NullTime
	if typeStr == "database/sql.NullTime" || typeStr == "sql.NullTime" {
		nt := fieldValue.Interface().(sql.NullTime)
		if nt.Valid {
			return nt.Time.Format(common.DateLayout)
		}
		return ""
	}

	// Handle sql.NullString
	if typeStr == "database/sql.NullString" || typeStr == "sql.NullString" {
		ns := fieldValue.Interface().(sql.NullString)
		if ns.Valid {
			return ns.String
		}
		return ""
	}

	// Handle sql.NullInt64
	if typeStr == "database/sql.NullInt64" || typeStr == "sql.NullInt64" {
		ni := fieldValue.Interface().(sql.NullInt64)
		if ni.Valid {
			return fmt.Sprintf("%d", ni.Int64)
		}
		return ""
	}

	// Default: convert to string
	return fmt.Sprintf("%v", fieldValue.Interface())
}

func (s *BaseService[T]) MapEntityToValues(entity *T, tableFields []domain.Fields) []domain.Fields {
	// Get the reflect value and type of the entity
	entityValue := reflect.ValueOf(entity)
	entityType := reflect.TypeOf(entity)

	// Ensure we are working with the actual struct, not a pointer
	if entityValue.Kind() == reflect.Ptr {
		entityValue = entityValue.Elem()
		entityType = entityType.Elem()
	}

	// Iterate over struct fields
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		dbTag, dbOk := field.Tag.Lookup("db")
		addUpdateTag, addUpdateOk := field.Tag.Lookup("addupdate")

		// Include if: (db exists AND addupdate="true") OR (db exists AND addupdate is not present)
		if dbOk && (addUpdateTag == "true" || !addUpdateOk) {
			if dbTag == "" || dbTag == "-" {
				continue // Skip fields without `db` tags or explicitly ignored
			}

			// Format the field value
			displayValue := s.formatFieldValue(entityValue.Field(i))

			// Check if the column exists in tableFields
			for j := range tableFields {
				if strings.EqualFold(tableFields[j].Name, dbTag) {
					tableFields[j].Value = displayValue
					break
				}
			}
		}
	}
	return tableFields
}
