package output

import (
	"encoding/json"
	"io"

	"github.com/thekiran/iad/internal/model"
)

func ExportJSON(w io.Writer, scan model.ScanOutput) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(scan)
}
