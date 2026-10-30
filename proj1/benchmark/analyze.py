"""Calculate speedups and draw two graphs from the measured times."""
import csv
from pathlib import Path
import sys

import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt

DATASETS = ("small", "mixture", "big")
THREADS = (2, 4, 6, 8, 12)


def main(folder):
    folder = Path(folder)
    with (folder / "timings.csv").open() as file:
        rows = list(csv.DictReader(file))
    best = {}
    for row in rows:
        key = (row["dataset"], row["mode"], int(row["threads"]))
        if key not in best or float(row["total"]) < float(best[key]["total"]):
            best[key] = row

    results = []
    for dataset in DATASETS:
        baseline = best[dataset, "s", 1]
        total = float(baseline["total"])
        load, filtering, save = [float(baseline[k]) for k in ("load", "filter", "save")]
        for mode in ("parfiles", "parslices"):
            # Files divide all image work. Slices divide only the filters.
            parallel_time = load + filtering + save if mode == "parfiles" else filtering
            p = min(1.0, parallel_time / total)
            limit = 1 / (1 - p) if p < 1 else float("inf")
            for n in THREADS:
                elapsed = float(best[dataset, mode, n]["total"])
                results.append(dict(dataset=dataset, mode=mode, threads=n, baseline=total,
                                    parallel_seconds=parallel_time, p=p, minimum=elapsed,
                                    speedup=total / elapsed, amdahl=1 / (1 - p + p / n), limit=limit))
    with (folder / "summary.csv").open("w", newline="") as file:
        writer = csv.DictWriter(file, fieldnames=list(results[0]))
        writer.writeheader()
        writer.writerows(results)

    for mode, filename, title in (("parfiles", "speedup-images.png", "Multiple images in parallel"),
                                  ("parslices", "speedup-slice.png", "Slices of one image in parallel")):
        fig, ax = plt.subplots(figsize=(8, 4.5), constrained_layout=True)
        for dataset in DATASETS:
            values = [r["speedup"] for r in results if r["mode"] == mode and r["dataset"] == dataset]
            ax.plot(THREADS, values, marker="o", label=dataset)
        ax.axhline(1, color="gray", linestyle="--", linewidth=1)
        ax.set(xlabel="Number of workers", ylabel="Speedup (sequential time / parallel time)",
               title=title + "\nPeanut CPU", xticks=THREADS)
        ax.grid(alpha=0.25)
        ax.legend()
        fig.savefig(folder / filename, dpi=160)
        plt.close(fig)

    print(f"Saved summary.csv and two speedup graphs in {folder}")


if __name__ == "__main__":
    main(sys.argv[1])
