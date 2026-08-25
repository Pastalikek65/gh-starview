package cache

import (
	"database/sql"
	"strings"

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
	// check existing schema and migrate if needed (old PK was name)
	var sqlText string
	err = db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='repos'`).Scan(&sqlText)
	if err == nil {
		// table exists, check if url is PK
		if !strings.Contains(sqlText, "url TEXT PRIMARY KEY") {
			// migrate: drop and recreate
			if _, err := db.Exec(`DROP TABLE repos`); err != nil {
				return nil, err
			}
		}
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS repos (
		name TEXT,
		description TEXT,
		language TEXT,
		stars INT,
		forks INT,
		updated_at TEXT,
		is_fork INT,
		url TEXT PRIMARY KEY
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
		url := r.URL
		if url == "" {
			// fallback for tests without URL: use synthetic URL based on name
			url = "https://github.com/_/" + r.Name
		}
		if _, err := stmt.Exec(r.Name, r.Description, r.Language, r.Stars, r.Forks, r.UpdatedAt, isFork, url); err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
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
		// hide synthetic url for tests
		if strings.HasPrefix(r.URL, "https://github.com/_/") {
			r.URL = ""
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) Close() error {
	return s.db.Close()
}
