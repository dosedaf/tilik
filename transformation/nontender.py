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
        'kode_paket',
        'nama_paket',
        'tanggal_pembuatan',
        'k/l/pd/instansi_lainnya',
        'status',
        'satuan_kerja',
        'jenis_pengadaan',
        'metode_pengadaan',
        'khusus_orang_asli_papua_(oap)',
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
    assert df['kode_paket'].dtype == 'int64'

    assert pd.api.types.is_string_dtype(df['nama_paket'])
    assert pd.api.types.is_string_dtype(df['k/l/pd/instansi_lainnya'])
    assert pd.api.types.is_string_dtype(df['status'])
    assert pd.api.types.is_string_dtype(df['satuan_kerja'])
    assert pd.api.types.is_string_dtype(df['jenis_pengadaan'])
    assert pd.api.types.is_string_dtype(df['metode_pengadaan'])

    assert pd.api.types.is_bool_dtype(df['khusus_orang_asli_papua_(oap)'])

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
    assert df['kode_paket'].notna().all()
    assert df['nama_paket'].notna().all()

if __name__ == "__main__":
    try:
        filename = "/home/yoda/projects/tilik/data/spse/wonogirikab/2026/spse_nontender_20260821_145425.csv"
        df = pd.read_csv(filename)

        # hal yang gua pelajarin
        # gimana kalo schema data beda2? pdhl di source yg sama.
        # jan di pretty2 in dah
        
        # kategori                           str -> AMAN
        # kode_paket                       int64 -> ganti ke kode (ntar di scraper aja)
        # nama_paket                         str -> sama, nama aja
        # tanggal_pembuatan                  str -> parse jadi datetime
        # k/l/pd/instansi_lainnya            str -> biarin, standarisasi kalo ada masalah
        # status                             str -> sama 
        # satuan_kerja                       str -> sama
        # jenis_pengadaan                    str -> sama
        # metode_pengadaan                   str -> sama
        # khusus_orang_asli_papua_(oap)      str -> parse ke bool
        # tahun_anggaran                     str -> clean value krn ada yg dup
        # nilai_pagu_(dalam_rupiah)        int64 -> pakein assertion dll
        # nilai_hps_(dalam_rupiah)         int64 -> sama
        # jenis_kontrak                      str -> biarin, standarisasi kalo ada masalah
        # lokasi_pekerjaan                   str -> sama
        # jumlah_peserta                     str -> parse jadi angka doang 
        # pemenang                           str -> biarin, kosong = normal
        # pemenang_berkontrak                str -> biarin, kosong = normal
        # url                                str -> biarin
        
        # standarize column names
        df.columns = df.columns.str.strip().str.lower().str.replace(" ", "_")

        # ini gaperlu kalo scraper udh diganti tp for now biarin aj
        # reindexxing
        
        # parse to date
        df['tanggal_pembuatan'] = df['tanggal_pembuatan'].replace(month_map, regex=True)
        df['tanggal_pembuatan'] = pd.to_datetime(df['tanggal_pembuatan'], format='%d %b %Y')

        oap_map = {
            "Ya": True,
            "Tidak": False,
        }

        # parse to bool
        df['khusus_orang_asli_papua_(oap)'] = (
            df['khusus_orang_asli_papua_(oap)']
            .replace(oap_map).astype(bool)
        )

        # parse to int
        df["jumlah_peserta"] = pd.to_numeric(
            df["jumlah_peserta"].str.extract(r"(\d+)")[0],
            errors="coerce"
)

        # delete dup words
        df['tahun_anggaran'] = df['tahun_anggaran'].apply(
            lambda x: " ".join(dict.fromkeys(x.split()))
        )

        validate(df)
        df.to_csv(f"{filename}_cleaned", index=False)

    except Exception as e:
        logging.exception("etl failed")


