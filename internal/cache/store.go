package cache

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Repo struct {
	Name        string
	Description string
	Language    string
	Stars       int
	Forks       int
	UpdatedAt   string
	IsFork      bool
	URL         string
}

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		return nil, err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS repos (
		name TEXT PRIMARY KEY,
		description TEXT,
		language TEXT,
		stars INT,
		forks INT,
		updated_at TEXT,
		is_fork INT,
		url TEXT
	)`); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Upsert(repos []Repo) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO repos(name,description,language,stars,forks,updated_at,is_fork,url) VALUES(?,?,?,?,?,?,?,?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, r := range repos {
		isFork := 0
		if r.IsFork {
			isFork = 1
		}
		if _, err := stmt.Exec(r.Name, r.Description, r.Language, r.Stars, r.Forks, r.UpdatedAt, isFork, r.URL); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) List(sortBy string) ([]Repo, error) {
	order := "stars DESC"
	switch sortBy {
	case "name":
		order = "name ASC"
	case "updated":
		order = "updated_at DESC"
	case "forks":
		order = "forks DESC"
	}
	rows, err := s.db.Query(`SELECT name,description,language,stars,forks,updated_at,is_fork,url FROM repos ORDER BY ` + order)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Repo
	for rows.Next() {
		var r Repo
		var isFork int
		if err := rows.Scan(&r.Name, &r.Description, &r.Language, &r.Stars, &r.Forks, &r.UpdatedAt, &isFork, &r.URL); err != nil {
			return nil, err
		}
		r.IsFork = isFork == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) Close() error {
	return s.db.Close()
}
