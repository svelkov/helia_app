package repository

import (
	"context"
	"fmt"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"reflect"
	"strings"
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

func (r *BaseRepository[T]) GetByID(ctx context.Context, idField string, idValue interface{}) (*T, error) {
	var entity T
	query := r.CreateGetByIDStatement(idField, idValue)
	// Execute the query
	err := r.DB.GetContext(ctx, &entity, query, idValue)
	if err != nil {
		return nil, fmt.Errorf("error fetching entity by ID: %w", err)
	}
	return &entity, nil
}

func (r *BaseRepository[T]) GetAll(ctx context.Context, pageSize int, offset int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]T, error) {
	var entity []T

	qb := r.CreateGetAllStatement(ctx, tableFields, idField, searchText)
	if qb == nil {
		return nil, fmt.Errorf("error creating query builder")
	}
	if searchText != "" {	
		// Add search conditions to the query builder
		qb.AddSearchConditions(tableFields, searchText)
	}
	qb.SetLimit(pageSize)
	qb.SetOffset(offset) 
	if sortBy != "" {
		qb.AddOrderBy(sortBy)
	}	
	if sortOrder != "" {
		qb.AddSortOrder(sortOrder)
	}		
	
	query, args := qb.Build()
	// Execute the query
	err := r.DB.SelectContext(ctx, &entity, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error fetching entity records, error: %w", err)
	}
	return &entity, nil
}

func (r *BaseRepository[T]) GetAllCustom(ctx context.Context, queryText, whereText string, args []interface{}, limitOffset, sortBy string) (*[]T, error) {
	var entity []T

	queryText = fmt.Sprintf("%s %s %s %s", queryText, whereText, sortBy, limitOffset)
	// Execute the query
	//fmt.Println(queryText)
	err := r.DB.SelectContext(ctx, &entity, queryText, args...)
	if err != nil {
		return nil, fmt.Errorf("error fetching entity records, error: %w", err)
	}
	return &entity, nil
}

func (r *BaseRepository[T]) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	countRec := 0
	query, args := r.CreateGetCountRecordsStatement(ctx, tableFields, searchText)

	// Execute the query
	err := r.DB.GetContext(ctx, &countRec, query, args...)
	if err != nil {
		return 0, fmt.Errorf("error fetching count of recordsD: %w", err)
	}

	return countRec, nil
}

func (r *BaseRepository[T]) GetTotalRecordsCustom(ctx context.Context, queryText, whereText string, args []interface{}, limitOffset, sortBy string) (int, error) {

	countRec := 0

	queryText = fmt.Sprintf("%s %s %s %s", queryText, whereText, limitOffset, sortBy)
	// Execute the query
	err := r.DB.GetContext(ctx, &countRec, queryText, args...)
	if err != nil {
		return 0, fmt.Errorf("error fetching entity records, error: %w", err)
	}
	return countRec, nil
}

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T, idField string, tableFields []domain.Fields) (int64, error) {
	query, values := r.CreateInsertStatement(ctx, entity, tableFields, idField)

	lastInsertedID, err := doTransaction(ctx, r.DB, ActionTypeInsert, query, values...)
	return lastInsertedID, err
}

func (r *BaseRepository[T]) Update(ctx context.Context, entity *T, idField string, idValue interface{}, tableFields []domain.Fields) error {
	query, values := r.CreateUpdateStatement(ctx, entity, idField, idValue, tableFields)
	_, err := doTransaction(ctx, r.DB, ActionTypeUpdate, query, values...)
	return err
}

func (r *BaseRepository[T]) Delete(ctx context.Context, idField string, id int64) error {
	query := fmt.Sprintf(`DELETE FROM  %s WHERE %s = $1`, r.TableName, idField)
	_, err := doTransaction(ctx, r.DB, ActionTypeDelete, query, id)
	return err
}

// ============ TRANSACTION-AWARE METHODS ============

// BeginTx starts a new database transaction
func (r *BaseRepository[T]) BeginTx() (db.Transaction, error) {
	return r.DB.Beginx()
}

// CreateWithTx executes an INSERT within an existing transaction
func (r *BaseRepository[T]) CreateWithTx(ctx context.Context, tx db.Transaction, entity *T, idField string, tableFields []domain.Fields) (int64, error) {
	query, values := r.CreateInsertStatement(ctx, entity, tableFields, idField)
	lastInsertedID := int64(0)
	err := tx.QueryRowContext(ctx, query, values...).Scan(&lastInsertedID)
	if err != nil {
		return 0, fmt.Errorf("query execution failed: %v", err)
	}
	return lastInsertedID, nil
}

// UpdateWithTx executes an UPDATE within an existing transaction
func (r *BaseRepository[T]) UpdateWithTx(ctx context.Context, tx db.Transaction, entity *T, idField string, idValue interface{}, tableFields []domain.Fields) error {
	query, values := r.CreateUpdateStatement(ctx, entity, idField, idValue, tableFields)
	_, err := tx.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("query execution failed: %v", err)
	}
	return nil
}

// DeleteWithTx executes a DELETE within an existing transaction
func (r *BaseRepository[T]) DeleteWithTx(ctx context.Context, tx db.Transaction, idField string, id int64) error {
	query := fmt.Sprintf(`DELETE FROM  %s WHERE %s = $1`, r.TableName, idField)
	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("query execution failed: %v", err)
	}
	return nil
}

// doTransaction executes a query within a transaction
func doTransaction(ctx context.Context, database db.Database, actionType, query string, values ...interface{}) (int64, error) {
	// Start a transaction
	lastInsertedID := int64(0)
	tx, err := database.Beginx()
	if err != nil {
		return 0, fmt.Errorf("could not begin transaction: %v", err)
	}
	defer tx.Rollback() // Ensures rollback if commit isn't reached
	// Execute the query with provided values
	if actionType == ActionTypeDelete || actionType == ActionTypeUpdate {
		_, err := tx.ExecContext(ctx, query, values...)
		if err != nil {
			return 0, fmt.Errorf("query execution failed: %v", err)
		}
		err = tx.Commit() // Commit if successful
		return lastInsertedID, err
	}
	err = tx.QueryRowContext(ctx, query, values...).Scan(&lastInsertedID)
	if err != nil {
		return 0, fmt.Errorf("query execution failed: %v", err)
	}
	// Commit if successful
	err = tx.Commit()
	return lastInsertedID, err
}
