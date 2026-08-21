import pandas as pd
import requests
import logging

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

def transform(raw_data: pd.DataFrame) -> pd.DataFrame:
    df = pd.DataFrame(raw_data)
    print(df.columns)

    return df

if __name__ == "__main__":
    try:
        raw_data = pd.read_csv("/home/yoda/projects/tilik/data/spse/wonogirikab/2026/spse_nontender_20260820_225319.csv")
        clean_data = transform(raw_data)
    except Exception as e:
        logging.error(f"{e}")


