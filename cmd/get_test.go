package cmd_test

import (
	"strings"
	"testing"

	"github.com/ktr0731/dept/cmd"
	"github.com/stretchr/testify/require"
)

func TestGetRun(t *testing.T) {
	t.Run("Run returns code 1 because no arguments passed", func(t *testing.T) {
		testUI := newTestUI()
		cmd, err := cmd.Get(testUI, nil)()
		require.NoError(t, err)

		code := cmd.Run(nil)
		if code != 1 {
			t.Errorf("Run must be 1 because command need to show help message")
		}
		if out := testUI.Writer().String(); !strings.HasPrefix(out, "Usage: ") {
			t.Errorf("Run must write help message to Writer, but actual '%s'", out)
		}
	})

	// code := cmd.Run([]string{"github.com/ktr0731/go-modules-test"})
	// assert.Equal(t, 0, code)
}
