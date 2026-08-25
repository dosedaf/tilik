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

if __name__ == "__main__":
    try:
        df = pd.read_csv("/home/yoda/projects/tilik/data/spse/wonogirikab/2026/spse_nontender_20260821_145425.csv")
        # dtypes -> tanggal pembuatan harusnya date, oap bool?, tahun date, jumlah int
        # isnull -> pemenang ama kontark ad yg kosong
        # white space di lokasi pekerjaan
        # assume schema is diff like diff year diff pemda diff schema
        # but since i know the scraper, ts not relevant
        #
        #
        # kategori                           str
        # kode_paket                       int64
        # nama_paket                         str
        # tanggal_pembuatan                  str
        # k/l/pd/instansi_lainnya            str
        # status                             str
        # satuan_kerja                       str
        # jenis_pengadaan                    str
        # metode_pengadaan                   str
        # khusus_orang_asli_papua_(oap)      str
        # tahun_anggaran                     str
        # nilai_pagu_(dalam_rupiah)        int64
        # nilai_hps_(dalam_rupiah)         int64
        # jenis_kontrak                      str
        # lokasi_pekerjaan                   str
        # jumlah_peserta                     str
        # pemenang                           str
        # pemenang_berkontrak                str
        # url                                str
        #
        # null aman
        # duplicated jg aman
        print(df.dtypes)
        
        # standarize column names
        df.columns = df.columns.str.strip().str.lower().str.replace(" ", "_")

        # clean missing values -> GAADA
        # clean and standarize values -> clean enough icl
        # no duplicates
        # handle invalid data
        # # decide what is invalid for each data type
        # # yg jelas ya hps pagu gaboleh min, peserta juga. keknya sisanya aman sih, kek kalo kategorikal mau diapain ege
        #
        
        words_to_remove = {"tender", "paket", "seleksi", "ulang", "batal", "gagal"}
        text = "bruh"

        for col in df.columns:
            if "kode" in col:
                df['kode'] = df[col]
                df = df.drop(columns=col)

            if "nama" in col:
                df['nama'] = df[col]
                df = df.drop(columns=col)


        cols_to_move = ['kategori', 'nama', 'kode']

        new_order = cols_to_move + [col for col in df.columns if col not in cols_to_move]

        df = df.reindex(columns=new_order)

        df['nama'] = df['nama'].apply(
            lambda x: " ".join([word for word in x.split() if word not in words_to_remove])
        )
        
        # fix data type + clean values / format
        # tanggal_pembuatan, sekalian parse date
        df['tanggal_pembuatan'] = df['tanggal_pembuatan'].replace(month_map, regex=True)
        df['tanggal_pembuatan'] = pd.to_datetime(df['tanggal_pembuatan'], format='%d %b %Y')

        # ke bool
        df['khusus_orang_asli_papua_(oap)'] = (df['khusus_orang_asli_papua_(oap)'] != 'Tidak').astype(bool)

        # ke int
        df['jumlah_peserta'] = df['jumlah_peserta'].str.split().str[0].astype(int)

        # fix duplicated words
        df['tahun_anggaran'] = df['tahun_anggaran'].apply(
            lambda x: " ".join(dict.fromkeys(x.split()))
        )

        # handle invalid data
        if (df['nilai_hps_(dalam_rupiah)'] < 0).sum() != 0:
            print("hps aneh")
        if (df['nilai_pagu_(dalam_rupiah)'] < 0).sum() != 0:
            print("pagu aneh")

        print(df.dtypes)
        df.to_csv("tes.csv")

    except Exception as e:
        logging.error(f"{e}")


