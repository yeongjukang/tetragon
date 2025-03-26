// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon

//go:build !windows

package syscallinfo

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/cilium/tetragon/pkg/syscallinfo/arm32"
	"github.com/cilium/tetragon/pkg/syscallinfo/arm64"
	"github.com/cilium/tetragon/pkg/syscallinfo/i386"
	"github.com/cilium/tetragon/pkg/syscallinfo/x64"
)

// NB: file below was generated by cmd/dump-syscall-info

//go:embed syscalls.json
var syscalls_ []byte

// SyscallArgInfo is the name and the type (as string) of a syscall argument
type SyscallArgInfo struct {
	Name string
	Type string
}

// SyscallArgs is the arguments for a given syscall
type SyscallArgs []SyscallArgInfo

// syscall table: name -> []SyscallArgs
var sysargsInfo map[string]SyscallArgs

func init() {
	// parse syscall table
	err := json.Unmarshal(syscalls_, &sysargsInfo)
	if err != nil {
		panic(err)
	}
}

func syscallNames(abi string) (map[int]string, error) {

	switch abi {
	case "x64":
		return x64.Names, nil
	case "i386":
		return i386.Names, nil
	case "arm64":
		return arm64.Names, nil
	case "arm32":
		return arm32.Names, nil
	default:
		return nil, fmt.Errorf("unknown abi '%s'", abi)
	}
}

// return all syscall names for the given ABI
func SyscallsNames(abi string) ([]string, error) {
	names, err := syscallNames(abi)
	if err != nil {
		return nil, err
	}

	ret := make([]string, 0, len(names))
	for _, v := range names {
		ret = append(ret, v)
	}
	return ret, nil
}

// SyscallID returns the id of a syscall based on its name and an ABI
// ABI may be empty, in which case the default ABI is used (e.g., x64 for x86_64)
// returns -1, if no system call was found
func SyscallID(sysName string, abi string) (int, error) {
	names, err := syscallNames(abi)
	if err != nil {
		return -1, err
	}
	for id, name := range names {
		if name == sysName {
			return id, nil
		}
	}
	return -1, fmt.Errorf("syscall %s for abi %s was not found", sysName, abi)
}

// GetSyscallName returns the name of a syscall based on its id
func GetSyscallName(abi string, sysID int) (string, error) {
	names, err := syscallNames(abi)
	if err != nil {
		return "", err
	}

	ret, ok := names[sysID]
	if !ok {
		return "", fmt.Errorf("unknown syscall id: %d", sysID)
	}
	return ret, nil
}

// GetSyscallArgs returns the arguments of a system call
func GetSyscallArgs(name string) (SyscallArgs, bool) {
	if args, ok := sysargsInfo[name]; ok {
		ret := make([]SyscallArgInfo, len(args))
		copy(ret, args)
		return SyscallArgs(ret), true
	}
	return nil, false
}

// Proto returns a string representing a  prototype for the system call
func (sai SyscallArgs) Proto(name string) string {
	args := make([]string, 0, len(sai))
	for i := range sai {
		args = append(args, fmt.Sprintf("%s %s", sai[i].Type, sai[i].Name))
	}
	return fmt.Sprintf("long %s(%s)", name, strings.Join(args, ", "))
}

func DefaultABI() (string, error) {
	switch a := runtime.GOARCH; a {
	case "amd64":
		return "x64", nil
	case "arm64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("unsupported arch: %s", a)
	}

}
