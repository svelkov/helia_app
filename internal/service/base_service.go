package service

import (
	"fmt"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/validation"
	"reflect"
	"strings"
)

// Generic service interface
type Service[T any] interface {
	Create(entity *T, idField string, tableFields []domain.Fields) ([]domain.FieldError, int64, error)
	GetByID(idField string, idValue int64) (*T, error)
	GetAll(page int, offset int, tableFields []domain.Fields, idField string, searchParams ...string) (*[]T, error)
	GetAllCustom(queryText, whereText string, args []interface{}, limitOffset, orderBy string) (*[]T, error)
	GetTotalRecordsCustom(queryText, whereText string, args []interface{}, limitOffset, orderBy string) (int, error)
	GetTotalRecords(tableFields []domain.Fields, searchParams ...string) (int, error)
	Update(entity *T, idField string, idValue interface{}, tableFields []domain.Fields) ([]domain.FieldError, error)
	Delete(idField string, id int64) error
	MapEntityToValues(entity *T, tableFields []domain.Fields) []domain.Fields
	GetFieldCache() map[string]reflect.StructField
}

type BaseService[T any] struct {
	Repo       repository.BaseRepository[T]
	Validator  validation.RuleBasedValidator[T]
	fieldCache map[string]reflect.StructField
}

// NewBaseService creates a new instance of BaseService.
func NewBaseService[T any](repository repository.BaseRepository[T], validator validation.RuleBasedValidator[T]) *BaseService[T] {
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

func (s *BaseService[T]) Create(entity *T, idField string, tableFields []domain.Fields) ([]domain.FieldError, int64, error) {
	fieldErrors, err := s.Validator.Validate(entity)

	if err != nil {
		return []domain.FieldError{}, 0, err
	}
	if len(fieldErrors) > 0 {
		return fieldErrors, 0, nil
	}
	lastInsertedID, err := s.Repo.Create(entity, idField, tableFields)
	return []domain.FieldError{}, lastInsertedID, err
}

func (s *BaseService[T]) GetFieldCache() map[string]reflect.StructField {
	return s.fieldCache
}

func (s *BaseService[T]) GetByID(idField string, idValue int64) (*T, error) {
	return s.Repo.GetByID(idField, idValue)
}

func (s *BaseService[T]) GetAll(page int, offset int, tableFields []domain.Fields, idField string, searchParams ...string) (*[]T, error) {
	return s.Repo.GetAll(page, offset, tableFields, idField, searchParams...)
}

func (s *BaseService[T]) GetAllCustom(queryText, whereText string, args []interface{}, limitOffset, orderBy string) (*[]T, error) {
	return s.Repo.GetAllCustom(queryText, whereText, args, limitOffset, orderBy)
}

func (s *BaseService[T]) GetTotalRecordsCustom(queryText, whereText string, args []interface{}, limitOffset, orderBy string) (int, error) {
	return s.Repo.GetTotalRecordsCustom(queryText, whereText, args, limitOffset, orderBy)
}

func (s *BaseService[T]) GetTotalRecords(tableFields []domain.Fields, searchParams ...string) (int, error) {
	return s.Repo.GetTotalRecords(tableFields, searchParams...)
}

func (s *BaseService[T]) Update(entity *T, idField string, idValue interface{}, tableFields []domain.Fields) ([]domain.FieldError, error) {
	fieldErrors, err := s.Validator.Validate(entity)
	if err != nil {
		return fieldErrors, err
	}
	return fieldErrors, s.Repo.Update(entity, idField, idValue, tableFields)
}

func (s *BaseService[T]) Delete(idField string, id int64) error {
	return s.Repo.Delete(idField, id)
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
	for i := range entityType.NumField() {
		field := entityType.Field(i)
		column := field.Tag.Get("db") // Get the `db` tag value
		if column == "" || column == "-" {
			continue // Skip fields without `db` tags or explicitly ignored
		}

		// Check if the column exists in tableFields
		for j, field := range tableFields {
			if strings.EqualFold(field.Name, column) {
				tableFields[j].Value = fmt.Sprintf("%v", entityValue.Field(i).Interface())
				break
			}
		}
	}
	return tableFields
}
