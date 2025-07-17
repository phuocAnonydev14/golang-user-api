package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(user CreateUserRequest) (*UserResponse, error) {
	id := uuid.NewString()
	
	query := `
		INSERT INTO users (id, username, email, age) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, username, email, age, created_at
	`
	
	var response UserResponse
	var createdAt interface{} // Ignore created_at for now
	
	err := r.db.QueryRow(context.Background(), query, id, user.Username, user.Email, user.Age).
		Scan(&response.ID, &response.Username, &response.Email, &response.Age, &createdAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return &response, nil
}

func (r *Repository) GetByID(id string) (*UserResponse, error) {
	query := `SELECT id, username, email, age FROM users WHERE id = $1`
	
	var user UserResponse
	err := r.db.QueryRow(context.Background(), query, id).
		Scan(&user.ID, &user.Username, &user.Email, &user.Age)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	
	return &user, nil
}

func (r *Repository) GetByEmail(email string) (*UserResponse, error) {
	query := `SELECT id, username, email, age FROM users WHERE email = $1`
	
	var user UserResponse
	err := r.db.QueryRow(context.Background(), query, email).
		Scan(&user.ID, &user.Username, &user.Email, &user.Age)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return &user, nil
}

func (r *Repository) GetAll() ([]UserResponse, error) {
	query := `SELECT id, username, email, age FROM users ORDER BY created_at DESC`
	
	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()
	
	var users []UserResponse
	for rows.Next() {
		var user UserResponse
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Age); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	
	return users, nil
}

func (r *Repository) Update(id string, user CreateUserRequest) (*UserResponse, error) {
	query := `
		UPDATE users 
		SET username = $2, email = $3, age = $4, updated_at = NOW()
		WHERE id = $1 
		RETURNING id, username, email, age
	`
	
	var response UserResponse
	err := r.db.QueryRow(context.Background(), query, id, user.Username, user.Email, user.Age).
		Scan(&response.ID, &response.Username, &response.Email, &response.Age)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	
	return &response, nil
}

func (r *Repository) Delete(id string) error {
	query := `DELETE FROM users WHERE id = $1`
	
	result, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}