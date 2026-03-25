package skills

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// RegistryFromFS builds a Registry by reading all files from the given
// sub-directory of an fs.FS (typically an embed.FS). Each file becomes a
// StaticSkill whose name is the filename without extension.
func RegistryFromFS(fsys fs.FS, dir string) (*Registry, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("read embedded dir %q: %w", dir, err)
	}

	var skillList []Skill
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		filename := e.Name()
		fullPath := path.Join(dir, filename)

		data, err := fs.ReadFile(fsys, fullPath)
		if err != nil {
			return nil, fmt.Errorf("read embedded file %q: %w", fullPath, err)
		}

		name := strings.TrimSuffix(filename, path.Ext(filename))
		description := strings.ReplaceAll(name, "-", " ")

		skillList = append(skillList, NewStaticSkill(name, description, string(data)))
	}

	return NewRegistry(skillList...), nil
}
