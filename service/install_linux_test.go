package service

import (
	"bytes"
	"fmt"
	"testing"

	kenv "github.com/burlingtonbertie99/mykeys/env"
	"github.com/stretchr/testify/require"
)

func TestUninstall(t *testing.T) {
	var out bytes.Buffer
	var err error
	env, err := NewEnv("KeyTest", build)
	require.NoError(t, err)
	err = Uninstall(&out, env)
	require.NoError(t, err)

	home := kenv.MustHomeDir()
	expected := fmt.Sprintf(`Removing "%s/.local/share/KeyTest".
Removing "%s/.config/KeyTest".
Removing "%s/.cache/KeyTest".
Uninstalled "KeyTest".
`, home, home, home)
	require.Equal(t, expected, out.String())
}
