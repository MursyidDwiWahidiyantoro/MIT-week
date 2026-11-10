/**
 * Copy for the Belajar page (/edukasi): short cards and the quiz.
 * Plain words for laypeople and the elderly. Both languages live side by side
 * so a card and its translation are edited together.
 */
export type Text = { id: string };

export interface Card {
  id: string;
  title: Text;
  body: Text;
}

export interface Scam extends Card {
  /** What to do, in one sentence. */
  act: Text;
}

export interface CheckStep extends Card {
  /** External tool, opened in a new tab. */
  link?: { href: string; label: string };
}

export interface Question {
  id: string;
  question: Text;
  options: Array<{ id: string; text: Text }>;
  answer: string;
  explain: Text;
}

export const AI_CUES: Card[] = [
  {
    id: 'hands',
    title: { id: 'Tangan dan jari' },
    body: {
      id: 'Jumlah jari tidak pas, jari menyatu, atau cara memegang benda terlihat aneh. AI masih sering salah di bagian ini.'
    }
  },
  {
    id: 'text',
    title: { id: 'Tulisan' },
    body: {
      id: 'Huruf di papan, kaos, atau layar berubah-ubah, buram, atau tidak membentuk kata yang masuk akal.'
    }
  },
  {
    id: 'physics',
    title: { id: 'Hukum alam' },
    body: {
      id: 'Air, asap, kain, dan rambut bergerak tidak wajar. Benda menembus satu sama lain, muncul, atau hilang.'
    }
  },
  {
    id: 'light',
    title: { id: 'Cahaya dan bayangan' },
    body: {
      id: 'Bayangan jatuh ke arah yang berbeda-beda. Pantulan di kaca atau mata tidak cocok dengan sekitarnya.'
    }
  },
  {
    id: 'lips',
    title: { id: 'Bibir dan suara' },
    body: {
      id: 'Gerak bibir tidak pas dengan ucapan. Kedipan jarang atau kaku. Kulit wajah terlalu halus.'
    }
  },
  {
    id: 'smooth',
    title: { id: 'Gerakan terlalu sempurna' },
    body: {
      id: 'Kamera meluncur mulus, orang bergerak seperti di iklan, tanpa goyangan kecil seperti video dari ponsel biasa.'
    }
  }
];

export const HOAX_CUES: Card[] = [
  {
    id: 'old',
    title: { id: 'Video lama, keterangan baru' },
    body: {
      id: 'Rekaman asli dari tahun atau tempat lain, diberi keterangan seolah baru terjadi. Banjir, kebakaran, dan kerusuhan sering dipakai.'
    }
  },
  {
    id: 'cut',
    title: { id: 'Dipotong menyesatkan' },
    body: {
      id: 'Hanya sebagian adegan yang ditampilkan, atau kalimat dipotong sehingga artinya berbalik dari yang sebenarnya.'
    }
  },
  {
    id: 'voice',
    title: { id: 'Suara atau narasi palsu' },
    body: {
      id: 'Suara tokoh ditiru, atau narasi baru ditempel di video peristiwa lain. Cocokkan isi ucapan dengan gambarnya.'
    }
  },
  {
    id: 'headline',
    title: { id: 'Tangkapan layar berita diedit' },
    body: {
      id: 'Judul diubah agar terdengar lebih mengejutkan. Buka berita aslinya di situs resminya sebelum percaya.'
    }
  }
];

export const CHECK_STEPS: CheckStep[] = [
  {
    id: 'tbh',
    title: { id: 'TurnBackHoax' },
    body: {
      id: 'Ketik kata kunci kabar itu. Situs ini dikelola Mafindo dan memuat ribuan kabar yang sudah diperiksa.'
    },
    link: { href: 'https://turnbackhoax.id', label: 'turnbackhoax.id' }
  },
  {
    id: 'cekfakta',
    title: { id: 'cekfakta.com' },
    body: {
      id: 'Kerja sama banyak media di Indonesia. Cari kabar yang sama di sini untuk membandingkan hasil.'
    },
    link: { href: 'https://cekfakta.com', label: 'cekfakta.com' }
  },
  {
    id: 'reverse',
    title: { id: 'Cari gambar terbalik' },
    body: {
      id: 'Tangkap layar satu adegan video, lalu unggah ke Google Lens. Hasilnya sering menunjukkan kapan dan di mana gambar itu pertama muncul.'
    },
    link: { href: 'https://lens.google.com', label: 'lens.google.com' }
  },
  {
    id: 'official',
    title: { id: 'Akun dan situs resmi' },
    body: {
      id: 'Kalau kabar menyebut instansi atau tokoh, cek pengumuman di akun resmi mereka. Jangan lewat tautan di pesan.'
    }
  }
];

export const SCAMS: Scam[] = [
  {
    id: 'bansos',
    title: { id: 'Bansos palsu' },
    body: {
      id: 'Pesan bilang Anda terpilih menerima bantuan, lalu meminta Anda mengisi data KTP atau memasukkan OTP lewat tautan.'
    },
    act: {
      id: 'Abaikan. Cek hak bantuan hanya lewat situs atau aplikasi resmi.'
    }
  },
  {
    id: 'undian',
    title: { id: 'Undian dan hadiah' },
    body: {
      id: 'Anda menang mobil atau uang, tetapi harus transfer biaya pajak atau administrasi lebih dulu.'
    },
    act: {
      id: 'Undian sungguhan tidak meminta bayar di depan. Putuskan kontak.'
    }
  },
  {
    id: 'petugas',
    title: { id: 'Petugas atau pejabat palsu' },
    body: {
      id: 'Mengaku dari bank, kurir, polisi, atau kantor pemerintah. Biasanya mendesak dan menakut-nakuti supaya Anda panik.'
    },
    act: {
      id: 'Tutup telepon, lalu hubungi nomor resmi yang Anda cari sendiri.'
    }
  },
  {
    id: 'deepfake',
    title: { id: 'Video pejabat hasil deepfake' },
    body: {
      id: 'Wajah dan suara tokoh ditiru untuk mengajak berinvestasi, menyumbang, atau mengklaim hadiah.'
    },
    act: {
      id: 'Cek di akun resmi tokoh itu dan di berita. Jangan percaya satu video saja.'
    }
  }
];

export const QUESTIONS: Question[] = [
  {
    id: 'hands',
    question: {
      id: 'Di sebuah video, tangan seseorang punya enam jari dan jarinya menyatu saat memegang gelas. Apa yang paling mungkin?'
    },
    options: [
      { id: 'a', text: { id: 'Kamera ponselnya kurang bagus.' } },
      { id: 'b', text: { id: 'Video itu kemungkinan dibuat atau diubah dengan AI.' } },
      { id: 'c', text: { id: 'Video itu pasti asli karena terlihat jelas.' } }
    ],
    answer: 'b',
    explain: {
      id: 'AI masih sering salah menggambar tangan. Ini bukan bukti pasti, tetapi alasan kuat untuk memeriksa sumbernya.'
    }
  },
  {
    id: 'old',
    question: {
      id: 'Video banjir besar beredar dengan keterangan “terjadi hari ini di kota kita”. Langkah pertama yang paling tepat?'
    },
    options: [
      { id: 'a', text: { id: 'Kirim ke grup supaya semua waspada.' } },
      { id: 'b', text: { id: 'Cari kapan dan di mana video itu pertama muncul.' } },
      { id: 'c', text: { id: 'Percaya karena banyak yang membagikan.' } }
    ],
    answer: 'b',
    explain: {
      id: 'Video lama sering dipasangi keterangan baru. Pencarian gambar terbalik biasanya menunjukkan kemunculan pertamanya.'
    }
  },
  {
    id: 'bansos',
    question: {
      id: 'Pesan WhatsApp: “Anda terpilih menerima bansos Rp 2.000.000. Klik tautan, isi data KTP dan kode OTP.” Sikap yang tepat?'
    },
    options: [
      { id: 'a', text: { id: 'Isi secepatnya supaya tidak hangus.' } },
      { id: 'b', text: { id: 'Abaikan dan cek lewat situs atau aplikasi resmi.' } },
      { id: 'c', text: { id: 'Teruskan ke keluarga supaya ikut dapat.' } }
    ],
    answer: 'b',
    explain: {
      id: 'Bantuan resmi tidak meminta OTP lewat tautan di chat. Kode OTP memberi orang lain akses ke akun Anda.'
    }
  },
  {
    id: 'deepfake',
    question: {
      id: 'Video seorang pejabat mengajak warga menanam uang di sebuah aplikasi dengan untung besar. Bibirnya sedikit tidak pas dengan suaranya. Apa yang masuk akal?'
    },
    options: [
      { id: 'a', text: { id: 'Pejabat itu memang mempromosikan, ikut saja.' } },
      { id: 'b', text: { id: 'Kemungkinan deepfake. Cek akun resmi dan berita dulu.' } },
      { id: 'c', text: { id: 'Aman, asal yang membagikan teman sendiri.' } }
    ],
    answer: 'b',
    explain: {
      id: 'Wajah dan suara kini mudah ditiru. Untung besar tanpa risiko adalah tanda penipuan, siapa pun yang tampil di video.'
    }
  },
  {
    id: 'undian',
    question: {
      id: 'Telepon: “Anda menang undian mobil. Transfer Rp 500 ribu dulu untuk pajaknya.” Mana yang benar?'
    },
    options: [
      { id: 'a', text: { id: 'Undian sungguhan tidak meminta bayar di depan, ini penipuan.' } },
      { id: 'b', text: { id: 'Wajar, hadiah memang kena pajak.' } },
      { id: 'c', text: { id: 'Boleh bayar kalau nominalnya kecil.' } }
    ],
    answer: 'a',
    explain: {
      id: 'Biaya kecil di depan adalah umpan. Setelah Anda membayar, akan ada biaya berikutnya.'
    }
  },
  {
    id: 'context',
    question: {
      id: 'Video gempa bernarasi “terjadi di Jakarta tadi malam”, tetapi papan toko dan plat kendaraannya berbahasa dan berformat negara lain. Artinya?'
    },
    options: [
      { id: 'a', text: { id: 'Narasi tidak cocok dengan isi video. Kemungkinan video lama atau dari negara lain.' } },
      { id: 'b', text: { id: 'Tidak penting, yang penting videonya seram.' } },
      { id: 'c', text: { id: 'Pasti buatan AI.' } }
    ],
    answer: 'a',
    explain: {
      id: 'Video asli pun bisa menyesatkan kalau keterangannya salah. Cocokkan tempat, waktu, dan tulisan di gambar dengan klaimnya.'
    }
  },
  {
    id: 'amba',
    question: {
      id: 'AMBA menampilkan “Tidak ditemukan tanda bahaya”. Apakah itu berarti pasti asli?'
    },
    options: [
      { id: 'a', text: { id: 'Ya, pasti asli.' } },
      { id: 'b', text: { id: 'Tidak. Artinya tidak ada tanda menonjol pada yang diperiksa. Untuk kabar penting, tetap cek sumber resmi.' } },
      { id: 'c', text: { id: 'Tidak, berarti pasti palsu.' } }
    ],
    answer: 'b',
    explain: {
      id: 'Hasil AMBA bersifat indikatif. Ia membantu Anda berhenti sejenak, bukan menggantikan sumber resmi.'
    }
  },
  {
    id: 'clue',
    question: {
      id: 'Mana yang paling sering menjadi petunjuk video buatan AI?'
    },
    options: [
      { id: 'a', text: { id: 'Resolusinya tinggi.' } },
      { id: 'b', text: { id: 'Tulisan di papan atau kaos berubah-ubah, dan benda saling menembus.' } },
      { id: 'c', text: { id: 'Durasinya pendek.' } }
    ],
    answer: 'b',
    explain: {
      id: 'Tulisan dan hukum alam adalah titik lemah AI. Resolusi dan durasi tidak membuktikan apa pun.'
    }
  }
];

export const RULES: Card[] = [
  {
    id: 'ask',
    title: { id: 'Minta kata sandi' },
    body: {
      id: 'Kalau ada yang mengaku keluarga lewat nomor baru, minta kata sandi keluarga. Salah jawab berarti bukan mereka.'
    }
  },
  {
    id: 'call',
    title: { id: 'Telepon nomor lama' },
    body: {
      id: 'Jangan pakai nomor di pesan itu. Hubungi nomor yang sudah lama tersimpan di kontak Anda.'
    }
  },
  {
    id: 'secret',
    title: { id: 'Tidak ada rahasia' },
    body: {
      id: 'Permintaan uang atau OTP yang harus disembunyikan dari keluarga adalah tanda bahaya, bukan alasan untuk diam.'
    }
  }
];

/** The safe words are easy-to-say nouns, so a whole family can remember them. */
export const SAFE_WORDS = [
  'Bakso', 'Sepeda', 'Kucing', 'Payung', 'Kopi', 'Bambu', 'Cempedak', 'Layang',
  'Mangga', 'Kereta', 'Bebek', 'Jendela', 'Sapu', 'Kelapa', 'Sepatu', 'Gunting'
];

/** Interface strings of the Belajar page. */
export const UI = {
  title: { id: 'Belajar — AMBA' },
  description: {
    id: 'Cara mengenali video buatan AI dan video hoaks, cara memeriksa sendiri, modus penipuan di Indonesia, dan kuis singkat.'
  },
  h1: { id: 'Belajar mengenali hoaks' },
  lede: {
    id: 'Ketuk satu tanda untuk melihat penjelasannya, lalu coba kuisnya.'
  },
  jump: { id: 'Isi halaman' },
  nav: {
    ai: { id: 'Video AI' },
    hoax: { id: 'Video hoaks' },
    check: { id: 'Cek sendiri' },
    scam: { id: 'Penipuan' },
    quiz: { id: 'Kuis' },
    family: { id: 'Keluarga' }
  },
  ai: {
    title: { id: 'Mengenali video buatan AI' },
    lede: { id: 'Enam hal yang bisa Anda lihat sendiri, tanpa alat apa pun.' },
    note: {
      id: 'Satu tanda belum cukup. AI terus membaik, jadi tanda ini bisa hilang. Gabungkan dengan pemeriksaan sumber.'
    }
  },
  hoax: {
    title: { id: 'Mengenali video hoaks' },
    lede: { id: 'Videonya asli, tetapi ceritanya dibengkokkan.' }
  },
  check: {
    title: { id: 'Cek sendiri' },
    lede: { id: 'Empat langkah, urut dari yang paling cepat.' },
    open: { id: 'Buka' },
    newTab: { id: 'tab baru' }
  },
  scam: {
    title: { id: 'Modus penipuan' },
    lede: { id: 'Polanya mirip: mendesak, menjanjikan, lalu meminta data atau uang.' },
    act: { id: 'Yang dilakukan' },
    rule: {
      id: 'Jangan berikan OTP, PIN, atau kata sandi kepada siapa pun, termasuk yang mengaku petugas bank.'
    }
  },
  quiz: {
    title: { id: 'Kuis singkat' },
    lede: { id: 'Pilih satu jawaban, lalu tekan Jawab. Penjelasan muncul setelah setiap jawaban.' },
    progress: { id: 'Pertanyaan {n} dari {total}' },
    answer: { id: 'Jawab' },
    next: { id: 'Pertanyaan berikutnya' },
    finish: { id: 'Lihat nilai' },
    pick: { id: 'Pilih satu jawaban dulu.' },
    right: { id: 'Jawaban tepat.' },
    wrong: { id: 'Kurang tepat.' },
    correctIs: { id: 'Jawaban yang tepat: {answer}' },
    tagRight: { id: 'Tepat' },
    tagYours: { id: 'Pilihan Anda' },
    scoreTitle: { id: 'Nilai Anda' },
    scoreOf: { id: '{score} dari {total} benar' },
    high: { id: 'Bagus. Tetap ingat: tanda-tanda ini membantu, tetapi sumber resmi yang menentukan.' },
    mid: { id: 'Cukup baik. Baca kartu di atas sekali lagi, lalu coba ulang.' },
    low: { id: 'Belum kuat. Tidak apa-apa. Baca kartu di atas, lalu coba ulang.' },
    again: { id: 'Ulangi kuis' },
    check: { id: 'Periksa tautan sekarang' }
  },
  family: {
    title: { id: 'Kata sandi keluarga' },
    lede: {
      id: 'Penipu bisa meniru suara, foto, dan nomor. Mereka tidak bisa menebak kata yang hanya diketahui keluarga Anda.'
    },
    how: {
      id: 'Dua kata yang mudah diingat, tidak berkaitan dengan nama, alamat, atau tanggal lahir. Simpan di kepala, bukan di chat.'
    },
    none: { id: 'Belum dibuat' },
    make: { id: 'Buat kata sandi' },
    again: { id: 'Buat ulang' },
    copy: { id: 'Salin' },
    copied: { id: 'Tersalin' },
    note: { id: 'Dibuat di perangkat ini saja. AMBA tidak menyimpan atau mengirim kata sandi.' },
    rules: { id: 'Tiga aturan untuk semua orang di rumah' },
    print: { id: 'Cetak kartu' },
    cardBrand: { id: 'AMBA: kartu keluarga' },
    cardLabel: { id: 'Kata sandi keluarga kami' },
    cardFoot: { id: 'Hasil pemeriksaan AMBA bersifat indikatif, bukan bukti pasti.' }
  }
} as const;
