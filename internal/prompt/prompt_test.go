package prompt

import "testing"

func TestCommandRejectsPathTraversal(t *testing.T) {
	if err := Command([]string{"../AGENTS"}); err == nil {
		t.Fatal("expected invalid prompt name error")
	}
}
