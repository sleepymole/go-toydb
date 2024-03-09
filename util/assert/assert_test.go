package assert

import "testing"

func TestTrue(t *testing.T) {
	testCases := []struct {
		name       string
		condition  bool
		msgAndArgs []any
		panicMsg   string
	}{
		{
			name:      "condition is true",
			condition: true,
		},
		{
			name:       "condition is false",
			condition:  false,
			msgAndArgs: []any{"%d != %d", 1, 2},
			panicMsg:   "assertion true failed, 1 != 2",
		},
		{
			name:      "condition is false with no message",
			condition: false,
			panicMsg:  "assertion true failed",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if r != tc.panicMsg {
						t.Errorf("expected panic message %s, but got %s", tc.panicMsg, r)
					}
				}
			}()
			True(tc.condition, tc.msgAndArgs...)
		})
	}
}
