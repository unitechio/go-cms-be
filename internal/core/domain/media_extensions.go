package domain

// Add to existing Media struct in internal/core/domain/media.go

// Add these fields to Media struct:
// FileHash    string                 `gorm:"size:64;index" json:"file_hash,omitempty"`
// Variants    map[string]interface{} `gorm:"type:jsonb" json:"variants,omitempty"`

// Migration SQL:
/*
ALTER TABLE media ADD COLUMN file_hash VARCHAR(64);
ALTER TABLE media ADD COLUMN variants JSONB;
CREATE INDEX idx_media_file_hash ON media(file_hash);
*/
