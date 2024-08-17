package fail

import "testing"

func TestApp(t *testing.T) {
	t.Error("ooh")
}
