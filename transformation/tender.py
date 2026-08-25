import pandas as pd
import logging

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

month_map = {
    "Januari": "Jan",
    "Februari": "Feb",
    "Maret": "Mar",
    "April": "Apr",
    "Mei": "May",
    "Juni": "Jun",
    "Juli": "Jul",
    "Agustus": "Aug",
    "September": "Sep",
    "Oktober": "Oct",
    "November": "Nov",
    "Desember": "Dec",
}

def validate(df):
    # required columns
    required = {
        'kategori',
        'kode_tender',
        'nama_tender',
        'tanggal_pembuatan',
        'k/l/pd/instansi_lainnya',
        'status',
        'satuan_kerja',
        'jenis_pengadaan',
        'metode_pengadaan',
        'reverse_auction',
        'tahun_anggaran',
        'nilai_pagu_(dalam_rupiah)',
        'nilai_hps_(dalam_rupiah)',
        'jenis_kontrak',
        'lokasi_pekerjaan',
        'jumlah_peserta',
        'pemenang',
        'pemenang_berkontrak',
        'url',
    }

    assert required.issubset(df.columns)

    # types
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

    #values
    assert (df['nilai_pagu_(dalam_rupiah)'] >= 0).all()
    assert (df['nilai_hps_(dalam_rupiah)'] >= 0).all()
    assert (df['jumlah_peserta'] >= 0).all()

    # required data
    assert df['kode_tender'].notna().all()
    assert df['nama_tender'].notna().all()

if __name__ == "__main__":
    try:
        df = pd.read_csv("/home/yoda/projects/tilik/data/spse/wonogirikab/2026/spse_tender_20260821_145421.csv")

        df.columns = df.columns.str.strip().str.lower().str.replace(" ", "_")

        df['tanggal_pembuatan'] = df['tanggal_pembuatan'].replace(month_map, regex=True)
        df['tanggal_pembuatan'] = pd.to_datetime(df['tanggal_pembuatan'], format='%d %b %Y')

        df['tahun_anggaran'] = df['tahun_anggaran'].apply(
            lambda x: " ".join(dict.fromkeys(x.split()))
        )

        df['pemenang_berkontrak'] = df['pemenang_berkontrak'].astype(str)

        validate(df)
        df.to_csv("cleaned.csv", index=False)

    except Exception as e:
        logging.exception("etl failed")


