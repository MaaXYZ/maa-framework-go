//go:build darwin || linux || windows

package maa

func Init(libDir string) error {

	handleLibDir(libDir)

	initFns := []func(libDir string) error{
		initFramework,
		initToolkit,
		initServer,
		initClient,
	}

	for _, initFn := range initFns {
		if err := initFn(libDir); err != nil {
			return err
		}
	}

	return nil
}

func Release() error {
	releaseFns := []func() error{
		unregisterFramework,
		unregisterToolkit,
		unregisterServer,
		unregisterClient,
	}

	for _, releaseFn := range releaseFns {
		if err := releaseFn(); err != nil {
			return err
		}
	}

	return nil
}
