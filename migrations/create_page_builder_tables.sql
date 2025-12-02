-- Create pages table
CREATE TABLE IF NOT EXISTS pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL,
    template VARCHAR(100) DEFAULT 'default',
    status VARCHAR(20) DEFAULT 'draft',
    seo_title VARCHAR(500),
    seo_description TEXT,
    og_image VARCHAR(500),
    author_id UUID NOT NULL,
    published_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT idx_pages_slug UNIQUE (slug)
);

CREATE INDEX idx_pages_deleted_at ON pages(deleted_at);
CREATE INDEX idx_pages_author_id ON pages(author_id);
CREATE INDEX idx_pages_status ON pages(status);

-- Create blocks table
CREATE TABLE IF NOT EXISTS blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(100) NOT NULL,
    category VARCHAR(100),
    schema JSONB NOT NULL DEFAULT '{}',
    preview_template TEXT,
    is_global BOOLEAN DEFAULT false
);

CREATE INDEX idx_blocks_deleted_at ON blocks(deleted_at);
CREATE INDEX idx_blocks_type ON blocks(type);
CREATE INDEX idx_blocks_category ON blocks(category);

-- Create page_blocks table (junction table)
CREATE TABLE IF NOT EXISTS page_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    page_id UUID NOT NULL,
    block_id UUID NOT NULL,
    parent_block_id UUID,
    "order" INTEGER DEFAULT 0,
    config JSONB DEFAULT '{}',
    language VARCHAR(10) DEFAULT 'en',
    CONSTRAINT fk_page_blocks_page FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE,
    CONSTRAINT fk_page_blocks_block FOREIGN KEY (block_id) REFERENCES blocks(id) ON DELETE CASCADE,
    CONSTRAINT fk_page_blocks_parent FOREIGN KEY (parent_block_id) REFERENCES page_blocks(id) ON DELETE CASCADE
);

CREATE INDEX idx_page_blocks_page_id ON page_blocks(page_id);
CREATE INDEX idx_page_blocks_parent_id ON page_blocks(parent_block_id);
CREATE INDEX idx_page_blocks_order ON page_blocks("order");

-- Create page_versions table
CREATE TABLE IF NOT EXISTS page_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    page_id UUID NOT NULL,
    version_number INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL,
    blocks_snapshot JSONB NOT NULL, -- Full snapshot of blocks config
    created_by UUID NOT NULL,
    change_description TEXT,
    CONSTRAINT fk_page_versions_page FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE
);

CREATE INDEX idx_page_versions_page_id ON page_versions(page_id);
CREATE INDEX idx_page_versions_created_by ON page_versions(created_by);

-- Create theme_settings table
CREATE TABLE IF NOT EXISTS theme_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(200) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN DEFAULT false
);

CREATE INDEX idx_theme_settings_deleted_at ON theme_settings(deleted_at);
