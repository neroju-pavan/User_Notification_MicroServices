package repositories

import (
	"context"
	"fmt"
	"time"

	"test123/errors"
	"test123/logger"
	"test123/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	DB *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{DB: db}
}

//
// ─────────────────────────────────────────── CREATE USER ─────
//

func (r *UserRepo) CreateUser(ctx context.Context, user models.User) error {
	logger.Info("UserRepo.CreateUser", "creating user", map[string]interface{}{
		"email":    user.Email,
		"username": user.Username,
	})

	existsByUsername, _ := r.GetUserByEmailOrUsername(ctx, user.Username)
	existsByEmail, _ := r.GetUserByEmailOrUsername(ctx, user.Email)

	if existsByEmail != nil || existsByUsername != nil {
		logger.Warn("UserRepo.CreateUser", "user exists", map[string]interface{}{
			"email":    user.Email,
			"username": user.Username,
		})
		return errors.ErrUserExists
	}

	loc, _ := time.LoadLocation("Asia/Kolkata")
	user.CreatedAt = time.Now().In(loc)

	query := `
		INSERT INTO users (name, email, username, password, mobile_number, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.DB.Exec(ctx, query,
		user.Name,
		user.Email,
		user.Username,
		user.Password,
		user.MobileNumber,
		user.CreatedAt,
	)

	if err != nil {
		logger.Error("UserRepo.CreateUser", "db insert failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
	}

	logger.Info("UserRepo.CreateUser", "user created successfully", map[string]interface{}{
		"email": user.Email,
	})

	return nil
}

//
// ─────────────────────────────────────────── GET ALL USERS ─────
//

func (r *UserRepo) GetAllUsers(ctx context.Context) ([]models.User, error) {
	logger.Info("UserRepo.GetAllUsers", "fetching users")

	query := `
        SELECT id, name, email, username, mobile_number, created_at
        FROM users;
    `

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		logger.Error("UserRepo.GetAllUsers", "db query failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Username, &u.MobileNumber, &u.CreatedAt); err != nil {
			logger.Error("UserRepo.GetAllUsers", "scan failed", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
		}
		users = append(users, u)
	}

	if len(users) == 0 {
		logger.Warn("UserRepo.GetAllUsers", "no users found", nil)
		return nil, errors.ErrUserNotFound
	}

	logger.Info("UserRepo.GetAllUsers", "users fetched successfully", map[string]interface{}{
		"count": len(users),
	})

	return users, nil
}

//
// ─────────────────────────────────────────── GET USER BY ID ─────
//

func (r *UserRepo) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	logger.Info("UserRepo.GetUserByID", "fetching user by id", map[string]interface{}{
		"id": id,
	})

	query := `
		SELECT id, name, email, username, mobile_number, created_at
		FROM users WHERE id = $1
	`

	var u models.User

	err := r.DB.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.Username, &u.MobileNumber, &u.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Warn("UserRepo.GetUserByID", "user not found", map[string]interface{}{"id": id})
			return nil, errors.ErrUserNotFound
		}
		logger.Error("UserRepo.GetUserByID", "db error", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
	}

	return &u, nil
}

//
// ─────────────────────────── GET USER BY EMAIL OR USERNAME ─────
//

func (r *UserRepo) GetUserByEmailOrUsername(ctx context.Context, key string) (*models.User, error) {
	logger.Debug("UserRepo.GetUserByEmailOrUsername", "searching user", map[string]interface{}{
		"key": key,
	})

	query := `
		SELECT id, name, email, username, password, mobile_number
		FROM users WHERE email=$1 OR username=$1
	`

	var u models.User

	err := r.DB.QueryRow(ctx, query, key).Scan(
		&u.ID, &u.Name, &u.Email, &u.Username, &u.Password, &u.MobileNumber,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		logger.Error("UserRepo.GetUserByEmailOrUsername", "db error", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
	}

	return &u, nil
}

//
// ─────────────────────────────────────────── UPDATE USER ─────
//

func (r *UserRepo) UpdateUser(ctx context.Context, user models.User) error {
	logger.Info("UserRepo.UpdateUser", "updating user", map[string]interface{}{
		"id": user.ID,
	})

	query := `
		UPDATE users 
		SET name=$1, email=$2, username=$3, mobile_number=$4
		WHERE id=$5
	`

	val, err := r.DB.Exec(ctx, query,
		user.Name, user.Email, user.Username, user.MobileNumber, user.ID,
	)
	if err != nil {
		logger.Error("UserRepo.UpdateUser", "update failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
	}

	if val.RowsAffected() == 0 {
		logger.Warn("UserRepo.UpdateUser", "user not found", map[string]interface{}{
			"id": user.ID,
		})
		return errors.ErrUserNotFound
	}

	return nil
}

//
// ─────────────────────────────────────────── UPDATE PASSWORD ─────
//

func (r *UserRepo) UpdatePassword(ctx context.Context, username string, password string) error {
	logger.Info("UserRepo.UpdatePassword", "updating password", map[string]interface{}{
		"username": username,
	})

	query := `UPDATE users SET password=$1 WHERE username=$2`

	val, err := r.DB.Exec(ctx, query, password, username)
	if err != nil {
		logger.Error("UserRepo.UpdatePassword", "update failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
	}

	if val.RowsAffected() == 0 {
		logger.Warn("UserRepo.UpdatePassword", "user not found", map[string]interface{}{
			"username": username,
		})
		return errors.ErrUserNotFound
	}

	return nil
}

//
// ─────────────────────────────────────────── DELETE USER ─────
//

func (r *UserRepo) DeleteUser(ctx context.Context, id int) error {
	logger.Warn("UserRepo.DeleteUser", "deleting user", map[string]interface{}{
		"id": id,
	})

	query := `DELETE FROM users WHERE id = $1`

	val, err := r.DB.Exec(ctx, query, id)
	if err != nil {
		logger.Error("UserRepo.DeleteUser", "db error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
	}

	if val.RowsAffected() == 0 {
		logger.Warn("UserRepo.DeleteUser", "user not found", map[string]interface{}{
			"id": id,
		})
		return errors.ErrUserNotFound
	}

	return nil
}

//
// ─────────────────────────────────────────── GET USER BY EMAIL ─────
//

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	logger.Info("UserRepo.GetUserByEmail", "fetching user", map[string]interface{}{
		"email": email,
	})

	query := `
		SELECT id, name, email, username, password, mobile_number
		FROM users WHERE email = $1
	`

	var u models.User

	err := r.DB.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.Username, &u.Password, &u.MobileNumber,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Warn("UserRepo.GetUserByEmail", "user not found", map[string]interface{}{
				"email": email,
			})
			return nil, errors.ErrUserNotFound
		}
		logger.Error("UserRepo.GetUserByEmail", "db error", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
	}

	return &u, nil
}

//
// ─────────────────────────────────────────── GET USER BY USER (ID) ─────
//

func (r *UserRepo) GetUserByUser(ctx context.Context, id int) (*models.User, error) {
	return r.GetUserByID(ctx, id)
}

//
// ─────────────────────────────────────────── GET USER BY USERNAME ─────
//

func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	logger.Info("UserRepo.GetUserByUsername", "fetching user", map[string]interface{}{
		"username": username,
	})

	query := `
		SELECT id, name, email, username, password, mobile_number
		FROM users WHERE username = $1
	`

	var u models.User

	err := r.DB.QueryRow(ctx, query, username).Scan(
		&u.ID, &u.Name, &u.Email, &u.Username, &u.Password, &u.MobileNumber,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Warn("UserRepo.GetUserByUsername", "user not found", map[string]interface{}{
				"username": username,
			})
			return nil, errors.ErrUserNotFound
		}

		logger.Error("UserRepo.GetUserByUsername", "db error", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
	}

	return &u, nil
}

//
// ─────────────────────────────────────────── FILTER + CURSOR PAGINATION ─────
//

func (r *UserRepo) GetUsersWithFiltersCursor(
	ctx context.Context,
	limit int,
	cursor *time.Time,
	usernameSearch string,
	fromDate, toDate *time.Time,
) ([]models.User, *time.Time, error) {

	logger.Debug("UserRepo.GetUsersWithFiltersCursor", "executing filter query", map[string]interface{}{
		"limit":     limit,
		"cursor":    cursor,
		"search":    usernameSearch,
		"from_date": fromDate,
		"to_date":   toDate,
	})

	if limit <= 0 {
		limit = 5
	}

	query := `
	WITH filtered_users AS (
		SELECT id, name, email, username, mobile_number, created_at
		FROM users
		WHERE ($1::timestamptz IS NULL OR created_at < $1)
		  AND ($2::timestamptz IS NULL OR created_at >= $2)
		  AND ($3::timestamptz IS NULL OR created_at <= $3)
		  AND ($4::text IS NULL OR username ILIKE $4)
	)
	SELECT id, name, email, username, mobile_number, created_at
	FROM filtered_users
	ORDER BY created_at DESC, id DESC
	LIMIT $5;
    `

	var searchPattern interface{}
	if usernameSearch != "" {
		searchPattern = usernameSearch + "%"
	} else {
		searchPattern = nil
	}

	rows, err := r.DB.Query(ctx, query,
		cursor, fromDate, toDate, searchPattern, limit,
	)
	if err != nil {
		logger.Error("UserRepo.GetUsersWithFiltersCursor", "db query failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, nil, fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
	}
	defer rows.Close()

	var users []models.User
	var nextCursor *time.Time

	for rows.Next() {
		var user models.User

		if err := rows.Scan(
			&user.ID, &user.Name, &user.Email, &user.Username,
			&user.MobileNumber, &user.CreatedAt,
		); err != nil {

			logger.Error("UserRepo.GetUsersWithFiltersCursor", "scan error", map[string]interface{}{
				"error": err.Error(),
			})

			return nil, nil, fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
		}

		users = append(users, user)
		nextCursor = &user.CreatedAt
	}

	if len(users) == 0 {
		logger.Warn("UserRepo.GetUsersWithFiltersCursor", "no users found with filters", nil)
		return nil, nil, errors.ErrUserNotFound
	}

	return users, nextCursor, nil
}
