package orders

import "testing"

func TestValidStatuses_KnownStatuses(t *testing.T) {
	known := []string{"pending", "paid", "processing", "shipped", "completed", "cancelled", "refunded"}
	for _, s := range known {
		if !validStatuses[s] {
			t.Errorf("validStatuses[%q] should be true", s)
		}
	}
}

func TestValidStatuses_UnknownStatuses(t *testing.T) {
	unknown := []string{"", "PENDING", "Paid", "unknown", "draft", "archived"}
	for _, s := range unknown {
		if validStatuses[s] {
			t.Errorf("validStatuses[%q] should be false", s)
		}
	}
}
