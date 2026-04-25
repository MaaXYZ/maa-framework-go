//go:build (darwin || linux || windows) && (amd64 || arm64)

package native

import (
	"errors"
	"fmt"
)

type Library struct {
	name    string
	init    func(string) error
	release func() error
}

var (
	libraries = []Library{
		{name: maaFrameworkName, init: initFramework, release: unregisterFramework},
		{name: maaToolkitName, init: initToolkit, release: unregisterToolkit},
		{name: maaAgentServerName, init: initAgentServer, release: unregisterAgentServer},
		{name: maaAgentClientName, init: initAgentClient, release: unregisterAgentClient},
	}
	loadedLibs []Library
)

func Init(libDir string) error {

	err := handleLibDir(libDir)
	if err != nil {
		return err
	}

	for _, lib := range libraries {
		if err := lib.init(libDir); err != nil {
			releaseErr := Release()

			if releaseErr != nil {
				return fmt.Errorf("%w; error while releasing already loaded libraries: %w", err, releaseErr)
			}
			return err
		}
		loadedLibs = append(loadedLibs, lib)
	}

	return nil
}

func Release() error {

	var (
		errs   []error
		failed []Library
	)

	for i := len(loadedLibs) - 1; i >= 0; i-- {
		lib := loadedLibs[i]
		if err := lib.release(); err != nil {
			errs = append(errs, err)
			failed = append(failed, lib)
		}
	}

	for i, j := 0, len(failed)-1; i < j; i, j = i+1, j-1 {
		failed[i], failed[j] = failed[j], failed[i]
	}

	loadedLibs = failed

	if len(errs) > 0 {
		return fmt.Errorf("failed to release libraries: %w", errors.Join(errs...))
	}

	return nil
}
