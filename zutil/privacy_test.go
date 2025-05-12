package zutil

import "testing"

func TestPrivacy(t *testing.T) {
	var args = []struct {
		str    string
		count  int
		result string
	}{
		{"888888888", 2, "888**8888"},
		{"888888888", 4, "88****888"},
		{"888888888", 6, "8******88"},
		{"888888888", 7, "8*******8"},
		{"888888888", 8, "8*******8"},
		{"888888888", 9, "8*******8"},
		{"888888888", 10, "8*******8"},
		{"88", 3, "8*"},
		{"88", 2, "8*"},
		{"88", 1, "8*"},
		{"888", 2, "8*8"},
		{"8", 2, "*"},
		{"", 2, ""},
	}
	t.Run("Privacy", func(t *testing.T) {
		for _, arg := range args {
			r := Privacy(arg.str, arg.count)
			if r != arg.result {
				t.Errorf("Privacy(%s,%d) = %s; want %s", arg.str, arg.count, r, arg.result)
			}
		}
	})
	var argsMust = []struct {
		str    string
		count  int
		result string
	}{
		{"888888888", 2, "888**8888"},
		{"888888888", 4, "88****888"},
		{"888888888", 6, "8******88"},
		{"888888888", 7, "8*******8"},
		{"888888888", 8, "8*******8"},
		{"888888888", 9, "8*******8"},
		{"888888888", 10, "8********8"},
		{"888888888", 15, "8*************8"},
		{"88", 3, "8**"},
		{"88", 2, "8*"},
		{"88", 1, "8*"},
		{"888", 2, "8*8"},
		{"8", 2, "**"},
		{"", 2, "**"},
	}
	t.Run("PrivacyMust", func(t *testing.T) {
		for _, arg := range argsMust {
			r := PrivacyMust(arg.str, arg.count, arg.count)
			if r != arg.result {
				t.Errorf("PrivacyMust(%s,%d,%d) = %s; want %s", arg.str, arg.count, arg.count, r, arg.result)
			}
		}
	})
}
