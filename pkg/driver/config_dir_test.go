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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func tempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("error creating directory %v", err)
	}
	return dir
}

// TestInitConfigDirPreExistingConfig asserts that a symlink is created to the legacy directory if efs-utils.conf is
// found there.
func TestInitConfigDirPreExistingConfig(t *testing.T) {
	dir := tempDir(t)
	defer os.RemoveAll(dir)

	// create legacy dir and a fake pre-existing conf file
	legacyDir := filepath.Join(dir, "legacy")
	if err := os.MkdirAll(legacyDir, os.ModePerm); err != nil {
		t.Fatalf("Unable to create directory: %v", err)
	}
	confFile := filepath.Join(legacyDir, "efs-utils.conf")
	if f, err := os.Create(confFile); err != nil {
		t.Fatalf("Unable to create file: %v", err)
	} else {
		f.Close()
	}

	// create the preferred dir which will go unused
	preferredDir := filepath.Join(dir, "preferred")
	if err := os.MkdirAll(preferredDir, os.ModePerm); err != nil {
		t.Fatalf("Unable to create directory: %v", err)
	}

	etcAmazonEfs := filepath.Join(dir, "symlink")

	// function under test
	if err := InitConfigDir(legacyDir, preferredDir, etcAmazonEfs); err != nil {
		t.Fatalf("InitConfigDir returned an error: %v", err)
	}

	// assert symlink is there
	if _, err := os.Lstat(etcAmazonEfs); err != nil {
		t.Errorf("symlink was not created at %s", etcAmazonEfs)
	}

	symlinkFile := path.Join(etcAmazonEfs, "foo.txt")
	expectedLocation := path.Join(legacyDir, "foo.txt")

	// create some file using the symlink
	if f, err := os.Create(symlinkFile); err != nil {
		t.Errorf("Unable to create file using symlink: %v", err)
	} else {
		f.Close()
	}

	// assert that the file was created in the legacy dir
	if _, err := os.Stat(expectedLocation); err != nil {
		t.Errorf("foo.txt was not created in the legacy directory '%s'", legacyDir)
	}
}

// TestInitConfigDirPreferred asserts that a symlink is created to the preferred directory if efs-utils.conf is
// not found in the legacy location.
func TestInitConfigDirPreferred(t *testing.T) {
	dir := tempDir(t)
	defer os.RemoveAll(dir)

	// create an empty legacy dir
	legacyDir := filepath.Join(dir, "legacy")
	if err := os.MkdirAll(legacyDir, os.ModePerm); err != nil {
		t.Fatalf("Unable to create directory: %v", err)
	}

	// create the preferred dir
	preferredDir := filepath.Join(dir, "preferred")
	if err := os.MkdirAll(preferredDir, os.ModePerm); err != nil {
		t.Fatalf("Unable to create directory: %v", err)
	}

	etcAmazonEfs := filepath.Join(dir, "symlink")

	// function under test
	if err := InitConfigDir(legacyDir, preferredDir, etcAmazonEfs); err != nil {
		t.Fatalf("InitConfigDir returned an error: %v", err)
	}

	// assert symlink is there
	if _, err := os.Lstat(etcAmazonEfs); err != nil {
		t.Errorf("symlink was not created at %s", etcAmazonEfs)
	}

	symlinkFile := path.Join(etcAmazonEfs, "foo.txt")
	expectedLocation := path.Join(preferredDir, "foo.txt")

	// create some file using the symlink
	if f, err := os.Create(symlinkFile); err != nil {
		t.Errorf("Unable to create file using symlink: %v", err)
	} else {
		f.Close()
	}

	// assert that the file was created in the legacy dir
	if _, err := os.Stat(expectedLocation); err != nil {
		t.Errorf("foo.txt was not created in the preferred directory '%s'", legacyDir)
	}
}

// TestInitConfigDirDoNothing asserts that a pre-existing symlink is not altered.
func TestInitConfigDirDoNothing(t *testing.T) {
	dir := tempDir(t)
	defer os.RemoveAll(dir)

	// create an empty legacy dir
	legacyDir := filepath.Join(dir, "legacy")
	if err := os.MkdirAll(legacyDir, os.ModePerm); err != nil {
		t.Fatalf("Unable to create directory: %v", err)
	}

	// create the preferred dir
	preferredDir := filepath.Join(dir, "preferred")
	if err := os.MkdirAll(preferredDir, os.ModePerm); err != nil {
		t.Fatalf("Unable to create directory: %v", err)
	}

	etcAmazonEfs := filepath.Join(dir, "symlink")

	// create a symlink, as if this container has been run previously
	if err := os.Symlink(preferredDir, etcAmazonEfs); err != nil {
		t.Fatalf("unable to create symlink %v", err)
	}

	// function under test
	if err := InitConfigDir(legacyDir, preferredDir, etcAmazonEfs); err != nil {
		t.Fatalf("InitConfigDir returned an error: %v", err)
	}

	// assert symlink is still there
	if _, err := os.Lstat(etcAmazonEfs); err != nil {
		t.Errorf("symlink is missing %s", etcAmazonEfs)
	}

	symlinkFile := path.Join(etcAmazonEfs, "foo.txt")
	expectedLocation := path.Join(preferredDir, "foo.txt")

	// create some file using the symlink
	if f, err := os.Create(symlinkFile); err != nil {
		t.Errorf("Unable to create file using symlink: %v", err)
	} else {
		f.Close()
	}

	// assert that the file was created in the legacy dir
	if _, err := os.Stat(expectedLocation); err != nil {
		t.Errorf("foo.txt was not created in the preferred directory '%s'", legacyDir)
	}
}

// TestInitConfigDirError should error because a directory exists where we ask the function to put a symlink.
func TestInitConfigDirError(t *testing.T) {
	dir := tempDir(t)
	defer os.RemoveAll(dir)

	// create a directory that should not exist
	etcAmazonEfs := filepath.Join(dir, "etc-amazon-efs")
	if err := os.MkdirAll(etcAmazonEfs, os.ModePerm); err != nil {
		t.Fatalf("Unable to create directory: %v", err)
	}

	// function under test
	if err := InitConfigDir("whatever1", "whatever2", etcAmazonEfs); err == nil {
		t.Errorf("InitConfigDir was expected to return an error but did not.")
	}
}
