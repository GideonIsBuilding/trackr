package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourname/job-tracker/internal/model"
)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

type ApplicationStore struct {
	db *pgxpool.Pool
}

func NewApplicationStore(db *pgxpool.Pool) *ApplicationStore {
	return &ApplicationStore{db: db}
}

type CreateApplicationInput struct {
	UserID          uuid.UUID
	Company         string
	Role            string
	JobURL          *string
	Location        *string
	Source          model.ApplicationSource
	Notes           *string
	AppliedAt       time.Time
	CoverLetter     bool
	CVTailored      bool
	Referral        bool
	PortfolioLink   *string
	VideoIntro      bool
	LinkedInConnect bool
}

// scanApplication scans all columns including checklist fields.
func scanApplication(row pgx.Row, app *model.Application) error {
	return row.Scan(
		&app.ID, &app.UserID, &app.Company, &app.Role, &app.JobURL,
		&app.Location, &app.Source, &app.Status, &app.AppliedAt,
		&app.LastActivityAt, &app.Notes,
		&app.CoverLetter, &app.CVTailored, &app.Referral,
		&app.PortfolioLink, &app.VideoIntro, &app.LinkedInConnect,
		&app.CreatedAt, &app.UpdatedAt,
	)
}

const appColumns = `
	id, user_id, company, role, job_url, location, source,
	status, applied_at, last_activity_at, notes,
	cover_letter, cv_tailored, referral,
	portfolio_link, video_intro, linkedin_connect,
	created_at, updated_at`

func (s *ApplicationStore) Create(ctx context.Context, in CreateApplicationInput) (*model.Application, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const insertApp = `
		INSERT INTO applications
			(user_id, company, role, job_url, location, source, status, applied_at, notes,
			 cover_letter, cv_tailored, referral, portfolio_link, video_intro, linkedin_connect)
		VALUES ($1,$2,$3,$4,$5,$6,'applied',$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING ` + appColumns

	var app model.Application
	err = scanApplication(tx.QueryRow(ctx, insertApp,
		in.UserID, in.Company, in.Role, in.JobURL, in.Location, in.Source,
		in.AppliedAt, in.Notes,
		in.CoverLetter, in.CVTailored, in.Referral,
		in.PortfolioLink, in.VideoIntro, in.LinkedInConnect,
	), &app)
	if err != nil {
		return nil, fmt.Errorf("inserting application: %w", err)
	}

	const insertHistory = `
		INSERT INTO status_history (application_id, from_status, to_status, note)
		VALUES ($1, NULL, 'applied', 'Application submitted')`
	if _, err = tx.Exec(ctx, insertHistory, app.ID); err != nil {
		return nil, fmt.Errorf("seeding status history: %w", err)
	}

	const insertReminder = `
		INSERT INTO reminders (application_id, trigger_after_days)
		VALUES ($1, 14) ON CONFLICT (application_id) DO NOTHING`
	if _, err = tx.Exec(ctx, insertReminder, app.ID); err != nil {
		return nil, fmt.Errorf("creating reminder: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}
	return &app, nil
}

func (s *ApplicationStore) List(ctx context.Context, userID uuid.UUID, status model.ApplicationStatus) ([]*model.Application, error) {
	q := `SELECT ` + appColumns + ` FROM applications WHERE user_id = $1`
	args := []any{userID}

	if status != "" {
		q += " AND status = $2"
		args = append(args, status)
	}
	q += " ORDER BY applied_at DESC"

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("listing applications: %w", err)
	}
	defer rows.Close()

	var apps []*model.Application
	for rows.Next() {
		var app model.Application
		if err := scanApplication(rows, &app); err != nil {
			return nil, fmt.Errorf("scanning application: %w", err)
		}
		apps = append(apps, &app)
	}
	return apps, rows.Err()
}

func (s *ApplicationStore) GetByID(ctx context.Context, id, userID uuid.UUID) (*model.Application, error) {
	q := `SELECT ` + appColumns + ` FROM applications WHERE id = $1 AND user_id = $2`
	var app model.Application
	err := scanApplication(s.db.QueryRow(ctx, q, id, userID), &app)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("fetching application: %w", err)
	}
	return &app, nil
}

// UpdateChecklist updates the checklist fields for an existing application.
func (s *ApplicationStore) UpdateChecklist(ctx context.Context, appID, userID uuid.UUID, c model.ApplicationChecklist) (*model.Application, error) {
	const q = `
		UPDATE applications SET
			cover_letter     = $3,
			cv_tailored      = $4,
			referral         = $5,
			portfolio_link   = $6,
			video_intro      = $7,
			linkedin_connect = $8,
			updated_at       = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING ` + appColumns

	var app model.Application
	err := scanApplication(s.db.QueryRow(ctx, q,
		appID, userID,
		c.CoverLetter, c.CVTailored, c.Referral,
		c.PortfolioLink, c.VideoIntro, c.LinkedInConnect,
	), &app)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("updating checklist: %w", err)
	}
	return &app, nil
}

func (s *ApplicationStore) UpdateStatus(ctx context.Context, appID, userID uuid.UUID, toStatus model.ApplicationStatus, note *string) (*model.StatusHistory, error) {
	app, err := s.GetByID(ctx, appID, userID)
	if err != nil {
		return nil, err
	}

	const q = `
		INSERT INTO status_history (application_id, from_status, to_status, note)
		VALUES ($1, $2, $3, $4)
		RETURNING id, application_id, from_status, to_status, note, changed_at`

	var h model.StatusHistory
	err = s.db.QueryRow(ctx, q, appID, app.Status, toStatus, note).Scan(
		&h.ID, &h.ApplicationID, &h.FromStatus, &h.ToStatus, &h.Note, &h.ChangedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting status history: %w", err)
	}
	return &h, nil
}

func (s *ApplicationStore) GetStatusHistory(ctx context.Context, appID, userID uuid.UUID) ([]*model.StatusHistory, error) {
	if _, err := s.GetByID(ctx, appID, userID); err != nil {
		return nil, err
	}

	const q = `
		SELECT id, application_id, from_status, to_status, note, changed_at
		FROM status_history WHERE application_id = $1
		ORDER BY changed_at ASC`

	rows, err := s.db.Query(ctx, q, appID)
	if err != nil {
		return nil, fmt.Errorf("fetching status history: %w", err)
	}
	defer rows.Close()

	var history []*model.StatusHistory
	for rows.Next() {
		var h model.StatusHistory
		if err := rows.Scan(&h.ID, &h.ApplicationID, &h.FromStatus, &h.ToStatus, &h.Note, &h.ChangedAt); err != nil {
			return nil, fmt.Errorf("scanning status history: %w", err)
		}
		history = append(history, &h)
	}
	return history, rows.Err()
}

func (s *ApplicationStore) Delete(ctx context.Context, appID, userID uuid.UUID) error {
	const q = `DELETE FROM applications WHERE id = $1 AND user_id = $2`
	result, err := s.db.Exec(ctx, q, appID, userID)
	if err != nil {
		return fmt.Errorf("deleting application: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
