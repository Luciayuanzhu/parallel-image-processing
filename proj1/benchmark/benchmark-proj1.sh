#!/bin/bash
#SBATCH --job-name=proj1_benchmark
#SBATCH --account=peanut_group
#SBATCH --partition=peanut-cpu
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem=14G
#SBATCH --exclusive
#SBATCH --time=01:00:00
#SBATCH --output=proj1-%j.out
#SBATCH --error=proj1-%j.err

set -euo pipefail
export PATH="$HOME/go/bin:$PATH"
# Submit from proj1/benchmark. Slurm puts a copy of this script elsewhere.
if [[ -n "${SLURM_JOB_ID:-}" ]]; then
    cd "$SLURM_SUBMIT_DIR"
    export GOMAXPROCS="$SLURM_CPUS_PER_TASK"
else
    cd "$(dirname "$0")"
fi
python3 run.py --output results-cluster
