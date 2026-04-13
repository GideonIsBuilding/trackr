package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourname/job-tracker/internal/model"
)

type AnalyticsStore struct {
	db *pgxpool.Pool
}

func NewAnalyticsStore(db *pgxpool.Pool) *AnalyticsStore {
	return &AnalyticsStore{db: db}
}

func (s *AnalyticsStore) GetSummary(ctx context.Context, userID uuid.UUID) (*model.AnalyticsSummary, error) {
	summary := &model.AnalyticsSummary{}

	// ── 1. Top-level counts & rates ──────────────────────────────────────────
	const topLevel = `
		WITH apps AS (
			SELECT id, status, applied_at, last_activity_at
			FROM applications WHERE user_id = $1
		),
		responded AS (
			SELECT DISTINCT sh.application_id
			FROM status_history sh
			JOIN apps a ON a.id = sh.application_id
			WHERE sh.to_status != 'applied'
		),
		interviewed AS (
			SELECT DISTINCT sh.application_id
			FROM status_history sh
			JOIN apps a ON a.id = sh.application_id
			WHERE sh.to_status IN ('interview','technical_assessment')
		),
		offered AS (
			SELECT DISTINCT sh.application_id
			FROM status_history sh
			JOIN apps a ON a.id = sh.application_id
			WHERE sh.to_status IN ('offer','negotiating','accepted')
		),
		first_response AS (
			SELECT
				sh.application_id,
				EXTRACT(EPOCH FROM (MIN(sh.changed_at) - a.applied_at)) / 86400 AS days
			FROM status_history sh
			JOIN apps a ON a.id = sh.application_id
			WHERE sh.to_status != 'applied'
			GROUP BY sh.application_id, a.applied_at
		)
		SELECT
			COUNT(*)::int,
			COALESCE(ROUND(COUNT(r.application_id) * 100.0 / NULLIF(COUNT(*), 0), 1), 0),
			COALESCE(ROUND(COUNT(i.application_id) * 100.0 / NULLIF(COUNT(*), 0), 1), 0),
			COALESCE(ROUND(COUNT(o.application_id) * 100.0 / NULLIF(COUNT(*), 0), 1), 0),
			COALESCE(ROUND(AVG(fr.days)::numeric, 1), 0)
		FROM apps a
		LEFT JOIN responded   r  ON r.application_id  = a.id
		LEFT JOIN interviewed i  ON i.application_id  = a.id
		LEFT JOIN offered     o  ON o.application_id  = a.id
		LEFT JOIN first_response fr ON fr.application_id = a.id`

	err := s.db.QueryRow(ctx, topLevel, userID).Scan(
		&summary.TotalApplications,
		&summary.ResponseRate,
		&summary.InterviewRate,
		&summary.OfferRate,
		&summary.AvgDaysToResponse,
	)
	if err != nil {
		return nil, fmt.Errorf("top level stats: %w", err)
	}

	// ── 2. Conversion funnel ─────────────────────────────────────────────────
	funnelStages := []struct {
		status model.ApplicationStatus
		label  string
	}{
		{model.StatusApplied, "Applied"},
		{model.StatusPhoneScreen, "Phone screen"},
		{model.StatusInterview, "Interview"},
		{model.StatusTechnicalAssessment, "Technical"},
		{model.StatusOffer, "Offer"},
		{model.StatusAccepted, "Accepted"},
	}

	const funnelQ = `
		SELECT COUNT(DISTINCT sh.application_id)
		FROM status_history sh
		JOIN applications a ON a.id = sh.application_id
		WHERE a.user_id = $1 AND sh.to_status = $2`

	var prevCount int
	for _, stage := range funnelStages {
		var count int
		if stage.status == model.StatusApplied {
			count = summary.TotalApplications
		} else {
			if err := s.db.QueryRow(ctx, funnelQ, userID, stage.status).Scan(&count); err != nil {
				return nil, fmt.Errorf("funnel stage %s: %w", stage.status, err)
			}
		}
		rate := 0.0
		if stage.status == model.StatusApplied {
			rate = 100.0
		} else if prevCount > 0 {
			rate = float64(count) * 100.0 / float64(prevCount)
		}
		prevCount = count
		summary.Funnel = append(summary.Funnel, model.FunnelStage{
			Status: stage.status,
			Label:  stage.label,
			Count:  count,
			Rate:   rate,
		})
	}

	// ── 3. By source ─────────────────────────────────────────────────────────
	const sourceQ = `
		WITH responded AS (
			SELECT DISTINCT sh.application_id
			FROM status_history sh
			JOIN applications a ON a.id = sh.application_id
			WHERE a.user_id = $1 AND sh.to_status != 'applied'
		)
		SELECT
			a.source,
			COUNT(*)::int,
			COUNT(r.application_id)::int,
			COALESCE(ROUND(COUNT(r.application_id) * 100.0 / NULLIF(COUNT(*), 0), 1), 0)
		FROM applications a
		LEFT JOIN responded r ON r.application_id = a.id
		WHERE a.user_id = $1
		GROUP BY a.source
		ORDER BY COUNT(*) DESC`

	sourceLabels := map[model.ApplicationSource]string{
		model.SourceLinkedIn:    "LinkedIn",
		model.SourceReferral:    "Referral",
		model.SourceCompanySite: "Company site",
		model.SourceJobBoard:    "Job board",
		model.SourceRecruiter:   "Recruiter",
		model.SourceOther:       "Other",
	}

	rows, err := s.db.Query(ctx, sourceQ, userID)
	if err != nil {
		return nil, fmt.Errorf("source stats: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ss model.SourceStat
		if err := rows.Scan(&ss.Source, &ss.Total, &ss.Responded, &ss.ResponseRate); err != nil {
			return nil, fmt.Errorf("scanning source stat: %w", err)
		}
		ss.Label = sourceLabels[ss.Source]
		summary.BySource = append(summary.BySource, ss)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating source rows: %w", err)
	}
	rows.Close()

	// ── 4. Checklist correlation — one query, all fields at once ─────────────
	// Avoids dynamic SQL entirely by computing all 5 fields in a single query.
	const correlQ = `
		WITH responded AS (
			SELECT DISTINCT sh.application_id
			FROM status_history sh
			JOIN applications a ON a.id = sh.application_id
			WHERE a.user_id = $1 AND sh.to_status != 'applied'
		)
		SELECT
			-- cover_letter
			COALESCE(ROUND(COUNT(r.application_id) FILTER (WHERE a.cover_letter)      * 100.0 / NULLIF(COUNT(*) FILTER (WHERE a.cover_letter),      0), 1), 0),
			COALESCE(ROUND(COUNT(r.application_id) FILTER (WHERE NOT a.cover_letter)  * 100.0 / NULLIF(COUNT(*) FILTER (WHERE NOT a.cover_letter),  0), 1), 0),
			COUNT(*) FILTER (WHERE a.cover_letter)::int,
			-- cv_tailored
			COALESCE(ROUND(COUNT(r.application_id) FILTER (WHERE a.cv_tailored)       * 100.0 / NULLIF(COUNT(*) FILTER (WHERE a.cv_tailored),       0), 1), 0),
			COALESCE(ROUND(COUNT(r.application_id) FILTER (WHERE NOT a.cv_tailored)   * 100.0 / NULLIF(COUNT(*) FILTER (WHERE NOT a.cv_tailored),   0), 1), 0),
			COUNT(*) FILTER (WHERE a.cv_tailored)::int,
			-- referral
			COALESCE(ROUND(COUNT(r.application_id) FILTER (WHERE a.referral)          * 100.0 / NULLIF(COUNT(*) FILTER (WHERE a.referral),          0), 1), 0),
			COALESCE(ROUND(COUNT(r.application_id) FILTER (WHERE NOT a.referral)      * 100.0 / NULLIF(COUNT(*) FILTER (WHERE NOT a.referral),      0), 1), 0),
			COUNT(*) FILTER (WHERE a.referral)::int,
			-- video_intro
			COALESCE(ROUND(COUNT(r.application_id) FILTER (WHERE a.video_intro)       * 100.0 / NULLIF(COUNT(*) FILTER (WHERE a.video_intro),       0), 1), 0),
			COALESCE(ROUND(COUNT(r.application_id) FILTER (WHERE NOT a.video_intro)   * 100.0 / NULLIF(COUNT(*) FILTER (WHERE NOT a.video_intro),   0), 1), 0),
			COUNT(*) FILTER (WHERE a.video_intro)::int,
			-- linkedin_connect
			COALESCE(ROUND(COUNT(r.application_id) FILTER (WHERE a.linkedin_connect)  * 100.0 / NULLIF(COUNT(*) FILTER (WHERE a.linkedin_connect),  0), 1), 0),
			COALESCE(ROUND(COUNT(r.application_id) FILTER (WHERE NOT a.linkedin_connect) * 100.0 / NULLIF(COUNT(*) FILTER (WHERE NOT a.linkedin_connect), 0), 1), 0),
			COUNT(*) FILTER (WHERE a.linkedin_connect)::int
		FROM applications a
		LEFT JOIN responded r ON r.application_id = a.id
		WHERE a.user_id = $1`

	type row struct {
		with, without float64
		n             int
	}
	var (
		cl, cv, ref, vi, li row
	)
	err = s.db.QueryRow(ctx, correlQ, userID).Scan(
		&cl.with, &cl.without, &cl.n,
		&cv.with, &cv.without, &cv.n,
		&ref.with, &ref.without, &ref.n,
		&vi.with, &vi.without, &vi.n,
		&li.with, &li.without, &li.n,
	)
	if err != nil {
		return nil, fmt.Errorf("checklist correlation: %w", err)
	}

	for _, item := range []struct {
		field string
		label string
		r     row
	}{
		{"cover_letter", "Cover letter", cl},
		{"cv_tailored", "Tailored CV", cv},
		{"referral", "Referral", ref},
		{"video_intro", "Video intro", vi},
		{"linkedin_connect", "LinkedIn connect", li},
	} {
		summary.Checklist = append(summary.Checklist, model.ChecklistCorrelation{
			Field:       item.field,
			Label:       item.label,
			WithItem:    item.r.with,
			WithoutItem: item.r.without,
			Lift:        item.r.with - item.r.without,
			SampleSize:  item.r.n,
		})
	}

	return summary, nil
}
