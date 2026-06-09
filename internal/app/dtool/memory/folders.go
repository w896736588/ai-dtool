package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	DefaultFolderName = `fragments`
	TrashFolderName   = `trash`
	foldersMetaFile   = `folders.json`
)

var folderNameRegexp = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type folderMetaFileData struct {
	Folders []Folder `json:"folders"`
}

func NormalizeFolderName(folderName string) string {
	folderName = strings.TrimSpace(folderName)
	if folderName == `` {
		return DefaultFolderName
	}
	return folderName
}

func defaultSystemFolders() []Folder {
	return []Folder{
		{FolderName: DefaultFolderName, Name: `默认文件夹`, System: true, Editable: false},
		{FolderName: TrashFolderName, Name: `回收站`, System: true, Editable: false},
	}
}

func foldersMetaPath(root string) string {
	return filepath.Join(root, foldersMetaFile)
}

func loadFolders(root string) ([]Folder, error) {
	if err := ensureFolderMeta(root); err != nil {
		return nil, err
	}
	body, err := os.ReadFile(foldersMetaPath(root))
	if err != nil {
		return nil, err
	}
	data := folderMetaFileData{}
	if err = json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	folderMap := make(map[string]Folder)
	for _, item := range defaultSystemFolders() {
		folderMap[item.FolderName] = item
	}
	for _, item := range data.Folders {
		item.FolderName = NormalizeFolderName(item.FolderName)
		if item.FolderName == `` {
			continue
		}
		if item.FolderName == DefaultFolderName || item.FolderName == TrashFolderName {
			sys := folderMap[item.FolderName]
			if strings.TrimSpace(item.Name) != `` {
				sys.Name = strings.TrimSpace(item.Name)
			}
			folderMap[item.FolderName] = sys
			continue
		}
		if strings.TrimSpace(item.Name) == `` {
			item.Name = item.FolderName
		}
		item.System = false
		item.Editable = true
		folderMap[item.FolderName] = item
	}
	result := make([]Folder, 0, len(folderMap))
	for _, item := range folderMap {
		result = append(result, item)
	}
	sortFolders(result)
	return result, nil
}

func saveFolders(root string, folders []Folder) error {
	normalized := make([]Folder, 0, len(folders))
	seen := make(map[string]struct{})
	for _, item := range folders {
		item.FolderName = NormalizeFolderName(item.FolderName)
		if item.FolderName == `` {
			continue
		}
		if _, ok := seen[item.FolderName]; ok {
			continue
		}
		seen[item.FolderName] = struct{}{}
		if item.FolderName == DefaultFolderName || item.FolderName == TrashFolderName {
			item.System = true
			item.Editable = false
			if strings.TrimSpace(item.Name) == `` {
				if item.FolderName == DefaultFolderName {
					item.Name = `默认文件夹`
				} else {
					item.Name = `回收站`
				}
			}
		} else {
			item.System = false
			item.Editable = true
			if strings.TrimSpace(item.Name) == `` {
				item.Name = item.FolderName
			}
		}
		normalized = append(normalized, item)
	}
	sortFolders(normalized)
	body, err := json.MarshalIndent(folderMetaFileData{Folders: normalized}, ``, `  `)
	if err != nil {
		return err
	}
	return os.WriteFile(foldersMetaPath(root), body, 0o644)
}

func ensureFolderMeta(root string) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	metaPath := foldersMetaPath(root)
	if _, err := os.Stat(metaPath); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return saveFolders(root, defaultSystemFolders())
}

func ensureFolderDirectories(root string, folders []Folder) error {
	for _, folder := range folders {
		if err := os.MkdirAll(filepath.Join(root, folder.FolderName), 0o755); err != nil {
			return err
		}
	}
	return nil
}

func validateNewFolderName(folderName string) error {
	folderName = strings.TrimSpace(folderName)
	if folderName == `` {
		return fmt.Errorf(`文件夹名称不能为空`)
	}
	if folderName == TrashFolderName {
		return fmt.Errorf(`文件夹名称不能为 trash`)
	}
	if folderName == DefaultFolderName {
		return fmt.Errorf(`文件夹名称不能为 fragments`)
	}
	if !folderNameRegexp.MatchString(folderName) {
		return fmt.Errorf(`文件夹名称仅支持字母、数字、下划线和中划线`)
	}
	return nil
}

func sortFolders(folders []Folder) {
	sort.SliceStable(folders, func(i, j int) bool {
		li := folders[i]
		lj := folders[j]
		ranki := folderSortRank(li.FolderName)
		rankj := folderSortRank(lj.FolderName)
		if ranki != rankj {
			return ranki < rankj
		}
		return strings.ToLower(li.Name) < strings.ToLower(lj.Name)
	})
}

func folderSortRank(folderName string) int {
	switch folderName {
	case DefaultFolderName:
		return 0
	case TrashFolderName:
		return 2
	default:
		return 1
	}
}
