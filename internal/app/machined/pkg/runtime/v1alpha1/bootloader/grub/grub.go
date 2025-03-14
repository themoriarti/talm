// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package grub provides the interface to the GRUB bootloader: config management, installation, etc.
package grub

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/cozystack/talm/internal/pkg/partition"
	"github.com/siderolabs/talos/pkg/machinery/constants"
	"github.com/siderolabs/talos/pkg/machinery/version"
)

// Config represents a grub configuration file (grub.cfg).
type Config struct {
	Default        BootLabel
	Fallback       BootLabel
	Entries        map[BootLabel]MenuEntry
	AddResetOption bool
}

// MenuEntry represents a grub menu entry in the grub config file.
type MenuEntry struct {
	Name    string
	Linux   string
	Cmdline string
	Initrd  string
}

func (e bootloaderNotInstalledError) Error() string {
	return bootloaderNotInstalled
}

// NewConfig creates a new grub configuration (nothing is written to disk).
func NewConfig() *Config {
	return &Config{
		Default:        BootA,
		Entries:        map[BootLabel]MenuEntry{},
		AddResetOption: true,
	}
}

// RequiredPartitions returns the list of partitions required by the bootloader.
func (c *Config) RequiredPartitions() []partition.Options {
	return []partition.Options{
		partition.NewPartitionOptions(constants.EFIPartitionLabel, false),
		partition.NewPartitionOptions(constants.BIOSGrubPartitionLabel, false),
		partition.NewPartitionOptions(constants.BootPartitionLabel, false),
	}
}

// Put puts a new menu entry to the grub config (nothing is written to disk).
func (c *Config) Put(entry BootLabel, cmdline, version string) error {
	c.Entries[entry] = buildMenuEntry(entry, cmdline, version)

	return nil
}

func (c *Config) validate() error {
	if _, ok := c.Entries[c.Default]; !ok {
		return fmt.Errorf("invalid default entry: %s", c.Default)
	}

	if c.Fallback != "" {
		if _, ok := c.Entries[c.Fallback]; !ok {
			return fmt.Errorf("invalid fallback entry: %s", c.Fallback)
		}
	}

	if c.Default == c.Fallback {
		return errors.New("default and fallback entries must not be the same")
	}

	return nil
}

func buildMenuEntry(entry BootLabel, cmdline, versionTag string) MenuEntry {
	return MenuEntry{
		Name:    fmt.Sprintf("%s - %s %s", entry, version.Name, versionTag),
		Linux:   filepath.Join("/", string(entry), constants.KernelAsset),
		Cmdline: cmdline,
		Initrd:  filepath.Join("/", string(entry), constants.InitramfsAsset),
	}
}
