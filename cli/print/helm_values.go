package print

import (
	"io"
)

func HelmValues(w io.Writer, values string) error {
	_, err := io.WriteString(w, values)
	return err
}
