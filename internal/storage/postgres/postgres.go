package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
	"inHouseAd/internal/entity"
	"inHouseAd/internal/http-server/handlers/auth/signin"
	"inHouseAd/internal/http-server/handlers/auth/signup"
)

var ErrNotFound = errors.New("record not found")

type Storage struct {
	db *sql.DB
}

func New(host, port, user, password, dbName string) (*Storage, error) {
	const op = "storage.postgres.New"

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	storage := &Storage{db: db}

	err = goose.Up(storage.db, "db/migrations")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return storage, nil
}

func (s *Storage) Register(email string, passwordHashed []byte) (int, error) {
	const op = "storage.postgres.CreateUser"

	query := `
		INSERT INTO users (email, password_hashed) 
		VALUES ($1, $2) 
		RETURNING id;
		`

	var id int

	err := s.db.QueryRow(query, email, passwordHashed).Scan(&id)
	if err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == "23505" {
			return 0, signup.ErrEmailTaken
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) Authorizate(email string) ([]byte, int, error) {
	const op = "storage.postgres.Authorizate"

	query := `
		SELECT users.password_hashed, users.id 
		FROM users 
		WHERE email = $1 
		LIMIT 1;
		`

	var hash []byte
	var id int

	err := s.db.QueryRow(query, email).Scan(&hash, &id)
	if err == sql.ErrNoRows {
		return nil, 0, signin.ErrInvalidEmail
	} else if err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == "23505" {
			return nil, 0, signup.ErrEmailTaken
		}
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	return hash, id, nil
}

func (s *Storage) Create(name string, uid int) (int, error) {
	const op = "storage.postgres.Create"

	var id int

	query := `
		INSERT INTO category (category_name) 
		VALUES ($1) 
		RETURNING id;
		`

	err := s.db.QueryRow(query, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) EditCategory(id int, newName string) (int, error) {
	const op = "storage.postgres.EditCategory"

	query := `
		UPDATE category 
		SET category_name = $1 
		WHERE id = $2 
		RETURNING id;
		`

	err := s.db.QueryRow(query, newName, id).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) DeleteCategory(id int) error {
	const op = "storage.postgres.DeleteCategory"

	query := `
			DELETE FROM category 
       		WHERE id = $1;
			`

	_, err := s.db.Exec(query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) AddGood(goodName string, categoryId int) (int, string, error) {
	const op = "storage.postgres.AddGood"

	var (
		goodId       int
		categoryName string
	)

	query := `
			INSERT INTO good (good_name) 
			VALUES ($1) 
			RETURNING id;
		`

	tx, err := s.db.Begin()
	if err != nil {
		return 0, "", fmt.Errorf("%s: %w", op, err)
	}

	err = tx.QueryRow(query, goodName).Scan(&goodId)
	if err != nil {
		tx.Rollback()
		return 0, "", fmt.Errorf("%s: %w", op, err)
	}

	query = `
			INSERT INTO good_category (good_id, category_id) 
			VALUES ($1, $2);
		`

	_, err = tx.Exec(query, goodId, categoryId)
	if err != nil {
		tx.Rollback()
		return 0, "", fmt.Errorf("%s: %w", op, err)
	}

	query = `
			SELECT category.category_name 
			FROM category 
			WHERE id = $1
			LIMIT 1;
		`

	err = tx.QueryRow(query, categoryId).Scan(&categoryName)
	if err != nil {
		tx.Rollback()
		return 0, "", fmt.Errorf("%s: %w", op, err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, "", fmt.Errorf("%s: %w", op, err)
	}

	return goodId, categoryName, nil
}

func (s *Storage) UpdateGood(goodId, categoryIdToAdd int, goodName string) (int, []string, string, error) {
	const op = "storage.postgres.UpdateGood"

	var (
		rGoodName     string
		categoryNames []string
	)

	tx, err := s.db.Begin()
	if err != nil {
		return 0, nil, "", fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if r := recover(); r != nil || err != nil {
			tx.Rollback()
		}
	}()

	query := `SELECT good_name FROM good WHERE id = $1;`
	if err := tx.QueryRow(query, goodId).Scan(&rGoodName); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil, "", fmt.Errorf("%s: good not found", op)
		}
		return 0, nil, "", fmt.Errorf("%s: %w", op, err)
	}

	if goodName != "" {
		query = `UPDATE good SET good_name = $1 WHERE id = $2 RETURNING good_name;`
		if err := tx.QueryRow(query, goodName, goodId).Scan(&rGoodName); err != nil {
			return 0, nil, "", fmt.Errorf("%s: %w", op, err)
		}
	}

	if categoryIdToAdd != 0 {
		query = `SELECT category_name FROM category WHERE id = $1;`
		var categoryName string
		if err := tx.QueryRow(query, categoryIdToAdd).Scan(&categoryName); err != nil {
			if err == sql.ErrNoRows {
				return 0, nil, "", fmt.Errorf("%s: category not found", op)
			}
			return 0, nil, "", fmt.Errorf("%s: %w", op, err)
		}

		query = `INSERT INTO good_category (good_id, category_id) VALUES ($1, $2);`
		if _, err := tx.Exec(query, goodId, categoryIdToAdd); err != nil {
			return 0, nil, "", fmt.Errorf("%s: %w", op, err)
		}
	}

	query = `
        SELECT c.category_name
        FROM category AS c
        JOIN good_category gc ON c.id = gc.category_id
		WHERE gc.good_id = $1;
		`
	rows, err := tx.Query(query, goodId)
	if err != nil {
		return 0, nil, "", fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var categoryName string
		if err := rows.Scan(&categoryName); err != nil {
			return 0, nil, "", fmt.Errorf("%s: %w", op, err)
		}
		categoryNames = append(categoryNames, categoryName)
	}

	if err := rows.Err(); err != nil {
		return 0, nil, "", fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, nil, "", fmt.Errorf("%s: %w", op, err)
	}

	return goodId, categoryNames, rGoodName, nil
}

func (s *Storage) DeleteGood(id int) error {
	const op = "storage.postgres.DeleteGood"

	query := `
			DELETE FROM good 
       		WHERE id = $1;
			`

	_, err := s.db.Exec(query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetCategoryList() ([]entity.CategoryList, error) {
	const op = "storage.postgres.GetCategoryList"

	var response []entity.CategoryList

	query := `
        SELECT *
        FROM category;
		`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var r entity.CategoryList
		if err := rows.Scan(&r.CategoryId, &r.CategoryName); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		response = append(response, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (s *Storage) GetGoodList(categoryId int) ([]entity.GoodList, error) {
	const op = "storage.postgres.GetGoodList"

	var response []entity.GoodList

	query := `
        SELECT g.id, g.good_name
        FROM good AS g 
        JOIN good_category AS gc 
        ON g.id = gc.good_id
        JOIN category AS c 
        ON gc.category_id = c.id
        WHERE gc.category_id = $1;
		`
	rows, err := s.db.Query(query, categoryId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var r entity.GoodList
		if err := rows.Scan(&r.GoodId, &r.GoodName); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		response = append(response, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}
