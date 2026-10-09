package admin

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	wireproxy "github.com/bropines/awg-wireproxy"
)

const validateTimeout = 20 * time.Second

// readConfig returns the config file contents. A missing file is not an error
// so that the panel can be used to create the first config.
func readConfig(path string) (text string, exists bool, err error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(b), true, nil
}

// validateConfig parses text exactly like the daemon would on start, without
// touching the real config file. The candidate is written next to the real
// file because relative WGConfig paths are resolved against its directory.
func validateConfig(path, text string) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".wireproxy-validate-*")
	if err != nil {
		return fmt.Errorf("cannot create a temporary file next to the config: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(text); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		_, err := wireproxy.ParseConfig(tmp.Name())
		done <- err
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(validateTimeout):
		return errors.New("validation timed out (is an Endpoint hostname resolvable?)")
	}
}

// writeConfig atomically replaces the config file, keeping the previous
// version at <path>.bak.
func writeConfig(path, text string) error {
	mode := os.FileMode(0o600) // configs hold private keys
	if st, err := os.Stat(path); err == nil {
		mode = st.Mode().Perm()
		if old, err := os.ReadFile(path); err == nil {
			_ = os.WriteFile(path+".bak", old, 0o600)
		}
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".wireproxy-save-*")
	if err == nil {
		tmpName := tmp.Name()
		_, werr := io.WriteString(tmp, text)
		if werr == nil {
			werr = tmp.Sync()
		}
		cerr := tmp.Close()
		if werr == nil && cerr == nil {
			if err = os.Chmod(tmpName, mode); err == nil {
				if err = os.Rename(tmpName, path); err == nil {
					return nil
				}
			}
		}
		os.Remove(tmpName)
	}

	// Fallback for setups where rename is impossible (e.g. a single file
	// bind-mounted into a container, or a read-only directory).
	if err := os.WriteFile(path, []byte(text), mode); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}
	return nil
}
