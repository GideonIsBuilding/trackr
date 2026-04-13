export type ApplicationStatus =
  | 'applied' | 'phone_screen' | 'interview' | 'technical_assessment'
  | 'offer' | 'negotiating' | 'accepted' | 'rejected' | 'withdrawn' | 'ghosted'

export type ApplicationSource =
  | 'linkedin' | 'referral' | 'company_site' | 'job_board' | 'recruiter' | 'other'

export interface ApplicationChecklist {
  cover_letter: boolean
  cv_tailored: boolean
  referral: boolean
  portfolio_link?: string
  video_intro: boolean
  linkedin_connect: boolean
}

export interface User {
  id: string
  email: string
  timezone: string
  created_at: string
  updated_at: string
}

export interface Application extends ApplicationChecklist {
  id: string
  user_id: string
  company: string
  role: string
  job_url?: string
  location?: string
  source: ApplicationSource
  status: ApplicationStatus
  applied_at: string
  last_activity_at: string
  notes?: string
  created_at: string
  updated_at: string
}

export interface StatusHistory {
  id: string
  application_id: string
  from_status?: ApplicationStatus
  to_status: ApplicationStatus
  note?: string
  changed_at: string
}

export interface Contact {
  id: string
  application_id: string
  name: string
  email?: string
  role_title?: string
  created_at: string
}

export interface Reminder {
  id: string
  application_id: string
  trigger_after_days: number
  is_active: boolean
  last_sent_at?: string
  snoozed_until?: string
  created_at: string
}

// --- Analytics ---
export interface FunnelStage {
  status: ApplicationStatus
  label: string
  count: number
  rate: number
}

export interface SourceStat {
  source: ApplicationSource
  label: string
  total: number
  responded: number
  response_rate: number
}

export interface ChecklistCorrelation {
  field: string
  label: string
  with_item: number
  without_item: number
  lift: number
  sample_size: number
}

export interface AnalyticsSummary {
  total_applications: number
  response_rate: number
  interview_rate: number
  offer_rate: number
  avg_days_to_response: number
  funnel: FunnelStage[]
  by_source: SourceStat[]
  checklist: ChecklistCorrelation[]
}

export const STATUS_LABELS: Record<ApplicationStatus, string> = {
  applied: 'Applied', phone_screen: 'Phone screen', interview: 'Interview',
  technical_assessment: 'Technical', offer: 'Offer', negotiating: 'Negotiating',
  accepted: 'Accepted', rejected: 'Rejected', withdrawn: 'Withdrawn', ghosted: 'Ghosted',
}

export const SOURCE_LABELS: Record<ApplicationSource, string> = {
  linkedin: 'LinkedIn', referral: 'Referral', company_site: 'Company site',
  job_board: 'Job board', recruiter: 'Recruiter', other: 'Other',
}

export const ALL_STATUSES: ApplicationStatus[] = [
  'applied', 'phone_screen', 'interview', 'technical_assessment',
  'offer', 'negotiating', 'accepted', 'rejected', 'withdrawn', 'ghosted',
]

export const TERMINAL_STATUSES: ApplicationStatus[] = ['accepted', 'rejected', 'withdrawn']
