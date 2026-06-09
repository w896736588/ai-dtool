package memory

import "time"

// Fragment 表示文件型知识片段。
type Fragment struct {
	ID         string
	Title      string
	Content    string
	FolderName string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	IsDeleted  bool
	FilePath   string
}

// Folder 表示知识片段文件夹元数据。
type Folder struct {
	FolderName string `json:"folder_name" yaml:"folder_name"`
	Name       string `json:"name" yaml:"name"`
	System     bool   `json:"system" yaml:"system"`
	Editable   bool   `json:"editable" yaml:"editable"`
}

// FrontMatter 表示 Markdown 文件头部元数据。
type FrontMatter struct {
	Title      string `yaml:"title"`
	FolderName string `yaml:"folder_name"`
	CreatedAt  string `yaml:"created_at"`
	UpdatedAt  string `yaml:"updated_at"`
}

// LegacyFragment 表示旧 sqlite 里的知识片段记录。
type LegacyFragment struct {
	ID         int
	Title      string
	Content    string
	IsDeleted  bool
	CreateTime int64
	UpdateTime int64
}

// MigrationFile 表示一次迁移输出的文件。
type MigrationFile struct {
	OldID    int
	NewID    string
	FilePath string
	Deleted  bool
}

// MigrationReport 汇总迁移结果。
type MigrationReport struct {
	ActiveCount int
	TrashCount  int
	Files       []MigrationFile
}
