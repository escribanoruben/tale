package password

import "testing"

func TestEntropyBitsAdjacentConstraint(t *testing.T) {
	unrestricted := EntropyBits(20, 20, Options{PreventAdjacentRepeats: false})
	restricted := EntropyBits(20, 20, Options{PreventAdjacentRepeats: true})
	if unrestricted <= restricted {
		t.Fatalf("unrestricted %.2f should exceed restricted %.2f", unrestricted, restricted)
	}
	if restricted < 84 || restricted > 86 {
		t.Fatalf("restricted entropy %.2f outside expected ~85 bits", restricted)
	}
}
