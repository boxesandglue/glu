package markdown

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/boxesandglue/boxesandglue/frontend"

	"github.com/boxesandglue/glu/internal/errkind"
)

func TestResolveSourcePath(t *testing.T) {
	tests := []struct {
		dir, p, want string
	}{
		{"/abs/dir", "rel.xml", "/abs/dir/rel.xml"},
		{"/abs/dir", "/other/abs.xml", "/other/abs.xml"},
		{"", "rel.xml", "rel.xml"},
		{"relbase", "rel.xml", "relbase/rel.xml"},
	}
	for _, tt := range tests {
		if got := resolveSourcePath(tt.dir, tt.p); got != tt.want {
			t.Errorf("resolveSourcePath(%q,%q) = %q; want %q", tt.dir, tt.p, got, tt.want)
		}
	}
}

func newTestFrontend(t *testing.T) *frontend.Document {
	t.Helper()
	fe, err := frontend.New(filepath.Join(t.TempDir(), "out.pdf"))
	if err != nil {
		t.Fatalf("frontend.New: %v", err)
	}
	return fe
}

func TestApplyAttachments(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"a.txt":     "alpha",
		"data.json": `{"k":1}`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	fe := newTestFrontend(t)
	specs := []AttachmentSpec{
		{File: "a.txt", Description: "first"},
		{File: "data.json", Name: "payload.json", MimeType: "application/json"},
	}
	if err := applyAttachments(fe, dir, specs); err != nil {
		t.Fatalf("applyAttachments: %v", err)
	}
	if len(fe.Doc.Attachments) != 2 {
		t.Fatalf("expected 2 attachments, got %d", len(fe.Doc.Attachments))
	}
	a0 := fe.Doc.Attachments[0]
	if a0.Name != "a.txt" || a0.MimeType != "application/octet-stream" || a0.Description != "first" {
		t.Errorf("attachment[0] defaults wrong: %+v", a0)
	}
	a1 := fe.Doc.Attachments[1]
	if a1.Name != "payload.json" || a1.MimeType != "application/json" {
		t.Errorf("attachment[1] override wrong: %+v", a1)
	}
}

func TestApplyAttachments_MissingFile(t *testing.T) {
	fe := newTestFrontend(t)
	err := applyAttachments(fe, t.TempDir(), []AttachmentSpec{{File: "doesnotexist.bin"}})
	if !errors.Is(err, errkind.IO) {
		t.Errorf("expected errkind.IO, got %v", err)
	}
}

func TestApplyAttachments_EmptyFileField(t *testing.T) {
	fe := newTestFrontend(t)
	err := applyAttachments(fe, "", []AttachmentSpec{{Name: "x.txt"}})
	if !errors.Is(err, errkind.Usage) {
		t.Errorf("expected errkind.Usage, got %v", err)
	}
}

func TestApplyAttachments_EmptyList(t *testing.T) {
	fe := newTestFrontend(t)
	if err := applyAttachments(fe, "", nil); err != nil {
		t.Errorf("empty list should be no-op: %v", err)
	}
	if len(fe.Doc.Attachments) != 0 {
		t.Errorf("expected no attachments, got %d", len(fe.Doc.Attachments))
	}
}
