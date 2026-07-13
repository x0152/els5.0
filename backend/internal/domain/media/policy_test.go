package media_test

import (
	"strings"
	"testing"

	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/test"
)

func TestImagePolicy_Validate(t *testing.T) {
	cases := []struct {
		name     string
		size     int64
		mime     string
		filename string
		wantErr  bool
		wantExt  string
	}{
		{name: "png_canonical_ext", size: 100, mime: "image/png", filename: "x.png", wantExt: ".png"},
		{name: "png_no_filename_uses_canonical", size: 100, mime: "image/png", filename: "", wantExt: ".png"},
		{name: "jpeg_alt_extension_jpeg", size: 100, mime: "image/jpeg", filename: "x.jpeg", wantExt: ".jpeg"},
		{name: "jpeg_canonical_jpg", size: 100, mime: "image/jpeg", filename: "x.jpg", wantExt: ".jpg"},
		{name: "jpeg_unknown_filename_ext_falls_back", size: 100, mime: "image/jpeg", filename: "x.tiff", wantExt: ".jpg"},
		{name: "webp", size: 100, mime: "image/webp", filename: "x.webp", wantExt: ".webp"},
		{name: "gif", size: 100, mime: "image/gif", filename: "x.gif", wantExt: ".gif"},
		{name: "size_zero", size: 0, mime: "image/png", wantErr: true},
		{name: "size_negative", size: -1, mime: "image/png", wantErr: true},
		{name: "size_over_limit", size: 100 * 1024 * 1024, mime: "image/png", wantErr: true},
		{name: "mime_disallowed_svg", size: 100, mime: "image/svg+xml", wantErr: true},
		{name: "mime_disallowed_pdf", size: 100, mime: "application/pdf", wantErr: true},
		{name: "mime_disallowed_text", size: 100, mime: "text/plain", wantErr: true},
		{name: "mime_disallowed_zip", size: 100, mime: "application/zip", wantErr: true},
		{name: "mime_case_normalized", size: 100, mime: "Image/PNG", filename: "x.png", wantExt: ".png"},
	}

	policy := media.ImagePolicy(0)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ext, err := policy.Validate(tc.size, tc.mime, tc.filename)
			if tc.wantErr {
				test.ErrIs(t, err, shared.ErrValidation)
				return
			}
			test.NoErr(t, err)
			if ext != tc.wantExt {
				t.Fatalf("ext: want %q, got %q", tc.wantExt, ext)
			}
		})
	}
}

func TestDocumentPolicy_Validate(t *testing.T) {
	cases := []struct {
		name     string
		size     int64
		mime     string
		filename string
		wantErr  bool
		wantExt  string
	}{
		{name: "pdf_canonical", size: 100, mime: "application/pdf", filename: "x.pdf", wantExt: ".pdf"},
		{name: "pdf_no_filename_uses_canonical", size: 100, mime: "application/pdf", filename: "", wantExt: ".pdf"},
		{name: "zip_canonical", size: 100, mime: "application/zip", filename: "x.zip", wantExt: ".zip"},
		{name: "docx_alt_extension", size: 100, mime: "application/zip", filename: "report.docx", wantExt: ".docx"},
		{name: "xlsx_alt_extension", size: 100, mime: "application/zip", filename: "data.xlsx", wantExt: ".xlsx"},
		{name: "pptx_alt_extension", size: 100, mime: "application/zip", filename: "deck.pptx", wantExt: ".pptx"},
		{name: "alien_filename_ext_falls_back_to_zip", size: 100, mime: "application/zip", filename: "x.tar", wantExt: ".zip"},
		{name: "size_zero", size: 0, mime: "application/pdf", wantErr: true},
		{name: "size_over_limit", size: 100 * 1024 * 1024, mime: "application/pdf", wantErr: true},
		{name: "image_mime_disallowed", size: 100, mime: "image/png", wantErr: true},
		{name: "exe_mime_disallowed", size: 100, mime: "application/x-msdownload", wantErr: true},
	}

	policy := media.DocumentPolicy(0)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ext, err := policy.Validate(tc.size, tc.mime, tc.filename)
			if tc.wantErr {
				test.ErrIs(t, err, shared.ErrValidation)
				return
			}
			test.NoErr(t, err)
			if ext != tc.wantExt {
				t.Fatalf("ext: want %q, got %q", tc.wantExt, ext)
			}
		})
	}
}

func TestUploadPolicy_AllowedMimes(t *testing.T) {
	cases := []struct {
		name   string
		policy media.UploadPolicy
		want   []string
	}{
		{name: "image", policy: media.ImagePolicy(0), want: []string{"image/png", "image/jpeg", "image/webp", "image/gif"}},
		{name: "document", policy: media.DocumentPolicy(0), want: []string{"application/pdf", "application/zip"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.policy.AllowedMimes()
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestImagePolicy_DefaultMaxSizeApplied(t *testing.T) {
	policy := media.ImagePolicy(0)
	if policy.MaxSize != 5*1024*1024 {
		t.Fatalf("default image max size: want 5MB, got %d", policy.MaxSize)
	}
	policy = media.ImagePolicy(1024)
	if policy.MaxSize != 1024 {
		t.Fatalf("custom image max size: want 1024, got %d", policy.MaxSize)
	}
}

func TestDocumentPolicy_DefaultMaxSizeApplied(t *testing.T) {
	policy := media.DocumentPolicy(0)
	if policy.MaxSize != 25*1024*1024 {
		t.Fatalf("default doc max size: want 25MB, got %d", policy.MaxSize)
	}
}
