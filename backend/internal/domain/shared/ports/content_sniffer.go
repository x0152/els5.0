package ports

import "io"

type ContentSniffer interface {
	Sniff(r io.Reader) (io.Reader, string, error)
}
