package model

import (
	"time"

	"github.com/google/uuid"
)

// --- Enums ---

type ApplicationStatus string

const (
	StatusApplied             ApplicationStatus = "applied"
	StatusPhoneScreen         ApplicationStatus = "phone_screen"
	StatusInterview           ApplicationStatus = "interview"
	StatusTechnicalAssessment ApplicationStatus = "technical_assessment"
	StatusOffer               ApplicationStatus = "offer"
	StatusNegotiating         ApplicationStatus = "negotiating"
	StatusAccepted            ApplicationStatus = "accepted"
	StatusRejected            ApplicationStatus = "rejected"
	StatusWithdrawn           ApplicationStatus = "withdrawn"
	StatusGhosted             ApplicationStatus = "ghosted"
)

func (s ApplicationStatus) IsValid() bool {
	switch s {
	case StatusApplied, StatusPhoneScreen, StatusInterview,
		StatusTechnicalAssessment, StatusOffer, StatusNegotiating,
		StatusAccepted, StatusRejected, StatusWithdrawn, StatusGhosted:
		return true
	}
	return false
}

func (s ApplicationStatus) IsTerminal() bool {
	switch s {
	case StatusAccepted, StatusRejected, StatusWithdrawn:
		return true
	}
	return false
}

type ApplicationSource string

const (
	SourceLinkedIn    ApplicationSource = "linkedin"
	SourceReferral    ApplicationSource = "referral"
	SourceCompanySite ApplicationSource = "company_site"
	SourceJobBoard    ApplicationSource = "job_board"
	SourceRecruiter   ApplicationSource = "recruiter"
	SourceOther       ApplicationSource = "other"
)

func (s ApplicationSource) IsValid() bool {
	switch s {
	case SourceLinkedIn, SourceReferral, SourceCompanySite,
		SourceJobBoard, SourceRecruiter, SourceOther:
		return true
	}
	return false
}

// --- Domain models ---

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Timezone     string    `json:"timezone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ApplicationChecklist captures what extras were submitted with an application.
// Used for outcome correlation analytics.
type ApplicationChecklist struct {
	CoverLetter     bool    `json:"cover_letter"`
	CVTailored      bool    `json:"cv_tailored"`
	Referral        bool    `json:"referral"`
	PortfolioLink   *string `json:"portfolio_link,omitempty"`
	VideoIntro      bool    `json:"video_intro"`
	LinkedInConnect bool    `json:"linkedin_connect"`
}

type Application struct {
	ID             uuid.UUID         `json:"id"`
	UserID         uuid.UUID         `json:"user_id"`
	Company        string            `json:"company"`
	Role           string            `json:"role"`
	JobURL         *string           `json:"job_url,omitempty"`
	Location       *string           `json:"location,omitempty"`
	Source         ApplicationSource `json:"source"`
	Status         ApplicationStatus `json:"status"`
	AppliedAt      time.Time         `json:"applied_at"`
	LastActivityAt time.Time         `json:"last_activity_at"`
	Notes          *string           `json:"notes,omitempty"`

	// Checklist
	ApplicationChecklist

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StatusHistory struct {
	ID            uuid.UUID          `json:"id"`
	ApplicationID uuid.UUID          `json:"application_id"`
	FromStatus    *ApplicationStatus `json:"from_status,omitempty"`
	ToStatus      ApplicationStatus  `json:"to_status"`
	Note          *string            `json:"note,omitempty"`
	ChangedAt     time.Time          `json:"changed_at"`
}

type Contact struct {
	ID            uuid.UUID `json:"id"`
	ApplicationID uuid.UUID `json:"application_id"`
	Name          string    `json:"name"`
	Email         *string   `json:"email,omitempty"`
	RoleTitle     *string   `json:"role_title,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type Reminder struct {
	ID               uuid.UUID  `json:"id"`
	ApplicationID    uuid.UUID  `json:"application_id"`
	TriggerAfterDays int        `json:"trigger_after_days"`
	IsActive         bool       `json:"is_active"`
	LastSentAt       *time.Time `json:"last_sent_at,omitempty"`
	SnoozedUntil     *time.Time `json:"snoozed_until,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type ReminderAlert struct {
	Reminder    Reminder
	Application Application
	User        User
	SilentDays  int
}

// --- Analytics types ---

// FunnelStage represents one step in the application conversion funnel.
type FunnelStage struct {
	Status ApplicationStatus `json:"status"`
	Label  string            `json:"label"`
	Count  int               `json:"count"`
	Rate   float64           `json:"rate"` // percentage of previous stage
}

// SourceStat shows response rate broken down by where the job was found.
type SourceStat struct {
	Source       ApplicationSource `json:"source"`
	Label        string            `json:"label"`
	Total        int               `json:"total"`
	Responded    int               `json:"responded"` // moved past "applied"
	ResponseRate float64           `json:"response_rate"`
}

// ChecklistCorrelation shows how much each checklist item correlates with getting a response.
type ChecklistCorrelation struct {
	Field       string  `json:"field"`
	Label       string  `json:"label"`
	WithItem    float64 `json:"with_item"`    // response rate when item was present
	WithoutItem float64 `json:"without_item"` // response rate when item was absent
	Lift        float64 `json:"lift"`         // difference (with - without)
	SampleSize  int     `json:"sample_size"`  // total apps with this item
}

// AnalyticsSummary is the full payload returned by GET /api/analytics
type AnalyticsSummary struct {
	TotalApplications int                    `json:"total_applications"`
	ResponseRate      float64                `json:"response_rate"`
	InterviewRate     float64                `json:"interview_rate"`
	OfferRate         float64                `json:"offer_rate"`
	AvgDaysToResponse float64                `json:"avg_days_to_response"`
	Funnel            []FunnelStage          `json:"funnel"`
	BySource          []SourceStat           `json:"by_source"`
	Checklist         []ChecklistCorrelation `json:"checklist"`
}
