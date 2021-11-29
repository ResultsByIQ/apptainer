// Copyright (c) 2018-2021, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package sources

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apptainer/apptainer/internal/pkg/util/fs"
	"github.com/apptainer/apptainer/pkg/sylog"
)

const (
	// Contents of /.apptainer.d/actions/exec
	execFileContent = `#!/bin/sh

for script in /.apptainer.d/env/*.sh; do
    if [ -f "$script" ]; then
        . "$script"
    fi
done

exec "$@"
`
	// Contents of /.apptainer.d/actions/run
	runFileContent = `#!/bin/sh

for script in /.apptainer.d/env/*.sh; do
    if [ -f "$script" ]; then
        . "$script"
    fi
done

if test -n "${APPTAINER_APPNAME:-}"; then

    if test -x "/scif/apps/${APPTAINER_APPNAME:-}/scif/runscript"; then
        exec "/scif/apps/${APPTAINER_APPNAME:-}/scif/runscript" "$@"
    else
        echo "No Apptainer runscript for contained app: ${APPTAINER_APPNAME:-}"
        exit 1
    fi

elif test -x "/.apptainer.d/runscript"; then
    exec "/.apptainer.d/runscript" "$@"
else
    echo "No Apptainer runscript found, executing /bin/sh"
    exec /bin/sh "$@"
fi
`
	// Contents of /.apptainer.d/actions/shell
	shellFileContent = `#!/bin/sh

for script in /.apptainer.d/env/*.sh; do
    if [ -f "$script" ]; then
        . "$script"
    fi
done

if test -n "$APPTAINER_SHELL" -a -x "$APPTAINER_SHELL"; then
    exec $APPTAINER_SHELL "$@"

    echo "ERROR: Failed running shell as defined by '\$APPTAINER_SHELL'" 1>&2
    exit 1

elif test -x /bin/bash; then
    SHELL=/bin/bash
    PS1="Apptainer $APPTAINER_NAME:\\w> "
    export SHELL PS1
    exec /bin/bash --norc "$@"
elif test -x /bin/sh; then
    SHELL=/bin/sh
    export SHELL
    exec /bin/sh "$@"
else
    echo "ERROR: /bin/sh does not exist in container" 1>&2
fi
exit 1
`
	// Contents of /.apptainer.d/actions/start
	startFileContent = `#!/bin/sh

# if we are here start notify PID 1 to continue
# DON'T REMOVE
kill -CONT 1

for script in /.apptainer.d/env/*.sh; do
    if [ -f "$script" ]; then
        . "$script"
    fi
done

if test -x "/.apptainer.d/startscript"; then
    exec "/.apptainer.d/startscript"
fi
`
	// Contents of /.apptainer.d/actions/test
	testFileContent = `#!/bin/sh

for script in /.apptainer.d/env/*.sh; do
    if [ -f "$script" ]; then
        . "$script"
    fi
done


if test -n "${APPTAINER_APPNAME:-}"; then

    if test -x "/scif/apps/${APPTAINER_APPNAME:-}/scif/test"; then
        exec "/scif/apps/${APPTAINER_APPNAME:-}/scif/test" "$@"
    else
        echo "No tests for contained app: ${APPTAINER_APPNAME:-}"
        exit 1
    fi
elif test -x "/.apptainer.d/test"; then
    exec "/.apptainer.d/test" "$@"
else
    echo "No test found in container, executing /bin/sh -c true"
    exec /bin/sh -c true
fi
`
	// Contents of /.apptainer.d/env/01-base.sh
	baseShFileContent = `#!/bin/sh
# 
# Copyright (c) 2017, ApptainerWare, LLC. All rights reserved.
# Copyright (c) 2015-2017, Gregory M. Kurtzer. All rights reserved.
# 
# Copyright (c) 2016-2017, The Regents of the University of California,
# through Lawrence Berkeley National Laboratory (subject to receipt of any
# required approvals from the U.S. Dept. of Energy).  All rights reserved.
# 
# This software is licensed under a customized 3-clause BSD license.  Please
# consult LICENSE.md file distributed with the sources of this project regarding
# your rights to use or distribute this software.
# 
# NOTICE.  This Software was developed under funding from the U.S. Department of
# Energy and the U.S. Government consequently retains certain rights. As such,
# the U.S. Government has been granted for itself and others acting on its
# behalf a paid-up, nonexclusive, irrevocable, worldwide license in the Software
# to reproduce, distribute copies to the public, prepare derivative works, and
# perform publicly and display publicly, and to permit other to do so.
# 
# 


`
	// Contents of /.apptainer.d/env/90-environment.sh and /.apptainer.d/env/91-environment.sh
	environmentShFileContent = `#!/bin/sh
# Custom environment shell code should follow

`
	// Contents of /.apptainer.d/env/95-apps.sh
	appsShFileContent = `#!/bin/sh
#
# Copyright (c) 2017, ApptainerWare, LLC. All rights reserved.
#
# See the COPYRIGHT.md file at the top-level directory of this distribution and at
# https://github.com/apptainer/apptainer/blob/master/COPYRIGHT.md.
#
# This file is part of the Apptainer Linux container project. It is subject to the license
# terms in the LICENSE.md file found in the top-level directory of this distribution and
# at https://github.com/apptainer/apptainer/blob/master/LICENSE.md. No part
# of Apptainer, including this file, may be copied, modified, propagated, or distributed
# except according to the terms contained in the LICENSE.md file.


if test -n "${APPTAINER_APPNAME:-}"; then

    # The active app should be exported
    export APPTAINER_APPNAME

    if test -d "/scif/apps/${APPTAINER_APPNAME:-}/"; then
        SCIF_APPS="/scif/apps"
        SCIF_APPROOT="/scif/apps/${APPTAINER_APPNAME:-}"
        export SCIF_APPROOT SCIF_APPS
        PATH="/scif/apps/${APPTAINER_APPNAME:-}:$PATH"

        # Automatically add application bin to path
        if test -d "/scif/apps/${APPTAINER_APPNAME:-}/bin"; then
            PATH="/scif/apps/${APPTAINER_APPNAME:-}/bin:$PATH"
        fi

        # Automatically add application lib to LD_LIBRARY_PATH
        if test -d "/scif/apps/${APPTAINER_APPNAME:-}/lib"; then
            LD_LIBRARY_PATH="/scif/apps/${APPTAINER_APPNAME:-}/lib:$LD_LIBRARY_PATH"
            export LD_LIBRARY_PATH
        fi

        # Automatically source environment
        if [ -f "/scif/apps/${APPTAINER_APPNAME:-}/scif/env/01-base.sh" ]; then
            . "/scif/apps/${APPTAINER_APPNAME:-}/scif/env/01-base.sh"
        fi
        if [ -f "/scif/apps/${APPTAINER_APPNAME:-}/scif/env/90-environment.sh" ]; then
            . "/scif/apps/${APPTAINER_APPNAME:-}/scif/env/90-environment.sh"
        fi

        export PATH
    else
        echo "Could not locate the container application: ${APPTAINER_APPNAME}"
        exit 1
    fi
fi

`
	// Contents of /.apptainer.d/env/99-base.sh
	base99ShFileContent = `#!/bin/sh
# 
# Copyright (c) 2017, ApptainerWare, LLC. All rights reserved.
# Copyright (c) 2015-2017, Gregory M. Kurtzer. All rights reserved.
# 
# Copyright (c) 2016-2017, The Regents of the University of California,
# through Lawrence Berkeley National Laboratory (subject to receipt of any
# required approvals from the U.S. Dept. of Energy).  All rights reserved.
# 
# This software is licensed under a customized 3-clause BSD license.  Please
# consult LICENSE.md file distributed with the sources of this project regarding
# your rights to use or distribute this software.
# 
# NOTICE.  This Software was developed under funding from the U.S. Department of
# Energy and the U.S. Government consequently retains certain rights. As such,
# the U.S. Government has been granted for itself and others acting on its
# behalf a paid-up, nonexclusive, irrevocable, worldwide license in the Software
# to reproduce, distribute copies to the public, prepare derivative works, and
# perform publicly and display publicly, and to permit other to do so.
# 
# 


if [ -z "$LD_LIBRARY_PATH" ]; then
    LD_LIBRARY_PATH="/.apptainer.d/libs"
else
    LD_LIBRARY_PATH="$LD_LIBRARY_PATH:/.apptainer.d/libs"
fi

PS1="Apptainer> "
export LD_LIBRARY_PATH PS1
`

	// Contents of /.apptainer.d/env/99-runtimevars.sh
	base99runtimevarsShFileContent = `#!/bin/sh
# Copyright (c) 2017-2019, Sylabs, Inc. All rights reserved.
#
# This software is licensed under a customized 3-clause BSD license.  Please
# consult LICENSE.md file distributed with the sources of this project regarding
# your rights to use or distribute this software.
#
#

if [ -n "${SING_USER_DEFINED_PREPEND_PATH:-}" ]; then
	PATH="${SING_USER_DEFINED_PREPEND_PATH}:${PATH}"
fi

if [ -n "${SING_USER_DEFINED_APPEND_PATH:-}" ]; then
	PATH="${PATH}:${SING_USER_DEFINED_APPEND_PATH}"
fi

if [ -n "${SING_USER_DEFINED_PATH:-}" ]; then
	PATH="${SING_USER_DEFINED_PATH}"
fi

unset SING_USER_DEFINED_PREPEND_PATH \
	  SING_USER_DEFINED_APPEND_PATH \
	  SING_USER_DEFINED_PATH

export PATH
`

	// Contents of /.apptainer.d/runscript
	runscriptFileContent = `#!/bin/sh

echo "There is no runscript defined for this container\n";
`
	// Contents of /.apptainer.d/startscript
	startscriptFileContent = `#!/bin/sh
`
)

func makeDirs(rootPath string) error {
	if err := os.MkdirAll(filepath.Join(rootPath, ".apptainer.d", "libs"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(rootPath, ".apptainer.d", "actions"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(rootPath, ".apptainer.d", "env"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(rootPath, "dev"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(rootPath, "proc"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(rootPath, "root"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(rootPath, "var", "tmp"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(rootPath, "tmp"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(rootPath, "etc"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(rootPath, "sys"), 0o755); err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(rootPath, "home"), 0o755)
}

func makeSymlinks(rootPath string) error {
	if _, err := os.Stat(filepath.Join(rootPath, "apptainer")); err != nil {
		if err = os.Symlink(".apptainer.d/runscript", filepath.Join(rootPath, "apptainer")); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(rootPath, ".run")); err != nil {
		if err = os.Symlink(".apptainer.d/actions/run", filepath.Join(rootPath, ".run")); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(rootPath, ".exec")); err != nil {
		if err = os.Symlink(".apptainer.d/actions/exec", filepath.Join(rootPath, ".exec")); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(rootPath, ".test")); err != nil {
		if err = os.Symlink(".apptainer.d/actions/test", filepath.Join(rootPath, ".test")); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(rootPath, ".shell")); err != nil {
		if err = os.Symlink(".apptainer.d/actions/shell", filepath.Join(rootPath, ".shell")); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(rootPath, "environment")); err != nil {
		if err = os.Symlink(".apptainer.d/env/90-environment.sh", filepath.Join(rootPath, "environment")); err != nil {
			return err
		}
	}
	return nil
}

func makeFile(name string, perm os.FileMode, s string) (err error) {
	// #4532 - If the file already exists ensure it has requested permissions
	// as OpenFile won't set on an existing file and some docker
	// containers have hosts or resolv.conf without write perm.
	if fs.IsFile(name) {
		if err = os.Chmod(name, perm); err != nil {
			return
		}
	}
	// Create the file if it's not in the container, or truncate and write s
	// into it otherwise.
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.WriteString(s)
	return
}

func makeFiles(rootPath string) error {
	if err := makeFile(filepath.Join(rootPath, "etc", "hosts"), 0o644, ""); err != nil {
		return err
	}
	if err := makeFile(filepath.Join(rootPath, "etc", "resolv.conf"), 0o644, ""); err != nil {
		return err
	}
	if err := makeFile(filepath.Join(rootPath, ".apptainer.d", "actions", "exec"), 0o755, execFileContent); err != nil {
		return err
	}
	if err := makeFile(filepath.Join(rootPath, ".apptainer.d", "actions", "run"), 0o755, runFileContent); err != nil {
		return err
	}
	if err := makeFile(filepath.Join(rootPath, ".apptainer.d", "actions", "shell"), 0o755, shellFileContent); err != nil {
		return err
	}
	if err := makeFile(filepath.Join(rootPath, ".apptainer.d", "actions", "start"), 0o755, startFileContent); err != nil {
		return err
	}
	if err := makeFile(filepath.Join(rootPath, ".apptainer.d", "actions", "test"), 0o755, testFileContent); err != nil {
		return err
	}
	if err := makeFile(filepath.Join(rootPath, ".apptainer.d", "env", "01-base.sh"), 0o755, baseShFileContent); err != nil {
		return err
	}
	if err := makeFile(filepath.Join(rootPath, ".apptainer.d", "env", "90-environment.sh"), 0o755, environmentShFileContent); err != nil {
		return err
	}
	if err := makeFile(filepath.Join(rootPath, ".apptainer.d", "env", "95-apps.sh"), 0o755, appsShFileContent); err != nil {
		return err
	}
	if err := makeFile(filepath.Join(rootPath, ".apptainer.d", "env", "99-base.sh"), 0o755, base99ShFileContent); err != nil {
		return err
	}
	if err := makeFile(filepath.Join(rootPath, ".apptainer.d", "env", "99-runtimevars.sh"), 0o755, base99runtimevarsShFileContent); err != nil {
		return err
	}
	if err := makeFile(filepath.Join(rootPath, ".apptainer.d", "runscript"), 0o755, runscriptFileContent); err != nil {
		return err
	}
	return makeFile(filepath.Join(rootPath, ".apptainer.d", "startscript"), 0o755, startscriptFileContent)
}

func makeBaseEnv(rootPath string) (err error) {
	var info os.FileInfo

	// Ensure we can write into the root of rootPath
	if info, err = os.Stat(rootPath); err != nil {
		err = fmt.Errorf("build: failed to stat rootPath: %v", err)
		return err
	}
	if info.Mode()&0o200 == 0 {
		sylog.Infof("Adding owner write permission to build path: %s\n", rootPath)
		if err = os.Chmod(rootPath, info.Mode()|0o200); err != nil {
			err = fmt.Errorf("build: failed to make rootPath writable: %v", err)
			return err
		}
	}

	if err = makeDirs(rootPath); err != nil {
		err = fmt.Errorf("build: failed to make environment dirs: %v", err)
		return err
	}
	if err = makeSymlinks(rootPath); err != nil {
		err = fmt.Errorf("build: failed to make environment symlinks: %v", err)
		return err
	}
	if err = makeFiles(rootPath); err != nil {
		err = fmt.Errorf("build: failed to make environment files: %v", err)
		return err
	}

	return err
}
