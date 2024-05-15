import subprocess
import sys
import os


OUTPUT_DIR = "build"
TARGETS_PATH = "target_programs"
DASHI = "dashi"
YAMINABE = "yaminabe"
NIMONO = "nimono"


def run_cmd(cmd: str, dir: str = "./", ignore_error: bool = False):
    print(f"\033[32m{cmd}\033[0m")
    cp = subprocess.run(cmd, shell=True, cwd=dir)

    if cp.returncode != 0 and not ignore_error:
        print(f"returncode: {cp.returncode}")
        exit(0)


# tasks
def task_clear():
    run_cmd(f"rm -rf ./{OUTPUT_DIR}")


def task_build_dashi():
    run_cmd(f"go build -o ../{OUTPUT_DIR}/{DASHI}", dir=DASHI)


def task_build_yaminabe():
    run_cmd("cargo build", dir=YAMINABE)
    run_cmd(f"cp target/debug/{YAMINABE} ../{OUTPUT_DIR}/{YAMINABE}", dir=YAMINABE)


def task_build_specimens():
    d = f"./{TARGETS_PATH}"
    dirs = [f for f in os.listdir(d) if os.path.isdir(os.path.join(d, f))]

    for dir_name in dirs:
        pwd = f"{d}/{dir_name}"

        run_cmd("cargo build", dir=pwd)


def task_build_nimono():
    run_cmd("clang -O2 -g -c -target bpf hello.c", dir=NIMONO)
    run_cmd(f"go build -o ../{OUTPUT_DIR}/{NIMONO}", dir=NIMONO)
    run_cmd(f"cp hello.o ../{OUTPUT_DIR}", dir=NIMONO)


def task_build():
    task_clear()
    task_build_specimens()
    task_build_dashi()
    task_build_nimono()
    task_build_yaminabe()


def task_run_hello():
    task_build()
    run_cmd("./build/yaminabe -t ./target_programs/hello/target/debug/hello -o 20")


TASKS = [
    task_clear,
    task_build_dashi,
    task_build_yaminabe,
    task_build_specimens,
    task_build_nimono,
    task_build,
    task_run_hello,
]

if __name__ == "__main__":
    args = sys.argv

    if len(args) == 2:
        for task in TASKS:
            if task.__name__ == args[1]:
                task()
                exit(0)

        print("Invalid task name.")
    else:
        print(f"Usage: {list(map(lambda x: x.__name__, TASKS))}")
