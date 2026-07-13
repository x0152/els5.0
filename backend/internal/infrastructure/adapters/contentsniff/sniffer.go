package contentsniff

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/els/backend/internal/domain/shared/ports"
)

const HeadSize = 512

var _ ports.ContentSniffer = (*Sniffer)(nil)

type Sniffer struct{}

func New() *Sniffer { return &Sniffer{} }

func (s *Sniffer) Sniff(r io.Reader) (io.Reader, string, error) {
	if r == nil {
		return nil, "", errors.New("contentsniff: nil reader")
	}
	br := bufio.NewReaderSize(r, HeadSize)
	head, err := br.Peek(HeadSize)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, bufio.ErrBufferFull) {
		return nil, "", fmt.Errorf("contentsniff: peek head: %w", err)
	}
	mime := http.DetectContentType(head)
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = mime[:i]
	}
	return br, strings.ToLower(strings.TrimSpace(mime)), nil
}
