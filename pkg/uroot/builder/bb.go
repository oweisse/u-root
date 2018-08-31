// Copyright 2015-2017 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/u-root/u-root/pkg/cpio"
	"github.com/u-root/u-root/pkg/uroot/builder/bb"
	"github.com/u-root/u-root/pkg/uroot/initramfs"
)

// Commands to skip building in bb mode.
var skip = map[string]struct{}{
	"bb": {},
}

// BBBuilder is an implementation of Builder that compiles many Go commands
// into one busybox-style binary.
//
// BBBuilder will also include symlinks for each command to the busybox binary.
//
// The compiled binary uses argv[0] to decide which Go command to run.
type BBBuilder struct{}

func (BBBuilder) DefaultBinaryDir() string {
	return "bbin"
}

// Build is an implementation of Builder.Build for a busybox-like initramfs.
//
// Build rewrites the source files of the packages given to create one
// busybox-like binary containing all commands in opts.Packages.
//
// The compiled binary uses argv[0] to decide which Go command to run.
//
// See bb/README.md for a detailed explanation of busybox mode.
func (BBBuilder) Build(af initramfs.Files, opts Opts) error {
	// Build the busybox binary.
	bbPath := filepath.Join(opts.TempDir, "bb")
	if err := bb.BuildBusybox(opts.Env, opts.Packages, bbPath); err != nil {
		return err
	}

	if len(opts.BinaryDir) == 0 {
		return fmt.Errorf("must specify binary directory")
	}

	if err := af.AddFile(bbPath, path.Join(opts.BinaryDir, "bb")); err != nil {
		return err
	}

	// Add symlinks for included commands to initramfs.
	for _, pkg := range opts.Packages {
		if _, ok := skip[path.Base(pkg)]; ok {
			continue
		}

		// Add a symlink /bbin/{cmd} -> /bbin/bb to our initramfs.
		if err := af.AddRecord(cpio.Symlink(filepath.Join(opts.BinaryDir, path.Base(pkg)), "bb")); err != nil {
			return err
		}
	}
	return nil
}
