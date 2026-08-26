from pathlib import Path
import os

base_path = Path('/home/yoda/projects/tilik/data/spse')

csvs = []

# gimana cara pilih kalo ad bnyk dobel2 ingestionnya
for pemda in base_path.iterdir():
    if pemda.is_dir():
        os.chdir(pemda)
        for year in os.listdir():
            os.chdir(year)
            for csv in os.listdir():
                if 'tender' in csv:
                    csvs.append(csv)
            os.chdir('..')
        os.chdir(base_path)

print(csvs)
