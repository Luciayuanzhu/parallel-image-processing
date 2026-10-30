"""Run the timing experiments, then calculate speedups and draw graphs."""
import argparse
import csv
import json
import os
from pathlib import Path
import platform
import subprocess
import sys
import time

HERE = Path(__file__).resolve().parent
ROOT = HERE.parent
DATASETS = ("small", "mixture", "big")
THREADS = (2, 4, 6, 8, 12)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", default="results-cluster")
    parser.add_argument("--repeats", type=int, default=5)
    args = parser.parse_args()
    if args.repeats < 5:
        parser.error("The assignment needs at least five trials.")
    output = (HERE / args.output).resolve()
    output.mkdir(parents=True, exist_ok=True)
    binary = HERE / "bin" / "measure"
    binary.parent.mkdir(exist_ok=True)
    subprocess.run(["go", "build", "-o", str(binary), "./benchmark/cmd"], cwd=ROOT, check=True)
    metadata = {
        "platform": platform.platform(), "machine": platform.machine(),
        "host": platform.node(), "cpu_count": os.cpu_count(),
        "go": subprocess.check_output(["go", "version"], text=True).strip(),
        "slurm_job": os.getenv("SLURM_JOB_ID", ""),
        "partition": os.getenv("SLURM_JOB_PARTITION", ""),
        "gomaxprocs": os.getenv("GOMAXPROCS", "Go default"),
        "date": time.strftime("%Y-%m-%d %H:%M:%S %Z"),
        "repeats": args.repeats,
        "png_compression": "BestSpeed (lossless)",
    }
    (output / "environment.json").write_text(json.dumps(metadata, indent=2))
    fields = ["dataset", "mode", "threads", "trial", "total", "load", "filter", "save"]
    with (output / "timings.csv").open("w", newline="") as file:
        writer = csv.DictWriter(file, fieldnames=fields)
        writer.writeheader()
        for dataset in DATASETS:
            settings = [("s", 1)] + [(mode, n) for mode in ("parfiles", "parslices") for n in THREADS]
            for mode, n in settings:
                for trial in range(1, args.repeats + 1):
                    env = os.environ.copy()
                    env["PROJ1_PROFILE"] = "1" if mode == "s" else "0"
                    # The sequential baseline runs with one Go execution slot.
                    # Parallel runs keep the allocated CPU limit (16 on Slurm).
                    if mode == "s":
                        env["GOMAXPROCS"] = "1"
                    result = subprocess.run([str(binary), dataset, mode, str(n)], cwd=ROOT / "editor",
                                            env=env, text=True, capture_output=True, check=True)
                    row = dict(dataset=dataset, mode=mode, threads=n, trial=trial,
                               **json.loads(result.stdout))
                    if mode == "s":
                        row.update(json.loads(result.stderr))
                    writer.writerow(row)
                    file.flush()
                print(f"Finished {dataset}: {mode}, {n} workers", flush=True)
    subprocess.run([sys.executable, str(HERE / "analyze.py"), str(output)], check=True)


if __name__ == "__main__":
    main()
