package scanner

import (
	_ "embed"
	"encoding/csv"
	"strings"
	"sync"
)

//go:embed oui.csv
var ouiCSVData string

var (
	ouiMap     map[string]string
	ouiMapOnce sync.Once
)

// getOUIMap lazily parses the embedded OUI CSV and returns a prefix→vendor map.
func getOUIMap() map[string]string {
	ouiMapOnce.Do(func() {
		ouiMap = parseOUICSV(ouiCSVData)
	})
	return ouiMap
}

// parseOUICSV parses the IEEE OUI CSV format.
// Expected columns: Registry, Assignment (6-hex OUI), Organization Name, ...
func parseOUICSV(data string) map[string]string {
	m := make(map[string]string, 30000)
	r := csv.NewReader(strings.NewReader(data))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	// Skip header row
	if _, err := r.Read(); err != nil {
		return m
	}

	for {
		record, err := r.Read()
		if err != nil {
			break
		}
		// record[0] = Registry (MA-L, MA-M, MA-S, IAB)
		// record[1] = Assignment (hex string, 6 chars for MA-L)
		// record[2] = Organization Name
		if len(record) < 3 {
			continue
		}
		assignment := strings.ToUpper(strings.TrimSpace(record[1]))
		// Only handle MA-L (6 hex chars = 3-byte OUI)
		if len(assignment) != 6 {
			continue
		}
		// Normalize to "XX:XX:XX"
		key := assignment[0:2] + ":" + assignment[2:4] + ":" + assignment[4:6]
		vendor := strings.TrimSpace(record[2])
		if vendor != "" {
			m[key] = vendor
		}
	}
	return m
}

// LookupVendor returns the IEEE-registered vendor name for a given MAC address.
// Returns an empty string if the OUI is unknown or MAC is invalid.
func LookupVendor(mac string) string {
	if mac == "" {
		return ""
	}
	mac = strings.ToUpper(mac)

	// Support both colon-separated (AA:BB:CC:DD:EE:FF) and
	// hyphen-separated (AA-BB-CC-DD-EE-FF) formats.
	mac = strings.ReplaceAll(mac, "-", ":")

	parts := strings.Split(mac, ":")
	if len(parts) < 3 {
		return ""
	}

	oui := parts[0] + ":" + parts[1] + ":" + parts[2]
	return getOUIMap()[oui]
}

