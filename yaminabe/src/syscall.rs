use num_enum::TryFromPrimitive;

#[derive(Debug, TryFromPrimitive)]
#[repr(u32)]
pub enum X86_64Syscalls {
    Read,
    Write,
    Open,
    Close,
    Stat,
    Fstat,
    Lstat,
    Poll,
    Lseek,
    Mmap,
    Mprotect,
    Munmap,
    Brk,
    RtSigaction,
    RtSigprocmask,
    Ioctl,
    Pread,
    Pwrite,
    Readv,
    Writev,
    Access,
    Pipe,
    Select,
    SchedYield,
    Mremap,
    Msync,
    Mincore,
    Madvise,
    Shmget,
    Shmat,
    Shmctl,
    Dup,
    Dup2,
    Pause,
    Nanosleep,
    Getitimer,
    Alarm,
    Setitimer,
    Getpid,
    Sendfile,
    Socket,
    Connect,
    Accept,
    Sendto,
    Recvfrom,
    Sendmsg,
    Recvmsg,
    Shutdown,
    Bind,
    Listen,
    Getsockname,
    Getpeername,
    Socketpair,
    Setsockopt,
    Getsockopt,
    Clone,
    Fork,
    Vfork,
    Execve,
    Exit,
    Wait4,
    Kill,
    Uname,
    Semget,
    Semop,
    Semctl,
    Shmdt,
    Msgget,
    Msgsnd,
    Msgrcv,
    Msgctl,
    Fcntl,
    Flock,
    Fsync,
    Fdatasync,
    Truncate,
    Ftruncate,
    Getdents,
    Getcwd,
    Chdir,
    Fchdir,
    Rename,
    Mkdir,
    Rmdir,
    Creat,
    Link,
    Unlink,
    Symlink,
    Readlink,
    Chmod,
    Fchmod,
    Chown,
    Fchown,
    Lchown,
    Umask,
    Gettimeofday,
    Getrlimit,
    Getrusage,
    Sysinfo,
    Times,
    Ptrace,
    Getuid,
    Syslog,
    Getgid,
    Setuid,
    Setgid,
    Geteuid,
    Getegid,
    Setpgid,
    Getppid,
    Getpgrp,
    Setsid,
    Setreuid,
    Setregid,
    Getgroups,
    Setgroups,
    Setresuid,
    Getresuid,
    Setresgid,
    Getresgid,
    Getpgid,
    Setfsuid,
    Setfsgid,
    Getsid,
    Capget,
    Capset,
    RtSigpending,
    RtSigtimedwait,
    RtSigqueueinfo,
    RtSigsuspend,
    Sigaltstack,
    Utime,
    Mknod,
    Uselib,
    Personality,
    Ustat,
    Statfs,
    Fstatfs,
    Sysfs,
    Getpriority,
    Setpriority,
    SchedSetparam,
    SchedGetparam,
    SchedSetscheduler,
    SchedGetscheduler,
    SchedGetPriorityMax,
    SchedGetPriorityMin,
    SchedRrGetInterval,
    Mlock,
    Munlock,
    Mlockall,
    Munlockall,
    Vhangup,
    ModifyLdt,
    PivotRoot,
    Sysctl,
    Prctl,
    ArchPrctl,
    Adjtimex,
    Setrlimit,
    Chroot,
    Sync,
    Acct,
    Settimeofday,
    Mount,
    Umount2,
    Swapon,
    Swapoff,
    Reboot,
    Sethostname,
    Setdomainname,
    Iopl,
    Ioperm,
    CreateModule,
    InitModule,
    DeleteModule,
    GetKernelSyms,
    QueryModule,
    Quotactl,
    Nfsservctl,
    Getpmsg,
    Putpmsg,
    AfsSyscall,
    Tuxcall,
    Security,
    Gettid,
    Readahead,
    Setxattr,
    Lsetxattr,
    Fsetxattr,
    Getxattr,
    Lgetxattr,
    Fgetxattr,
    Listxattr,
    Llistxattr,
    Flistxattr,
    Removexattr,
    Lremovexattr,
    Fremovexattr,
    Tkill,
    Time,
    Futex,
    SchedSetaffinity,
    SchedGetaffinity,
    SetThreadArea,
    IoSetup,
    IoDestroy,
    IoGetevents,
    IoSubmit,
    IoCancel,
    GetThreadArea,
    LookupDcookie,
    EpollCreate,
    EpollCtlOld,
    EpollWaitOld,
    RemapFilePages,
    Getdents64,
    SetTidAddress,
    RestartSyscall,
    Semtimedop,
    Fadvise64,
    TimerCreate,
    TimerSettime,
    TimerGettime,
    TimerGetoverrun,
    TimerSelete,
    ClockSettime,
    ClockGettime,
    ClockGetres,
    ClockNanosleep,
    ExitGroup,
    EpollWait,
    EpollCtl,
    Tgkill,
    Utimes,
    Vserver,
    Mbind,
    SetMempolicy,
    GetMempolicy,
    MqOpen,
    MqUnlink,
    MqTimedsend,
    MqTimedreceive,
    MqNotify,
    MqGetsetattr,
    KexecLoad,
    Waitid,
    AddKey,
    RequestKey,
    Keyctl,
    IoprioSet,
    IoprioGet,
    InotifyInit,
    InotifyAddWatch,
    InotifyRmWatch,
    MigratePages,
    Openat,
    Mkdirat,
    Mknodat,
    Fchownat,
    Futimesat,
    Newfstatat,
    Unlinkat,
    Renameat,
    Linkat,
    Symlinkat,
    Readlinkat,
    Fchmodat,
    Faccessat,
    Pselect6,
    Ppoll,
    Unshare,
    SetRobustList,
    GetRobustList,
    Splice,
    Tee,
    SyncFileRange,
    Vmsplice,
    MovePages,
    Utimensat,
    EpollPwait,
    Signalfd,
    TimerfdCreate,
    Eventfd,
    Fallocate,
    TimerfdSettime,
    TimerfdGettime,
    Accept4,
    Signalfd4,
    Eventfd2,
    EpollCreate1,
    Dup3,
    Pipe2,
    InotifyInit1,
    Preadv,
    Pwritev,
    RtTgsigqueueinfo,
    PerfEventOpen,
    Recvmmsg,
    FanotifyInit,
    FanotifyMark,
    Prlimit64,
    NameToHandleAt,
    OpenByHandleAt,
    ClockAdjtime,
    Syncfs,
    Sendmmsg,
    Setns,
    Getcpu,
    ProcessVmReadv,
    ProcessVmWritev,
    Kcmp,
    FinitModule,
    SchedSetattr,
    SchedGetattr,
    Renameat2,
    Seccomp,
    Getrandom,
    MemfdCreate,
    KexecFileLoad,
    Bpf,
    Execveat,
    Userfaultfd,
    Membarrier,
    Mlock2,
    CopyFileRange,
    Preadv2,
    Pwritev2,
    PkeyMprotect,
    PkeyAlloc,
    PkeyFree,
    Statx,
    IoPgetevents,
    Rseq,
    PidfdSendSignal,
    IoUringSetup,
    IoUringEnter,
    IoUringRegister,
    OpenTree,
    MoveMount,
    Fsopen,
    Fsconfig,
    Fsmount,
    Fspick,
    PidfdOpen,
    Clone3,
    CloseRange,
    Openat2,
    PidfdGetfd,
    Faccessat2,
    ProcessMadvise,
    EpollPwait2,
    MountSetattr,
    QuotactlFd,
    LandlockCreateRuleset,
    LandlockAddRule,
    LandlockRestrictSelf,
    MemfdSecret,
    ProcessMrelease,
    FutexWaitv,
    SetMempolicyHomeNode,
    Cachestat,
    Fchmodat2,
    MapShadowStack,
    FutexWake,
    FutexWait,
    FutexRequeue,
    Statmount,
    Listmount,
    LsmGetSelfAttr,
    LsmSetSelfAttr,
    LsmListModules,
}