package domain

import (
	"time"
)

// AuditAction represents the type of action performed
type AuditAction string

const (
	AuditActionCreate AuditAction = "create"
	AuditActionRead   AuditAction = "read"
	AuditActionUpdate AuditAction = "update"
	AuditActionDelete AuditAction = "delete"
	AuditActionLogin  AuditAction = "login"
	AuditActionLogout AuditAction = "logout"
	AuditActionExport AuditAction = "export"
	AuditActionImport AuditAction = "import"
)

// AuditLog represents an audit log entry for tracking user actions
type AuditLog struct {
	ID           uint        `gorm:"primarykey" json:"id"`
	UserID       *uint       `gorm:"index" json:"user_id,omitempty"`
	Action       AuditAction `gorm:"type:varchar(50);not null;index" json:"action"`
	Resource     string      `gorm:"size:100;not null;index" json:"resource"` // e.g., "users", "posts"
	ResourceID   *uint       `gorm:"index" json:"resource_id,omitempty"`
	Description  string      `gorm:"type:text" json:"description"`
	IPAddress    string      `gorm:"size:45" json:"ip_address"`
	UserAgent    string      `gorm:"size:500" json:"user_agent"`
	Method       string      `gorm:"size:10" json:"method"` // HTTP method
	Path         string      `gorm:"size:500" json:"path"`  // Request path
	StatusCode   int         `json:"status_code"`
	Duration     int64       `json:"duration"`                                 // Request duration in milliseconds
	RequestBody  *string     `gorm:"type:text" json:"request_body,omitempty"`  // Full request body (CLOB)
	ResponseBody *string     `gorm:"type:text" json:"response_body,omitempty"` // Full response body (CLOB)
	OldValues    *string     `gorm:"type:jsonb" json:"old_values,omitempty"`   // JSON of old values (for updates)
	NewValues    *string     `gorm:"type:jsonb" json:"new_values,omitempty"`   // JSON of new values (for updates)
	Metadata     *string     `gorm:"type:jsonb" json:"metadata,omitempty"`     // Additional metadata
	CreatedAt    time.Time   `gorm:"index" json:"created_at"`                  // Start time
	FinishedAt   *time.Time  `gorm:"index" json:"finished_at,omitempty"`       // Finish time

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}

// SystemSetting represents system-wide configuration settings
type SystemSetting struct {
	BaseModel
	Key         string `gorm:"uniqueIndex;size:200;not null" json:"key"`
	Value       string `gorm:"type:text" json:"value"`
	Type        string `gorm:"size:50;not null" json:"type"` // string, number, boolean, json
	Category    string `gorm:"size:100" json:"category"`
	Description string `gorm:"type:text" json:"description"`
	IsPublic    bool   `gorm:"default:false" json:"is_public"`  // Can be accessed without authentication
	IsEditable  bool   `gorm:"default:true" json:"is_editable"` // Can be edited via UI
}

// TableName specifies the table name for SystemSetting
func (SystemSetting) TableName() string {
	return "system_settings"
}

// ActivityLog represents user activity tracking
type ActivityLog struct {
	BaseModel
	UserID      uint   `gorm:"index;not null" json:"user_id"`
	Activity    string `gorm:"size:200;not null" json:"activity"`
	Description string `gorm:"type:text" json:"description"`
	IPAddress   string `gorm:"size:45" json:"ip_address"`
	UserAgent   string `gorm:"size:500" json:"user_agent"`
	Metadata    string `gorm:"type:jsonb" json:"metadata,omitempty"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for ActivityLog
func (ActivityLog) TableName() string {
	return "activity_logs"
}

// EmailTemplate represents email templates
type EmailTemplate struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;size:200;not null" json:"name"`
	Subject     string `gorm:"size:500;not null" json:"subject"`
	Body        string `gorm:"type:text;not null" json:"body"`
	Type        string `gorm:"size:50;not null" json:"type"` // html, text
	Category    string `gorm:"size:100" json:"category"`
	Variables   string `gorm:"type:jsonb" json:"variables"` // Available template variables
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	Description string `gorm:"type:text" json:"description"`
}

// TableName specifies the table name for EmailTemplate
func (EmailTemplate) TableName() string {
	return "email_templates"
}

// EmailLog represents sent email logs
type EmailLog struct {
	BaseModel
	To         string     `gorm:"size:500;not null" json:"to"`
	From       string     `gorm:"size:500" json:"from"`
	Subject    string     `gorm:"size:500;not null" json:"subject"`
	Body       string     `gorm:"type:text" json:"body"`
	TemplateID *uint      `json:"template_id,omitempty"`
	Status     string     `gorm:"size:50;not null" json:"status"` // pending, sent, failed
	SentAt     *time.Time `json:"sent_at,omitempty"`
	Error      string     `gorm:"type:text" json:"error,omitempty"`
	Metadata   string     `gorm:"type:jsonb" json:"metadata,omitempty"`

	// Relationships
	Template *EmailTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
}

// TableName specifies the table name for EmailLog
func (EmailLog) TableName() string {
	return "email_logs"
}
