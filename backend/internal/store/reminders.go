package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourname/job-tracker/internal/model"
)

// --- Contact Store ---

type ContactStore struct {
	db *pgxpool.Pool
}

func NewContactStore(db *pgxpool.Pool) *ContactStore {
	return &ContactStore{db: db}
}

func (s *ContactStore) Create(ctx context.Context, appID uuid.UUID, name string, email, roleTitle *string) (*model.Contact, error) {
	const q = `
		INSERT INTO contacts (application_id, name, email, role_title)
		VALUES ($1,$2,$3,$4)
		RETURNING id, application_id, name, email, role_title, created_at`

	var c model.Contact
	err := s.db.QueryRow(ctx, q, appID, name, email, roleTitle).Scan(
		&c.ID, &c.ApplicationID, &c.Name, &c.Email, &c.RoleTitle, &c.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating contact: %w", err)
	}
	return &c, nil
}

func (s *ContactStore) ListByApplication(ctx context.Context, appID uuid.UUID) ([]*model.Contact, error) {
	const q = `
		SELECT id, application_id, name, email, role_title, created_at
		FROM contacts WHERE application_id = $1 ORDER BY created_at ASC`

	rows, err := s.db.Query(ctx, q, appID)
	if err != nil {
		return nil, fmt.Errorf("listing contacts: %w", err)
	}
	defer rows.Close()

	var contacts []*model.Contact
	for rows.Next() {
		var c model.Contact
		if err := rows.Scan(&c.ID, &c.ApplicationID, &c.Name, &c.Email, &c.RoleTitle, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning contact: %w", err)
		}
		contacts = append(contacts, &c)
	}
	return contacts, rows.Err()
}

// --- Reminder Store ---

type ReminderStore struct {
	db *pgxpool.Pool
}

func NewReminderStore(db *pgxpool.Pool) *ReminderStore {
	return &ReminderStore{db: db}
}

func (s *ReminderStore) Upsert(ctx context.Context, appID uuid.UUID, triggerAfterDays int) (*model.Reminder, error) {
	const q = `
		INSERT INTO reminders (application_id, trigger_after_days)
		VALUES ($1,$2)
		ON CONFLICT (application_id) DO UPDATE
		  SET trigger_after_days = EXCLUDED.trigger_after_days,
		      is_active = TRUE
		RETURNING id, application_id, trigger_after_days, is_active, last_sent_at, snoozed_until, created_at`

	var r model.Reminder
	err := s.db.QueryRow(ctx, q, appID, triggerAfterDays).Scan(
		&r.ID, &r.ApplicationID, &r.TriggerAfterDays,
		&r.IsActive, &r.LastSentAt, &r.SnoozedUntil, &r.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upserting reminder: %w", err)
	}
	return &r, nil
}

func (s *ReminderStore) Snooze(ctx context.Context, appID uuid.UUID, until time.Time) error {
	const q = `UPDATE reminders SET snoozed_until = $2 WHERE application_id = $1`
	_, err := s.db.Exec(ctx, q, appID, until)
	return err
}

func (s *ReminderStore) MarkSent(ctx context.Context, reminderID uuid.UUID) error {
	const q = `UPDATE reminders SET last_sent_at = NOW() WHERE id = $1`
	_, err := s.db.Exec(ctx, q, reminderID)
	return err
}

func (s *ReminderStore) ScanDue(ctx context.Context) ([]*model.ReminderAlert, error) {
	const q = `
		SELECT
			r.id, r.application_id, r.trigger_after_days, r.is_active, r.last_sent_at, r.snoozed_until, r.created_at,
			a.id, a.user_id, a.company, a.role, a.job_url, a.location, a.source,
			a.status, a.applied_at, a.last_activity_at, a.notes,
			a.cover_letter, a.cv_tailored, a.referral, a.portfolio_link, a.video_intro, a.linkedin_connect,
			a.created_at, a.updated_at,
			u.id, u.email, u.password_hash, u.timezone, u.created_at, u.updated_at,
			EXTRACT(EPOCH FROM (NOW() - a.last_activity_at)) / 86400 AS silent_days
		FROM reminders r
		JOIN applications a ON a.id = r.application_id
		JOIN users       u ON u.id = a.user_id
		WHERE r.is_active = TRUE
		  AND (r.snoozed_until IS NULL OR r.snoozed_until < NOW())
		  AND a.status NOT IN ('accepted','rejected','withdrawn')
		  AND (NOW() - a.last_activity_at) >= (r.trigger_after_days || ' days')::INTERVAL`

	rows, err := s.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("scanning due reminders: %w", err)
	}
	defer rows.Close()

	var alerts []*model.ReminderAlert
	for rows.Next() {
		var alert model.ReminderAlert
		var silentDays float64
		r := &alert.Reminder
		a := &alert.Application
		u := &alert.User

		if err := rows.Scan(
			&r.ID, &r.ApplicationID, &r.TriggerAfterDays, &r.IsActive,
			&r.LastSentAt, &r.SnoozedUntil, &r.CreatedAt,
			&a.ID, &a.UserID, &a.Company, &a.Role, &a.JobURL, &a.Location,
			&a.Source, &a.Status, &a.AppliedAt, &a.LastActivityAt, &a.Notes,
			&a.CoverLetter, &a.CVTailored, &a.Referral, &a.PortfolioLink,
			&a.VideoIntro, &a.LinkedInConnect,
			&a.CreatedAt, &a.UpdatedAt,
			&u.ID, &u.Email, &u.PasswordHash, &u.Timezone, &u.CreatedAt, &u.UpdatedAt,
			&silentDays,
		); err != nil {
			return nil, fmt.Errorf("scanning reminder alert: %w", err)
		}
		alert.SilentDays = int(silentDays)
		alerts = append(alerts, &alert)
	}
	return alerts, rows.Err()
}
