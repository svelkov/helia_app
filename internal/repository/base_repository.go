package repository

import (
	"fmt"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// // Generic repository interface
// type Repository[T any] interface {
// 	Create(entity *T) error
// 	GetByID(idField string, id int) (*T, error)
// 	GetAll(page int, offset int) (*[]T, error)
// 	Update(entity *T) error
// 	Delete(entity *T, id int) error
// 	GetTotalRecords() (int, error)
// }

const (
	// ActionTypeInsert is used for insert operations
	ActionTypeInsert = "insert"
	// ActionTypeUpdate is used for update operations
	ActionTypeUpdate = "update"
	// ActionTypeDelete is used for delete operations
	ActionTypeDelete = "delete"
)

// Base implementation
type BaseRepository[T any] struct {
	DB         db.Database
	TableName  string
	fieldCache map[string]reflect.StructField
}

// NewBaseRepository creates a new instance of BaseRepository.
func NewBaseRepository[T any](database db.Database, tableName string) *BaseRepository[T] {
	br := &BaseRepository[T]{
		DB:         database,
		TableName:  tableName,
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
		br.fieldCache[strings.ToLower(field.Name)] = field
	}

	return br
}
func (r *BaseRepository[T]) GetFieldCache() map[string]reflect.StructField {
	return r.fieldCache
}

func (r *BaseRepository[T]) GetByID(idField string, idValue interface{}) (*T, error) {
	var entity T
	query := r.CreateGetByIDStatement(idField, idValue)
	// Execute the query
	err := r.DB.Get(&entity, query, idValue)
	if err != nil {
		return nil, fmt.Errorf("error fetching entity by ID: %w", err)
	}
	return &entity, nil
}

func (r *BaseRepository[T]) GetAll(c *gin.Context, pageSize int, offset int, tableFields []domain.Fields, idField string, searchParams ...string) (*[]T, error) {
	var entity []T

	query, args := r.CreateGetAllStatement(c, tableFields, idField, searchParams...)
	param := len(args) + 1

	endPaging := ""
	if pageSize > 0 {
		args = append(args, pageSize, offset)
		endPaging = fmt.Sprintf(` LIMIT $%d OFFSET $%d`, param, param+1)
	}
	query = fmt.Sprintf("%s %s", query, endPaging)
	// Execute the query
	err := r.DB.Select(&entity, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error fetching entity records, error: %w", err)
	}
	return &entity, nil
}

func (r *BaseRepository[T]) GetAllCustom(c *gin.Context, queryText, whereText string, args []interface{}, limitOffset, sortBy string) (*[]T, error) {
	var entity []T

	queryText = fmt.Sprintf("%s %s %s %s", queryText, whereText, sortBy, limitOffset)
	// Execute the query
	//fmt.Println(queryText)
	err := r.DB.Select(&entity, queryText, args...)
	if err != nil {
		return nil, fmt.Errorf("error fetching entity records, error: %w", err)
	}
	return &entity, nil
}

func (r *BaseRepository[T]) GetTotalRecords(c *gin.Context, tableFields []domain.Fields, searchParams ...string) (int, error) {
	countRec := 0
	query, args := r.CreateGetCountRecordsStatement(c, tableFields, searchParams...)

	// Execute the query
	err := r.DB.Get(&countRec, query, args...)
	if err != nil {
		return 0, fmt.Errorf("error fetching count of recordsD: %w", err)
	}

	return countRec, nil
}

func (r *BaseRepository[T]) GetTotalRecordsCustom(c *gin.Context, queryText, whereText string, args []interface{}, limitOffset, sortBy string) (int, error) {

	countRec := 0

	queryText = fmt.Sprintf("%s %s %s %s", queryText, whereText, limitOffset, sortBy)
	// Execute the query
	err := r.DB.Get(&countRec, queryText, args...)
	if err != nil {
		return 0, fmt.Errorf("error fetching entity records, error: %w", err)
	}
	return countRec, nil
}

func (r *BaseRepository[T]) Create(c *gin.Context, entity *T, idField string, tableFields []domain.Fields) (int64, error) {
	query, values := r.CreateInsertStatement(c, entity, tableFields, idField)
	lastInsertedID, err := doTransaction(r.DB, ActionTypeInsert, query, values...)
	return lastInsertedID, err
}

func (r *BaseRepository[T]) Update(c *gin.Context, entity *T, idField string, idValue interface{}, tableFields []domain.Fields) error {
	query, values := r.CreateUpdateStatement(c, entity, idField, idValue, tableFields)
	_, err := doTransaction(r.DB, ActionTypeUpdate, query, values...)
	return err
}

func (r *BaseRepository[T]) Delete(idField string, id int64) error {
	query := fmt.Sprintf(`DELETE FROM  %s WHERE %s = $1`, r.TableName, idField)
	_, err := doTransaction(r.DB, ActionTypeDelete, query, id)
	return err
}

// doTransaction executes a query within a transaction
func doTransaction(database db.Database, actionType, query string, values ...interface{}) (int64, error) {
	// Start a transaction
	lastInsertedID := int64(0)
	tx, err := database.Beginx()
	if err != nil {
		return 0, fmt.Errorf("could not begin transaction: %v", err)
	}
	defer tx.Rollback() // Ensures rollback if commit isn't reached
	// Execute the query with provided values
	if actionType == ActionTypeDelete || actionType == ActionTypeUpdate {
		_, err := tx.Exec(query, values...)
		if err != nil {
			return 0, fmt.Errorf("query execution failed: %v", err)
		}
		err = tx.Commit() // Commit if successful
		return lastInsertedID, err
	}
	err = tx.QueryRow(query, values...).Scan(&lastInsertedID)
	if err != nil {
		return 0, fmt.Errorf("query execution failed: %v", err)
	}
	// Commit if successful
	err = tx.Commit()
	return lastInsertedID, err
}
