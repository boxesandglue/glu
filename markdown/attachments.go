package markdown

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/boxesandglue/boxesandglue/backend/document"
	"github.com/boxesandglue/boxesandglue/frontend"

	"github.com/boxesandglue/glu/internal/errkind"
)

// applyAttachments embeds the files listed in fm.Attachments into the PDF.
// Empty list is a no-op. Paths in AttachmentSpec.File are resolved relative
// to sourceDir; absolute paths are kept as-is. Missing fields fall back to
// sensible defaults (Name → basename, MimeType → application/octet-stream).
func applyAttachments(fe *frontend.Document, sourceDir string, attachments []AttachmentSpec) error {
	for i, spec := range attachments {
		if spec.File == "" {
			return fmt.Errorf("%w: attachments[%d]: file is required", errkind.Usage, i)
		}
		path := resolveSourcePath(sourceDir, spec.File)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("%w: reading attachment %s: %s", errkind.IO, path, err.Error())
		}
		name := spec.Name
		if name == "" {
			name = filepath.Base(spec.File)
		}
		mime := spec.MimeType
		if mime == "" {
			mime = "application/octet-stream"
		}
		fe.Doc.AttachFile(document.Attachment{
			Name:        name,
			Description: spec.Description,
			MimeType:    mime,
			Data:        data,
		})
	}
	return nil
}

// resolveSourcePath joins p with sourceDir unless p is already absolute.
// Empty sourceDir means "use cwd" (consistent with os.ReadFile's default).
func resolveSourcePath(sourceDir, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	if sourceDir == "" {
		return p
	}
	return filepath.Join(sourceDir, p)
}
