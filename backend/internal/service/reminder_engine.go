package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/yourname/job-tracker/internal/model"
	"github.com/yourname/job-tracker/internal/store"
)

// Notifier is the interface the reminder engine uses to deliver alerts.
// Swap in email, push notifications, webhooks, etc. without touching engine logic.
type Notifier interface {
	Notify(ctx context.Context, alert *model.ReminderAlert) error
}

type ReminderEngine struct {
	reminders *store.ReminderStore
	notifier  Notifier
	interval  time.Duration
	log       *slog.Logger
}

func NewReminderEngine(
	reminders *store.ReminderStore,
	notifier Notifier,
	interval time.Duration,
	log *slog.Logger,
) *ReminderEngine {
	return &ReminderEngine{
		reminders: reminders,
		notifier:  notifier,
		interval:  interval,
		log:       log,
	}
}

// Run starts the reminder engine loop. It blocks until ctx is cancelled.
// Call this in its own goroutine: go engine.Run(ctx)
func (e *ReminderEngine) Run(ctx context.Context) {
	e.log.Info("reminder engine started", "check_interval", e.interval)
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	// Run once immediately at startup
	e.scan(ctx)

	for {
		select {
		case <-ctx.Done():
			e.log.Info("reminder engine shutting down")
			return
		case <-ticker.C:
			e.scan(ctx)
		}
	}
}

func (e *ReminderEngine) scan(ctx context.Context) {
	e.log.Debug("scanning for due reminders")

	alerts, err := e.reminders.ScanDue(ctx)
	if err != nil {
		e.log.Error("scanning reminders failed", "error", err)
		return
	}

	if len(alerts) == 0 {
		e.log.Debug("no due reminders")
		return
	}

	e.log.Info("due reminders found", "count", len(alerts))

	for _, alert := range alerts {
		if err := e.notify(ctx, alert); err != nil {
			e.log.Error("notification failed",
				"reminder_id", alert.Reminder.ID,
				"application_id", alert.Application.ID,
				"error", err,
			)
		}
	}
}

func (e *ReminderEngine) notify(ctx context.Context, alert *model.ReminderAlert) error {
	if err := e.notifier.Notify(ctx, alert); err != nil {
		return fmt.Errorf("notifier: %w", err)
	}

	// Mark as sent so it doesn't fire again until next silence window
	if err := e.reminders.MarkSent(ctx, alert.Reminder.ID); err != nil {
		return fmt.Errorf("marking reminder sent: %w", err)
	}

	e.log.Info("reminder sent",
		"user_email", alert.User.Email,
		"company", alert.Application.Company,
		"role", alert.Application.Role,
		"silent_days", alert.SilentDays,
	)
	return nil
}

// --- Log-only notifier (default, swap for email in production) ---

type LogNotifier struct {
	log *slog.Logger
}

func NewLogNotifier(log *slog.Logger) *LogNotifier {
	return &LogNotifier{log: log}
}

func (n *LogNotifier) Notify(ctx context.Context, alert *model.ReminderAlert) error {
	n.log.Warn("FOLLOW-UP REMINDER",
		"user", alert.User.Email,
		"company", alert.Application.Company,
		"role", alert.Application.Role,
		"status", alert.Application.Status,
		"silent_days", alert.SilentDays,
		"suggestion", fmt.Sprintf(
			"You applied to %s at %s %d days ago with no update. Consider sending a follow-up email.",
			alert.Application.Role, alert.Application.Company, alert.SilentDays,
		),
	)
	return nil
}
