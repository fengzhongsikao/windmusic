package local

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	models "windmusic/internal/music"
	"windmusic/music/appdata"

	_ "modernc.org/sqlite"
)

type libraryDB struct {
	db *sql.DB
}

func openLibraryDB() (*libraryDB, error) {
	root, err := appdata.AppDataRootDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(root, "local-library.db")
	sqlDB, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)")
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)

	ldb := &libraryDB{db: sqlDB}
	if err := ldb.initSchema(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return ldb, nil
}

func (d *libraryDB) Close() error {
	if d == nil || d.db == nil {
		return nil
	}
	return d.db.Close()
}

func (d *libraryDB) initSchema() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS scan_entries (
			path TEXT PRIMARY KEY,
			mod_time_unix INTEGER NOT NULL,
			song_json TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS song_extras (
			path TEXT PRIMARY KEY,
			cover_key TEXT,
			lyric TEXT NOT NULL DEFAULT ''
		)`,
	}
	for _, stmt := range stmts {
		if _, err := d.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (d *libraryDB) loadScanCache() (*localScanCacheFile, error) {
	rows, err := d.db.Query(`SELECT path, mod_time_unix, song_json FROM scan_entries`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cache := newScanCache()
	for rows.Next() {
		var path string
		var mod int64
		var songJSON string
		if err := rows.Scan(&path, &mod, &songJSON); err != nil {
			return nil, err
		}
		var song models.LocalSong
		if err := json.Unmarshal([]byte(songJSON), &song); err != nil {
			continue
		}
		cache.Entries[path] = localScanCacheEntry{
			ModTimeUnix: mod,
			Song:        song,
		}
	}
	return cache, rows.Err()
}

func (d *libraryDB) saveScanCache(cache *localScanCacheFile) error {
	if cache == nil {
		cache = newScanCache()
	}
	if cache.Entries == nil {
		cache.Entries = map[string]localScanCacheEntry{}
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO scan_entries(path, mod_time_unix, song_json) VALUES(?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			mod_time_unix = excluded.mod_time_unix,
			song_json = excluded.song_json`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	alive := make([]string, 0, len(cache.Entries))
	for path, entry := range cache.Entries {
		payload, err := json.Marshal(entry.Song)
		if err != nil {
			return err
		}
		if _, err := stmt.Exec(path, entry.ModTimeUnix, string(payload)); err != nil {
			return err
		}
		alive = append(alive, path)
	}
	if err := deleteStalePaths(tx, "scan_entries", alive); err != nil {
		return err
	}
	return tx.Commit()
}

func (d *libraryDB) loadExtras() (*localExtrasFile, error) {
	rows, err := d.db.Query(`SELECT path, cover_key, lyric FROM song_extras`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	extras := newExtrasFile()
	for rows.Next() {
		var path, coverKey, lyric string
		if err := rows.Scan(&path, &coverKey, &lyric); err != nil {
			return nil, err
		}
		extras.Entries[path] = localSongExtraRef{
			CoverKey: strings.TrimSpace(coverKey),
			Lyric:    lyric,
		}
	}
	return extras, rows.Err()
}

func (d *libraryDB) saveExtras(extras *localExtrasFile) error {
	extras.ensureMaps()
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO song_extras(path, cover_key, lyric) VALUES(?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			cover_key = excluded.cover_key,
			lyric = excluded.lyric`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	alive := make([]string, 0, len(extras.Entries))
	for path, entry := range extras.Entries {
		if _, err := stmt.Exec(path, strings.TrimSpace(entry.CoverKey), entry.Lyric); err != nil {
			return err
		}
		alive = append(alive, path)
	}
	if err := deleteStalePaths(tx, "song_extras", alive); err != nil {
		return err
	}
	return tx.Commit()
}

func deleteStalePaths(tx *sql.Tx, table string, alive []string) error {
	if len(alive) == 0 {
		_, err := tx.Exec(`DELETE FROM ` + table)
		return err
	}

	aliveSet := make(map[string]struct{}, len(alive))
	for _, path := range alive {
		aliveSet[path] = struct{}{}
	}

	rows, err := tx.Query(`SELECT path FROM ` + table)
	if err != nil {
		return err
	}
	defer rows.Close()

	stale := make([]string, 0)
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return err
		}
		if _, ok := aliveSet[path]; !ok {
			stale = append(stale, path)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	const chunkSize = 400
	for i := 0; i < len(stale); i += chunkSize {
		end := i + chunkSize
		if end > len(stale) {
			end = len(stale)
		}
		chunk := stale[i:end]
		placeholders := make([]string, len(chunk))
		args := make([]any, len(chunk))
		for j, path := range chunk {
			placeholders[j] = "?"
			args[j] = path
		}
		query := `DELETE FROM ` + table + ` WHERE path IN (` + strings.Join(placeholders, ",") + `)`
		if _, err := tx.Exec(query, args...); err != nil {
			return err
		}
	}
	return nil
}

func (s *LocalLibraryStore) ensureLibraryDB() error {
	if s.db != nil {
		return nil
	}
	ldb, err := openLibraryDB()
	if err != nil {
		return err
	}
	covers, err := newCoverFileStore()
	if err != nil {
		_ = ldb.Close()
		return err
	}
	s.db = ldb
	if s.coverFiles == nil {
		s.coverFiles = covers
	}
	return nil
}

func (s *LocalLibraryStore) closeLibraryDB() {
	if s.db != nil {
		_ = s.db.Close()
		s.db = nil
	}
}
