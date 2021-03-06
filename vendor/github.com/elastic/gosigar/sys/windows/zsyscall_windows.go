// MACHINE GENERATED BY 'go generate' COMMAND; DO NOT EDIT

package windows

import "unsafe"
import "syscall"

var _ unsafe.Pointer

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")
	modpsapi    = syscall.NewLazyDLL("psapi.dll")
	modntdll    = syscall.NewLazyDLL("ntdll.dll")
	modadvapi32 = syscall.NewLazyDLL("advapi32.dll")

	procGlobalMemoryStatusEx      = modkernel32.NewProc("GlobalMemoryStatusEx")
	procGetLogicalDriveStringsW   = modkernel32.NewProc("GetLogicalDriveStringsW")
	procGetProcessMemoryInfo      = modpsapi.NewProc("GetProcessMemoryInfo")
	procGetProcessImageFileNameW  = modpsapi.NewProc("GetProcessImageFileNameW")
	procGetSystemTimes            = modkernel32.NewProc("GetSystemTimes")
	procGetDriveTypeW             = modkernel32.NewProc("GetDriveTypeW")
	procEnumProcesses             = modpsapi.NewProc("EnumProcesses")
	procGetDiskFreeSpaceExW       = modkernel32.NewProc("GetDiskFreeSpaceExW")
	procProcess32FirstW           = modkernel32.NewProc("Process32FirstW")
	procProcess32NextW            = modkernel32.NewProc("Process32NextW")
	procCreateToolhelp32Snapshot  = modkernel32.NewProc("CreateToolhelp32Snapshot")
	procNtQuerySystemInformation  = modntdll.NewProc("NtQuerySystemInformation")
	procNtQueryInformationProcess = modntdll.NewProc("NtQueryInformationProcess")
	procLookupPrivilegeNameW      = modadvapi32.NewProc("LookupPrivilegeNameW")
	procLookupPrivilegeValueW     = modadvapi32.NewProc("LookupPrivilegeValueW")
	procAdjustTokenPrivileges     = modadvapi32.NewProc("AdjustTokenPrivileges")
)

func _GlobalMemoryStatusEx(buffer *MemoryStatusEx) (err error) {
	r1, _, e1 := syscall.Syscall(procGlobalMemoryStatusEx.Addr(), 1, uintptr(unsafe.Pointer(buffer)), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _GetLogicalDriveStringsW(bufferLength uint32, buffer *uint16) (length uint32, err error) {
	r0, _, e1 := syscall.Syscall(procGetLogicalDriveStringsW.Addr(), 2, uintptr(bufferLength), uintptr(unsafe.Pointer(buffer)), 0)
	length = uint32(r0)
	if length == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _GetProcessMemoryInfo(handle syscall.Handle, psmemCounters *ProcessMemoryCountersEx, cb uint32) (err error) {
	r1, _, e1 := syscall.Syscall(procGetProcessMemoryInfo.Addr(), 3, uintptr(handle), uintptr(unsafe.Pointer(psmemCounters)), uintptr(cb))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _GetProcessImageFileName(handle syscall.Handle, outImageFileName *uint16, size uint32) (length uint32, err error) {
	r0, _, e1 := syscall.Syscall(procGetProcessImageFileNameW.Addr(), 3, uintptr(handle), uintptr(unsafe.Pointer(outImageFileName)), uintptr(size))
	length = uint32(r0)
	if length == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _GetSystemTimes(idleTime *syscall.Filetime, kernelTime *syscall.Filetime, userTime *syscall.Filetime) (err error) {
	r1, _, e1 := syscall.Syscall(procGetSystemTimes.Addr(), 3, uintptr(unsafe.Pointer(idleTime)), uintptr(unsafe.Pointer(kernelTime)), uintptr(unsafe.Pointer(userTime)))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _GetDriveType(rootPathName *uint16) (dt DriveType, err error) {
	r0, _, e1 := syscall.Syscall(procGetDriveTypeW.Addr(), 1, uintptr(unsafe.Pointer(rootPathName)), 0, 0)
	dt = DriveType(r0)
	if dt == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _EnumProcesses(processIds *uint32, sizeBytes uint32, bytesReturned *uint32) (err error) {
	r1, _, e1 := syscall.Syscall(procEnumProcesses.Addr(), 3, uintptr(unsafe.Pointer(processIds)), uintptr(sizeBytes), uintptr(unsafe.Pointer(bytesReturned)))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _GetDiskFreeSpaceEx(directoryName *uint16, freeBytesAvailable *uint64, totalNumberOfBytes *uint64, totalNumberOfFreeBytes *uint64) (err error) {
	r1, _, e1 := syscall.Syscall6(procGetDiskFreeSpaceExW.Addr(), 4, uintptr(unsafe.Pointer(directoryName)), uintptr(unsafe.Pointer(freeBytesAvailable)), uintptr(unsafe.Pointer(totalNumberOfBytes)), uintptr(unsafe.Pointer(totalNumberOfFreeBytes)), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _Process32First(handle syscall.Handle, processEntry32 *ProcessEntry32) (err error) {
	r1, _, e1 := syscall.Syscall(procProcess32FirstW.Addr(), 2, uintptr(handle), uintptr(unsafe.Pointer(processEntry32)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _Process32Next(handle syscall.Handle, processEntry32 *ProcessEntry32) (err error) {
	r1, _, e1 := syscall.Syscall(procProcess32NextW.Addr(), 2, uintptr(handle), uintptr(unsafe.Pointer(processEntry32)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _CreateToolhelp32Snapshot(flags uint32, processID uint32) (handle syscall.Handle, err error) {
	r0, _, e1 := syscall.Syscall(procCreateToolhelp32Snapshot.Addr(), 2, uintptr(flags), uintptr(processID), 0)
	handle = syscall.Handle(r0)
	if handle == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _NtQuerySystemInformation(systemInformationClass uint32, systemInformation *byte, systemInformationLength uint32, returnLength *uint32) (nhaaatus uint32, err error) {
	r0, _, e1 := syscall.Syscall6(procNtQuerySystemInformation.Addr(), 4, uintptr(systemInformationClass), uintptr(unsafe.Pointer(systemInformation)), uintptr(systemInformationLength), uintptr(unsafe.Pointer(returnLength)), 0, 0)
	nhaaatus = uint32(r0)
	if nhaaatus == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _NtQueryInformationProcess(processHandle syscall.Handle, processInformationClass uint32, processInformation *byte, processInformationLength uint32, returnLength *uint32) (nhaaatus uint32, err error) {
	r0, _, e1 := syscall.Syscall6(procNtQueryInformationProcess.Addr(), 5, uintptr(processHandle), uintptr(processInformationClass), uintptr(unsafe.Pointer(processInformation)), uintptr(processInformationLength), uintptr(unsafe.Pointer(returnLength)), 0)
	nhaaatus = uint32(r0)
	if nhaaatus == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _LookupPrivilegeName(systemName string, luid *int64, buffer *uint16, size *uint32) (err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(systemName)
	if err != nil {
		return
	}
	return __LookupPrivilegeName(_p0, luid, buffer, size)
}

func __LookupPrivilegeName(systemName *uint16, luid *int64, buffer *uint16, size *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procLookupPrivilegeNameW.Addr(), 4, uintptr(unsafe.Pointer(systemName)), uintptr(unsafe.Pointer(luid)), uintptr(unsafe.Pointer(buffer)), uintptr(unsafe.Pointer(size)), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _LookupPrivilegeValue(systemName string, name string, luid *int64) (err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(systemName)
	if err != nil {
		return
	}
	var _p1 *uint16
	_p1, err = syscall.UTF16PtrFromString(name)
	if err != nil {
		return
	}
	return __LookupPrivilegeValue(_p0, _p1, luid)
}

func __LookupPrivilegeValue(systemName *uint16, name *uint16, luid *int64) (err error) {
	r1, _, e1 := syscall.Syscall(procLookupPrivilegeValueW.Addr(), 3, uintptr(unsafe.Pointer(systemName)), uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(luid)))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _AdjustTokenPrivileges(token syscall.Token, releaseAll bool, input *byte, outputSize uint32, output *byte, requiredSize *uint32) (success bool, err error) {
	var _p0 uint32
	if releaseAll {
		_p0 = 1
	} else {
		_p0 = 0
	}
	r0, _, e1 := syscall.Syscall6(procAdjustTokenPrivileges.Addr(), 6, uintptr(token), uintptr(_p0), uintptr(unsafe.Pointer(input)), uintptr(outputSize), uintptr(unsafe.Pointer(output)), uintptr(unsafe.Pointer(requiredSize)))
	success = r0 != 0
	if true {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
