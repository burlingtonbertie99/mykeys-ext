package service

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
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
	expected := fmt.Sprintf(`Removing "%s/Library/Application Support/KeyTest".
Removing "%s/Library/Logs/KeyTest".
Uninstalled "KeyTest".
`, home, home)
	require.Equal(t, expected, out.String())
}

func TestUninstallSymlink(t *testing.T) {
	var out bytes.Buffer
	var err error
	env, err := NewEnv("KeyTest", build)
	require.NoError(t, err)

	env.linkDir = filepath.Join(os.TempDir())
	err = installSymlink(env)
	require.NoError(t, err)

	err = Uninstall(&out, env)
	require.NoError(t, err)

	home := kenv.MustHomeDir()
	expected := fmt.Sprintf(`Removing "%s/Library/Application Support/KeyTest".
Removing "%s/Library/Logs/KeyTest".
Removed "%s/keystest".
Uninstalled "KeyTest".
`, home, home, env.linkDir)
	require.Equal(t, expected, out.String())
}
