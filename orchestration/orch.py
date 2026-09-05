from airflow.sdk import dag, task
from pathlib import Path
import pendulum

now = pendulum.now()

PROJECT_ROOT = Path(__file__).resolve().parent.parent
INGESTION_DIR = PROJECT_ROOT / "ingestion"
TRANSFORMATION_DIR = PROJECT_ROOT / "transformation"
VENV = f"{PROJECT_ROOT}/.venv/bin"

@dag(
    dag_id="tilik_dag_sigma",
    start_date=now
)
def tilik_dag():
    @task.bash
    def ingest() -> str:
        cmd = f'cd {INGESTION_DIR} && go run cmd/ingestion/main.go'
        return cmd

    @task.bash
    def transform() -> str:
        cmd =f'cd {TRANSFORMATION_DIR} && {VENV}/python silver.py'
        return cmd

    ingest() >> transform()

tilik_dag()
