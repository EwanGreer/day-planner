// Package sqlite provides a SQLite-backed implementation of core.Store.
package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EwanGreer/day-planner/internal/core"
	_ "modernc.org/sqlite" // register the sqlite driver
)

// Compile-time assertion that SQLiteStore implements core.Store.
var _ core.Store = (*SQLiteStore)(nil)

const (
	dateFmt      = "2006-01-02"
	timestampFmt = time.RFC3339
)

// SQLiteStore is a core.Store backed by a local SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

// New opens (or creates) the SQLite database at path and runs all pending
// schema migrations. Pass any path; tests typically pass t.TempDir()+"/test.db".
func New(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

// Close releases the underlying database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// --- DayPlan ---

// SaveDayPlan persists a DayPlan, replacing any existing row for the same date.
func (s *SQLiteStore) SaveDayPlan(plan *core.DayPlan) error {
	tasksJSON, err := json.Marshal(plan.Tasks)
	if err != nil {
		return fmt.Errorf("marshal tasks: %w", err)
	}

	goalsJSON, err := json.Marshal(plan.Goals)
	if err != nil {
		return fmt.Errorf("marshal goals: %w", err)
	}

	createdAt := plan.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	_, err = s.db.Exec(
		`INSERT OR REPLACE INTO day_plans (date, tasks, goals, score, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		plan.Date.Format(dateFmt),
		string(tasksJSON),
		string(goalsJSON),
		plan.Score,
		createdAt.Format(timestampFmt),
	)
	return err
}

// LoadDayPlan retrieves the DayPlan for the given date.
// Returns nil, nil when no row exists for that date.
func (s *SQLiteStore) LoadDayPlan(date time.Time) (*core.DayPlan, error) {
	row := s.db.QueryRow(
		`SELECT date, tasks, goals, score, created_at FROM day_plans WHERE date = ?`,
		date.Format(dateFmt),
	)

	var (
		dateStr      string
		tasksJSON    string
		goalsJSON    string
		score        int
		createdAtStr string
	)

	if err := row.Scan(&dateStr, &tasksJSON, &goalsJSON, &score, &createdAtStr); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("scan day_plan: %w", err)
	}

	parsedDate, err := time.Parse(dateFmt, dateStr)
	if err != nil {
		return nil, fmt.Errorf("parse date: %w", err)
	}

	createdAt, err := time.Parse(timestampFmt, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	var tasks []core.Task
	if err := json.Unmarshal([]byte(tasksJSON), &tasks); err != nil {
		return nil, fmt.Errorf("unmarshal tasks: %w", err)
	}

	var goals []string
	if err := json.Unmarshal([]byte(goalsJSON), &goals); err != nil {
		return nil, fmt.Errorf("unmarshal goals: %w", err)
	}

	return &core.DayPlan{
		Date:      parsedDate,
		Tasks:     tasks,
		Goals:     goals,
		Score:     score,
		CreatedAt: createdAt,
	}, nil
}

// --- CompletionRecord ---

// SaveCompletion inserts a new completion record.
func (s *SQLiteStore) SaveCompletion(record core.CompletionRecord) error {
	_, err := s.db.Exec(
		`INSERT INTO completion_records (date, task_id, status, completed_at)
		 VALUES (?, ?, ?, ?)`,
		record.Date.Format(dateFmt),
		record.TaskID,
		string(record.Status),
		record.CompletedAt.Format(timestampFmt),
	)
	return err
}

// LoadCompletions returns all completion records for the given date.
func (s *SQLiteStore) LoadCompletions(date time.Time) ([]core.CompletionRecord, error) {
	rows, err := s.db.Query(
		`SELECT date, task_id, status, completed_at FROM completion_records WHERE date = ?`,
		date.Format(dateFmt),
	)
	if err != nil {
		return nil, fmt.Errorf("query completions: %w", err)
	}
	defer rows.Close()

	var records []core.CompletionRecord
	for rows.Next() {
		var (
			dateStr      string
			taskID       string
			status       string
			completedStr string
		)
		if err := rows.Scan(&dateStr, &taskID, &status, &completedStr); err != nil {
			return nil, fmt.Errorf("scan completion_record: %w", err)
		}

		parsedDate, err := time.Parse(dateFmt, dateStr)
		if err != nil {
			return nil, fmt.Errorf("parse date: %w", err)
		}

		completedAt, err := time.Parse(timestampFmt, completedStr)
		if err != nil {
			return nil, fmt.Errorf("parse completed_at: %w", err)
		}

		records = append(records, core.CompletionRecord{
			Date:        parsedDate,
			TaskID:      taskID,
			Status:      core.CompletionStatus(status),
			CompletedAt: completedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate completions: %w", err)
	}

	return records, nil
}

// --- Streak ---

// SaveStreak persists the singleton streak row, replacing any existing row.
func (s *SQLiteStore) SaveStreak(streak core.Streak) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO streaks (id, current, longest, last_active_date)
		 VALUES (1, ?, ?, ?)`,
		streak.Current,
		streak.Longest,
		streak.LastActiveDate.Format(dateFmt),
	)
	return err
}

// LoadStreak returns the current streak. Returns a zero-value &Streak{} (not
// nil, not an error) when no streak row has been saved yet.
func (s *SQLiteStore) LoadStreak() (*core.Streak, error) {
	row := s.db.QueryRow(
		`SELECT current, longest, last_active_date FROM streaks WHERE id = 1`,
	)

	var (
		current        int
		longest        int
		lastActiveSstr string
	)

	if err := row.Scan(&current, &longest, &lastActiveSstr); err == sql.ErrNoRows {
		return &core.Streak{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("scan streak: %w", err)
	}

	lastActive, err := time.Parse(dateFmt, lastActiveSstr)
	if err != nil {
		return nil, fmt.Errorf("parse last_active_date: %w", err)
	}

	return &core.Streak{
		Current:        current,
		Longest:        longest,
		LastActiveDate: lastActive,
	}, nil
}

// --- Reflection ---

// SaveReflection inserts a new reflection.
func (s *SQLiteStore) SaveReflection(reflection core.Reflection) error {
	createdAt := reflection.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	_, err := s.db.Exec(
		`INSERT INTO reflections (date, text, created_at) VALUES (?, ?, ?)`,
		reflection.Date.Format(dateFmt),
		reflection.Text,
		createdAt.Format(timestampFmt),
	)
	return err
}

// LoadReflections returns all reflections for the given date.
func (s *SQLiteStore) LoadReflections(date time.Time) ([]core.Reflection, error) {
	rows, err := s.db.Query(
		`SELECT date, text, created_at FROM reflections WHERE date = ?`,
		date.Format(dateFmt),
	)
	if err != nil {
		return nil, fmt.Errorf("query reflections: %w", err)
	}
	defer rows.Close()

	var reflections []core.Reflection
	for rows.Next() {
		var (
			dateStr      string
			text         string
			createdAtStr string
		)
		if err := rows.Scan(&dateStr, &text, &createdAtStr); err != nil {
			return nil, fmt.Errorf("scan reflection: %w", err)
		}

		parsedDate, err := time.Parse(dateFmt, dateStr)
		if err != nil {
			return nil, fmt.Errorf("parse date: %w", err)
		}

		createdAt, err := time.Parse(timestampFmt, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}

		reflections = append(reflections, core.Reflection{
			Date:      parsedDate,
			Text:      text,
			CreatedAt: createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reflections: %w", err)
	}

	return reflections, nil
}

// --- NudgeWindow ---

// SaveNudgeWindow inserts a new nudge window.
func (s *SQLiteStore) SaveNudgeWindow(window core.NudgeWindow) error {
	_, err := s.db.Exec(
		`INSERT INTO nudge_windows (start_time, end_time, label) VALUES (?, ?, ?)`,
		window.StartTime.Format(timestampFmt),
		window.EndTime.Format(timestampFmt),
		window.Label,
	)
	return err
}

// LoadNudgeWindows returns all nudge windows.
func (s *SQLiteStore) LoadNudgeWindows() ([]core.NudgeWindow, error) {
	rows, err := s.db.Query(`SELECT start_time, end_time, label FROM nudge_windows`)
	if err != nil {
		return nil, fmt.Errorf("query nudge_windows: %w", err)
	}
	defer rows.Close()

	var windows []core.NudgeWindow
	for rows.Next() {
		var (
			startStr string
			endStr   string
			label    string
		)
		if err := rows.Scan(&startStr, &endStr, &label); err != nil {
			return nil, fmt.Errorf("scan nudge_window: %w", err)
		}

		startTime, err := time.Parse(timestampFmt, startStr)
		if err != nil {
			return nil, fmt.Errorf("parse start_time: %w", err)
		}

		endTime, err := time.Parse(timestampFmt, endStr)
		if err != nil {
			return nil, fmt.Errorf("parse end_time: %w", err)
		}

		windows = append(windows, core.NudgeWindow{
			StartTime: startTime,
			EndTime:   endTime,
			Label:     label,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate nudge_windows: %w", err)
	}

	return windows, nil
}
