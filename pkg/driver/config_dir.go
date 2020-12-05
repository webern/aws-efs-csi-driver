/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"fmt"
	"k8s.io/klog"
	"os"
	"path"
)

// InitConfigDir decides which of two directories will be used to store driver config files. It creates a symlink to
// the chosen location, and returns the path of that symlink. legacyDir is the path to a config directory where previous
// versions of this driver may have written config files. In previous versions of this driver, a host path that was not
// writeable on Bottlerocket hosts was being used, so we introduce preferredDir so that, going forward, we can use a new
// location on the host. etcAmazonEfs is the path where the symlink will be written. In practice, this will always be
// /etc/amazon/efs, but we take it as an input so the function can be tested.
//
// Examples:
// On a host that has EFS mounts created by an earlier version of this driver, InitConfigDir will detect a conf file in
// legacyDir and write a symlink to that directory.
//
// On a host that does not have pre-existing legacyDir EFS mount configs, InitConfigDir will detect no files in
// legacyDir and will instead point the symlink to preferredDir.
//
// This allows us to gracefully change the location that we store config files on the host, without disrupting pre-
// existing mounts.
func InitConfigDir(legacyDir, preferredDir, etcAmazonEfs string) error {

	// if there is already a symlink in place, we have nothing to do
	if info, err := os.Lstat(etcAmazonEfs); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			klog.Infof("Symlink exists at '%s', no need to create one", etcAmazonEfs)
			return nil
		} else {
			return fmt.Errorf("something already exists at '%s' and it is not a symlink", etcAmazonEfs)
		}
	}

	// check if a conf file exists in the legacy directory and symlink to the directory if so
	existingConfFile := path.Join(legacyDir, "efs-utils.conf")
	if _, err := os.Stat(existingConfFile); err == nil {
		if err = os.Symlink(legacyDir, etcAmazonEfs); err != nil {
			return fmt.Errorf(
				"unable to create symlink from '%s' to '%s': %s",
				etcAmazonEfs,
				legacyDir,
				err.Error())
		}
		klog.Infof("Pre-existing config files are being used from '%s'", legacyDir)
		return nil
	}

	klog.Infof("Creating symlink from '%s' to '%s'", etcAmazonEfs, preferredDir)

	// make sure the config directory exists
	if target, err := os.Stat(preferredDir); err != nil {
		return fmt.Errorf("config directory '%s' does not exist: '%s'", preferredDir, err.Error())
	} else if !target.IsDir() {
		return fmt.Errorf("config directory '%s' is not a directory", preferredDir)
	}

	// create a symlink to the config directory
	if err := os.Symlink(preferredDir, etcAmazonEfs); err != nil {
		return fmt.Errorf(
			"unable to create symlink from '%s' to '%s': %s",
			etcAmazonEfs,
			preferredDir,
			err.Error())
	}

	return nil
}
