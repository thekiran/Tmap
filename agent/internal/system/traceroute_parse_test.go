package system

import (
	"reflect"
	"testing"
)

// TestParseHops checks that the shared hop parser handles both the Windows
// (IP last on the line) and unix (IP first) layouts, and represents timeouts.
func TestParseHops(t *testing.T) {
	windows := `
Tracing route to 8.8.8.8 over a maximum of 15 hops

  1     1 ms     1 ms     1 ms  192.168.1.1
  2     *        *        *     Request timed out.
  3    18 ms    17 ms    18 ms  81.212.0.1
  4    19 ms    20 ms    18 ms  8.8.8.8

Trace complete.`

	unix := `traceroute to 8.8.8.8 (8.8.8.8), 30 hops max, 60 byte packets
 1  192.168.1.1  0.512 ms  0.480 ms  0.470 ms
 2  * * *
 3  81.212.0.1  18.0 ms  17.5 ms  18.2 ms
 4  8.8.8.8  19.0 ms  18.8 ms  19.1 ms`

	want := []string{"192.168.1.1", "*", "81.212.0.1", "8.8.8.8"}

	if got := parseHops(windows); !reflect.DeepEqual(got, want) {
		t.Errorf("windows hops = %v, want %v", got, want)
	}
	if got := parseHops(unix); !reflect.DeepEqual(got, want) {
		t.Errorf("unix hops = %v, want %v", got, want)
	}
}
