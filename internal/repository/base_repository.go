package repository

import (
	"fmt"
	"helia/internal/domain"

	"github.com/jmoiron/sqlx"
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
	DB        *sqlx.DB
	TableName string
}

// NewBaseRepository creates a new instance of BaseRepository.
func NewBaseRepository[T any](db *sqlx.DB, tableName string) *BaseRepository[T] {
	return &BaseRepository[T]{
		DB:        db,
		TableName: tableName,
	}
}

func (r *BaseRepository[T]) GetByID(idField string, idValue interface{}) (*T, error) {
	var entity T
	query := r.CreateGetByID(idField, idValue)
	// Execute the query
	err := r.DB.Get(&entity, query, idValue)
	if err != nil {
		return nil, fmt.Errorf("error fetching entity by ID: %w", err)
	}
	return &entity, nil
}

func (r *BaseRepository[T]) GetAll(pageSize int, offset int, tableFields []domain.Fields, idField string, searchParams ...string) (*[]T, error) {
	var entity []T
	args := []interface{}{}
	query := r.CreateGetAll(&args, tableFields, idField, searchParams...)
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

func (r *BaseRepository[T]) GetAllCustom(queryText, whereText string, args []interface{}, limitOffset, sortBy string) (*[]T, error) {
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

func (r *BaseRepository[T]) GetTotalRecords(tableFields []domain.Fields, searchParams ...string) (int, error) {
	args := []interface{}{}
	countRec := 0
	query := r.CreateGetCountRecords(tableFields, &args, searchParams...)

	// Execute the query
	err := r.DB.Get(&countRec, query, args...)
	if err != nil {
		return 0, fmt.Errorf("error fetching count of recordsD: %w", err)
	}

	return countRec, nil
}

func (r *BaseRepository[T]) GetTotalRecordsCustom(queryText, whereText string, args []interface{}, limitOffset, sortBy string) (int, error) {

	countRec := 0

	queryText = fmt.Sprintf("%s %s %s %s", queryText, whereText, limitOffset, sortBy)
	// Execute the query
	err := r.DB.Get(&countRec, queryText, args...)
	if err != nil {
		return 0, fmt.Errorf("error fetching entity records, error: %w", err)
	}
	return countRec, nil
}

func (r *BaseRepository[T]) Create(entity *T, idField string, tableFields []domain.Fields) (int64, error) {
	query, values := r.CreateInsertStatement(entity, tableFields)
	lastInsertedID, err := doTransaction(r.DB, ActionTypeInsert, idField, query, values...)
	return lastInsertedID, err
}

func (r *BaseRepository[T]) Update(entity *T, idField string, idValue interface{}, tableFields []domain.Fields) error {
	query, values := r.CreateUpdateStatement(entity, idField, idValue, tableFields)
	_, err := doTransaction(r.DB, ActionTypeUpdate, idField, query, values...)
	return err
}

func (r *BaseRepository[T]) Delete(idField string, id int) error {
	query := fmt.Sprintf(`DELETE FROM  %s WHERE %s = $1`, r.TableName, idField)
	_, err := doTransaction(r.DB, ActionTypeDelete, idField, query, id)
	return err
}

// doTransaction executes a query within a transaction
func doTransaction(db *sqlx.DB, actionType, idField, query string, values ...interface{}) (int64, error) {
	// Start a transaction
	lastInsertedID := int64(0)
	tx, err := db.Beginx()
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
	query = fmt.Sprintf("%s RETURNING %s", query, idField) // Append RETURNING clause for insert
	err = tx.QueryRow(query, values...).Scan(&lastInsertedID)
	if err != nil {
		return 0, fmt.Errorf("query execution failed: %v", err)
	}
	// Commit if successful
	err = tx.Commit()
	return lastInsertedID, err
}
