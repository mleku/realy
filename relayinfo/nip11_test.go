package relayinfo

import "testing"

func TestAddSupportedNIP(t *testing.T) {
	info := NewInfo(nil)
	info.AddNIPs(12, 12, 13, 1, 12, 44, 2, 13, 2, 13, 0, 17, 19, 1, 18)
	for i, v := range []int{0, 1, 2, 12, 13, 17, 18, 19, 44} {
		if !info.HasNIP(v) {
			t.Errorf("expected info.nips[%d] to equal %v, got %v",
				i, v, info.Nips)
			return
		}
	}
}
