import pandas as pd
import logging
from pathlib import Path
from dataclasses import dataclass
from typing import Callable

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

month_map = {
    "Januari": "Jan", "Februari": "Feb", "Maret": "Mar", "April": "Apr",
    "Mei": "May", "Juni": "Jun", "Juli": "Jul", "Agustus": "Aug",
    "September": "Sep", "Oktober": "Oct", "November": "Nov", "Desember": "Dec",
}

def normalize_columns(df: pd.DataFrame) -> pd.DataFrame:
    df.columns = df.columns.str.strip().str.lower().str.replace(" ", "_")
    return df


def parse_indo_date(series: pd.Series) -> pd.Series:
    series = series.replace(month_map, regex=True)
    return pd.to_datetime(series, format='%d %b %Y')


def dedupe_words(series: pd.Series) -> pd.Series:
    """Collapses repeated tokens, e.g. 'APBD APBD' -> 'APBD'."""
    return series.astype(str).apply(lambda x: " ".join(dict.fromkeys(x.split())))


def split_list_field(series: pd.Series, sep: str = ";") -> pd.Series:
    return series.fillna("").astype(str).apply(
        lambda x: sep.join(part.strip() for part in x.split(sep) if part.strip())
    )


def require_columns(df: pd.DataFrame, required: set[str]):
    missing = required.difference(df.columns)
    assert not missing, f'missing columns: {missing}'


def transform_tender(df: pd.DataFrame) -> pd.DataFrame:
    df = normalize_columns(df)
    df['kode_tender'] = df['kode_tender'].astype(int)
    df['tanggal_pembuatan'] = parse_indo_date(df['tanggal_pembuatan'])
    df['tahun_anggaran'] = dedupe_words(df['tahun_anggaran'])
    df['pemenang_berkontrak'] = df['pemenang_berkontrak'].astype(str)
    return df


def validate_tender(df: pd.DataFrame):
    required = {
        'kategori', 'kode_tender', 'nama_tender', 'tanggal_pembuatan',
        'k/l/pd/instansi_lainnya', 'status', 'satuan_kerja', 'jenis_pengadaan',
        'metode_pengadaan', 'reverse_auction', 'tahun_anggaran',
        'nilai_pagu_(dalam_rupiah)', 'nilai_hps_(dalam_rupiah)', 'jenis_kontrak',
        'lokasi_pekerjaan', 'jumlah_peserta', 'pemenang', 'pemenang_berkontrak', 'url',
    }
    require_columns(df, required)

    assert df['kode_tender'].dtype == 'int64'
    assert pd.api.types.is_string_dtype(df['nama_tender'])
    assert pd.api.types.is_string_dtype(df['k/l/pd/instansi_lainnya'])
    assert pd.api.types.is_string_dtype(df['status'])
    assert pd.api.types.is_string_dtype(df['satuan_kerja'])
    assert pd.api.types.is_string_dtype(df['jenis_pengadaan'])
    assert pd.api.types.is_string_dtype(df['metode_pengadaan'])
    assert pd.api.types.is_string_dtype(df['reverse_auction'])
    assert pd.api.types.is_datetime64_dtype(df['tanggal_pembuatan'])
    assert pd.api.types.is_string_dtype(df['tahun_anggaran'])
    assert df['nilai_pagu_(dalam_rupiah)'].dtype == 'int64'
    assert df['nilai_hps_(dalam_rupiah)'].dtype == 'int64'
    assert pd.api.types.is_string_dtype(df['jenis_kontrak'])
    assert pd.api.types.is_string_dtype(df['lokasi_pekerjaan'])
    assert df['jumlah_peserta'].dtype == 'int64'
    assert pd.api.types.is_string_dtype(df['pemenang'])
    assert pd.api.types.is_string_dtype(df['pemenang_berkontrak'])
    assert pd.api.types.is_string_dtype(df['url'])

    assert (df['nilai_pagu_(dalam_rupiah)'] >= 0).all()
    assert (df['nilai_hps_(dalam_rupiah)'] >= 0).all()
    assert (df['jumlah_peserta'] >= 0).all()
    assert df['kode_tender'].notna().all()
    assert df['nama_tender'].notna().all()

def transform_nontender(df: pd.DataFrame) -> pd.DataFrame:
    df = normalize_columns(df)
    df['kode_paket'] = df['kode_paket'].astype(int)
    df['tanggal_pembuatan'] = parse_indo_date(df['tanggal_pembuatan'])
    df['tahun_anggaran'] = dedupe_words(df['tahun_anggaran'])
    df['pemenang_berkontrak'] = df['pemenang_berkontrak'].astype(str)
    return df


def validate_nontender(df: pd.DataFrame):
    required = {
        'kategori', 'kode_paket', 'nama_paket', 'tanggal_pembuatan',
        'k/l/pd/instansi_lainnya', 'status', 'satuan_kerja', 'jenis_pengadaan',
        'metode_pengadaan', 'khusus_orang_asli_papua_(oap)', 'tahun_anggaran',
        'nilai_pagu_(dalam_rupiah)', 'nilai_hps_(dalam_rupiah)', 'jenis_kontrak',
        'lokasi_pekerjaan', 'jumlah_peserta',
        'pemenang', 'pemenang_berkontrak', 'url',
        # belom
        # syarat_kualifikasi, 
    }
    require_columns(df, required)

    assert df['kode_paket'].dtype == 'int64'
    assert pd.api.types.is_string_dtype(df['nama_paket'])
    assert pd.api.types.is_datetime64_dtype(df['tanggal_pembuatan'])
    assert pd.api.types.is_string_dtype(df['tahun_anggaran'])
    assert df['nilai_pagu_(dalam_rupiah)'].dtype == 'int64'
    assert df['nilai_hps_(dalam_rupiah)'].dtype == 'int64'
    assert df['jumlah_peserta'].dtype == 'int64'
    assert pd.api.types.is_string_dtype(df['pemenang'])
    assert pd.api.types.is_string_dtype(df['pemenang_berkontrak'])

    assert (df['nilai_pagu_(dalam_rupiah)'] >= 0).all()
    assert (df['nilai_hps_(dalam_rupiah)'] >= 0).all()
    assert (df['jumlah_peserta'] >= 0).all()
    assert df['kode_paket'].notna().all()
    assert df['nama_paket'].notna().all()

def transform_pencatatan(df: pd.DataFrame) -> pd.DataFrame:
    df = normalize_columns(df)
    df['kode_paket'] = df['kode_paket'].astype(int)
    df['tanggal_pembuatan'] = parse_indo_date(df['tanggal_pembuatan'])
    df['tahun_anggaran'] = dedupe_words(df['tahun_anggaran'])
    # df['pemenang_berkontrak'] = split_list_field(df['pemenang_berkontrak'])
    df['realisasi'] = split_list_field(df['realisasi'])
    return df


def validate_pencatatan(df: pd.DataFrame):
    required = {
        'kategori', 'kode_paket', 'nama_paket', 'tanggal_pembuatan',
        'k/l/pd/instansi_lainnya', 'status', 'satuan_kerja', 'jenis_pengadaan',
        'metode_pengadaan', 'tahun_anggaran', 'nilai_pagu_paket_(dalam_rupiah)',
        'realisasi', 'url',
    }
    require_columns(df, required)

    assert df['kode_paket'].dtype == 'int64'
    assert pd.api.types.is_string_dtype(df['nama_paket'])
    assert pd.api.types.is_datetime64_dtype(df['tanggal_pembuatan'])
    assert pd.api.types.is_string_dtype(df['tahun_anggaran'])
    assert df['nilai_pagu_paket_(dalam_rupiah)'].dtype == 'int64'
    # assert pd.api.types.is_string_dtype(df['pemenang_berkontrak'])
    assert pd.api.types.is_string_dtype(df['realisasi'])

    assert (df['nilai_pagu_paket_(dalam_rupiah)'] >= 0).all()
    assert df['kode_paket'].notna().all()
    assert df['nama_paket'].notna().all()

def transform_swakelola(df: pd.DataFrame) -> pd.DataFrame:
    df = normalize_columns(df)
    df['kode_swakelola'] = df['kode_swakelola'].astype(int)
    df['tanggal_pembuatan'] = parse_indo_date(df['tanggal_pembuatan'])
    df['tahun_anggaran'] = dedupe_words(df['tahun_anggaran'])
    # df['pelaksana'] = split_list_field(df['pelaksana'])
    df['realisasi'] = split_list_field(df['realisasi'])
    return df


def validate_swakelola(df: pd.DataFrame):
    required = {
        'kategori', 'kode_swakelola', 'nama_swakelola', 'tanggal_pembuatan',
        'k/l/pd', 'status', 'satuan_kerja', 'tahun_anggaran',
        'nilai_pagu_paket_(dalam_rupiah)', 
        'realisasi', 'url',
    }
    require_columns(df, required)

    assert df['kode_swakelola'].dtype == 'int64'
    assert pd.api.types.is_string_dtype(df['nama_swakelola'])
    assert pd.api.types.is_datetime64_dtype(df['tanggal_pembuatan'])
    assert pd.api.types.is_string_dtype(df['tahun_anggaran'])
    assert df['nilai_pagu_paket_(dalam_rupiah)'].dtype == 'int64'
    # assert pd.api.types.is_string_dtype(df['pelaksana'])
    assert pd.api.types.is_string_dtype(df['realisasi'])

    assert (df['nilai_pagu_paket_(dalam_rupiah)'] >= 0).all()
    assert df['kode_swakelola'].notna().all()
    assert df['nama_swakelola'].notna().all()

@dataclass
class Category:
    filename: str
    transform: Callable[[pd.DataFrame], pd.DataFrame]
    validate: Callable[[pd.DataFrame], None]


CATEGORIES: dict[str, Category] = {
    "tender": Category("tender.csv", transform_tender, validate_tender),
    "nontender": Category("nontender.csv", transform_nontender, validate_nontender),
    "pencatatan": Category("pencatatan.csv", transform_pencatatan, validate_pencatatan),
    "swakelola": Category("swakelola.csv", transform_swakelola, validate_swakelola),
}

def discover_files(base_path: Path) -> dict[str, list[Path]]:
    return {
        cat_name: list(base_path.rglob(cat.filename))
        for cat_name, cat in CATEGORIES.items()
    }


def latest_ingestion_per_period(paths: list[Path]) -> list[Path]:
    best: dict[tuple[str, str], tuple[str, Path]] = {}
    skipped = 0

    for p in paths:
        try:
            pemda, year, ingestion_id = p.parts[-4], p.parts[-3], p.parts[-2]
        except IndexError:
            logging.warning(f"unexpected path shape, skipping: {p}")
            skipped += 1
            continue

        key = (pemda, year)

        if key not in best or ingestion_id > best[key][0]:
            if key in best:
                logging.info(
                    f"superseding ingestion '{best[key][0]}' "
                    f"with '{ingestion_id}' for {pemda}/{year}"
                )

            best[key] = (ingestion_id, p)

    if skipped:
        logging.warning(
            f"skipped {skipped} path(s) with unexpected structure"
        )

    return [p for _, p in best.values()]


base_path = Path('/home/yoda/projects/tilik/data/spse')

if __name__ == "__main__":
    files_by_category = discover_files(base_path)

    for cat_name, cat in CATEGORIES.items():
        raw_paths = files_by_category[cat_name]
        deduped_paths = latest_ingestion_per_period(raw_paths)

        logging.info(
            f"{cat_name}: {len(raw_paths)} file(s) found, "
            f"{len(deduped_paths)} kept after dedup"
        )

        for csv_path in deduped_paths:
            try:
                df = pd.read_csv(csv_path)

                df = cat.transform(df)
                cat.validate(df)

                rel = csv_path.relative_to(base_path)
                out_path = (
                    Path('data')
                    / rel.parent
                    / f'{rel.stem}_cleaned.csv'
                )

                out_path.parent.mkdir(
                    parents=True,
                    exist_ok=True,
                )

                df.to_csv(out_path, index=False)

                logging.info(f"wrote {out_path}")

            except Exception:
                logging.exception(
                    f"etl failed for {csv_path}"
                )
