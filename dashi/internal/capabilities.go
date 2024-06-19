package internal

import "golang.org/x/sys/unix"

var capFlagMap = map[string]uintptr{
	"CAP_AUDIT_CONTROL":      unix.CAP_AUDIT_CONTROL,
	"CAP_AUDIT_READ":         unix.CAP_AUDIT_READ,
	"CAP_AUDIT_WRITE":        unix.CAP_AUDIT_WRITE,
	"CAP_BLOCK_SUSPEND":      unix.CAP_BLOCK_SUSPEND,
	"CAP_BPF":                unix.CAP_BPF,
	"CAP_CHECKPOINT_RESTORE": unix.CAP_CHECKPOINT_RESTORE,
	"CAP_CHOWN":              unix.CAP_CHOWN,
	"CAP_DAC_OVERRIDE":       unix.CAP_DAC_OVERRIDE,
	"CAP_DAC_READ_SEARCH":    unix.CAP_DAC_READ_SEARCH,
	"CAP_FOWNER":             unix.CAP_FOWNER,
	"CAP_FSETID":             unix.CAP_FSETID,
	"CAP_IPC_LOCK":           unix.CAP_IPC_LOCK,
	"CAP_IPC_OWNER":          unix.CAP_IPC_OWNER,
	"CAP_KILL":               unix.CAP_KILL,
	"CAP_LAST_CAP":           unix.CAP_LAST_CAP,
	"CAP_LEASE":              unix.CAP_LEASE,
	"CAP_LINUX_IMMUTABLE":    unix.CAP_LINUX_IMMUTABLE,
	"CAP_MAC_ADMIN":          unix.CAP_MAC_ADMIN,
	"CAP_MAC_OVERRIDE":       unix.CAP_MAC_OVERRIDE,
	"CAP_MKNOD":              unix.CAP_MKNOD,
	"CAP_NET_ADMIN":          unix.CAP_NET_ADMIN,
	"CAP_NET_BIND_SERVICE":   unix.CAP_NET_BIND_SERVICE,
	"CAP_NET_BROADCAST":      unix.CAP_NET_BROADCAST,
	"CAP_NET_RAW":            unix.CAP_NET_RAW,
	"CAP_PERFMON":            unix.CAP_PERFMON,
	"CAP_SETGID":             unix.CAP_SETGID,
	"CAP_SETFCAP":            unix.CAP_SETFCAP,
	"CAP_SETPCAP":            unix.CAP_SETPCAP,
	"CAP_SETUID":             unix.CAP_SETUID,
	"CAP_SYS_ADMIN":          unix.CAP_SYS_ADMIN,
	"CAP_SYS_BOOT":           unix.CAP_SYS_BOOT,
	"CAP_SYS_CHROOT":         unix.CAP_SYS_CHROOT,
	"CAP_SYS_MODULE":         unix.CAP_SYS_MODULE,
	"CAP_SYS_NICE":           unix.CAP_SYS_NICE,
	"CAP_SYS_PACCT":          unix.CAP_SYS_PACCT,
	"CAP_SYS_PTRACE":         unix.CAP_SYS_PTRACE,
	"CAP_SYS_RAWIO":          unix.CAP_SYS_RAWIO,
	"CAP_SYS_RESOURCE":       unix.CAP_SYS_RESOURCE,
	"CAP_SYS_TIME":           unix.CAP_SYS_TIME,
	"CAP_SYS_TTY_CONFIG":     unix.CAP_SYS_TTY_CONFIG,
	"CAP_SYSLOG":             unix.CAP_SYSLOG,
	"CAP_WAKE_ALARM":         unix.CAP_WAKE_ALARM,
}

func SetKeepCaps() error {
	if err := unix.Prctl(unix.PR_SET_KEEPCAPS, 1, 0, 0, 0); err != nil {
		return err
	}

	return nil
}

func ClearKeepCaps() error {
	if err := unix.Prctl(unix.PR_SET_KEEPCAPS, 0, 0, 0, 0); err != nil {
		return err
	}

	return nil
}
