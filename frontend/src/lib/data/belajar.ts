/**
 * Copy for the Belajar page (/edukasi): short cards and the quiz.
 * Plain words for laypeople and the elderly. Both languages live side by side
 * so a card and its translation are edited together.
 */
export type Lang = 'id' | 'en';
export type Text = Record<Lang, string>;

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
    title: { id: 'Tangan dan jari', en: 'Hands and fingers' },
    body: {
      id: 'Jumlah jari tidak pas, jari menyatu, atau cara memegang benda terlihat aneh. AI masih sering salah di bagian ini.',
      en: 'The wrong number of fingers, fingers that merge, or an odd grip on an object. AI still often gets this wrong.'
    }
  },
  {
    id: 'text',
    title: { id: 'Tulisan', en: 'Writing' },
    body: {
      id: 'Huruf di papan, kaos, atau layar berubah-ubah, buram, atau tidak membentuk kata yang masuk akal.',
      en: 'Letters on signs, shirts or screens keep changing, look blurred, or do not form real words.'
    }
  },
  {
    id: 'physics',
    title: { id: 'Hukum alam', en: 'Physics' },
    body: {
      id: 'Air, asap, kain, dan rambut bergerak tidak wajar. Benda menembus satu sama lain, muncul, atau hilang.',
      en: 'Water, smoke, cloth and hair move unnaturally. Objects pass through each other, appear, or vanish.'
    }
  },
  {
    id: 'light',
    title: { id: 'Cahaya dan bayangan', en: 'Light and shadow' },
    body: {
      id: 'Bayangan jatuh ke arah yang berbeda-beda. Pantulan di kaca atau mata tidak cocok dengan sekitarnya.',
      en: 'Shadows fall in different directions. Reflections in glass or eyes do not match the surroundings.'
    }
  },
  {
    id: 'lips',
    title: { id: 'Bibir dan suara', en: 'Lips and voice' },
    body: {
      id: 'Gerak bibir tidak pas dengan ucapan. Kedipan jarang atau kaku. Kulit wajah terlalu halus.',
      en: 'Lip movement does not match the words. Blinking is rare or stiff. Facial skin looks too smooth.'
    }
  },
  {
    id: 'smooth',
    title: { id: 'Gerakan terlalu sempurna', en: 'Too-perfect motion' },
    body: {
      id: 'Kamera meluncur mulus, orang bergerak seperti di iklan, tanpa goyangan kecil seperti video dari ponsel biasa.',
      en: 'The camera glides, people move like in an advert, with none of the small shake of an ordinary phone video.'
    }
  }
];

export const HOAX_CUES: Card[] = [
  {
    id: 'old',
    title: { id: 'Video lama, keterangan baru', en: 'Old footage, new caption' },
    body: {
      id: 'Rekaman asli dari tahun atau tempat lain, diberi keterangan seolah baru terjadi. Banjir, kebakaran, dan kerusuhan sering dipakai.',
      en: 'Real footage from another year or place, captioned as if it just happened. Floods, fires and riots are common.'
    }
  },
  {
    id: 'cut',
    title: { id: 'Dipotong menyesatkan', en: 'Cut to mislead' },
    body: {
      id: 'Hanya sebagian adegan yang ditampilkan, atau kalimat dipotong sehingga artinya berbalik dari yang sebenarnya.',
      en: 'Only part of the scene is shown, or a sentence is cut so its meaning flips.'
    }
  },
  {
    id: 'voice',
    title: { id: 'Suara atau narasi palsu', en: 'Fake voice or voice-over' },
    body: {
      id: 'Suara tokoh ditiru, atau narasi baru ditempel di video peristiwa lain. Cocokkan isi ucapan dengan gambarnya.',
      en: 'A person’s voice is imitated, or a new voice-over is laid on footage of another event. Match what is said to what is shown.'
    }
  },
  {
    id: 'headline',
    title: { id: 'Tangkapan layar berita diedit', en: 'Edited news screenshot' },
    body: {
      id: 'Judul diubah agar terdengar lebih mengejutkan. Buka berita aslinya di situs resminya sebelum percaya.',
      en: 'The headline is changed to sound more shocking. Open the original article on its own site before believing it.'
    }
  }
];

export const CHECK_STEPS: CheckStep[] = [
  {
    id: 'tbh',
    title: { id: 'TurnBackHoax', en: 'TurnBackHoax' },
    body: {
      id: 'Ketik kata kunci kabar itu. Situs ini dikelola Mafindo dan memuat ribuan kabar yang sudah diperiksa.',
      en: 'Search the key words of the claim. Run by Mafindo, it holds thousands of already-checked claims.'
    },
    link: { href: 'https://turnbackhoax.id', label: 'turnbackhoax.id' }
  },
  {
    id: 'cekfakta',
    title: { id: 'cekfakta.com', en: 'cekfakta.com' },
    body: {
      id: 'Kerja sama banyak media di Indonesia. Cari kabar yang sama di sini untuk membandingkan hasil.',
      en: 'A joint effort by many Indonesian media. Search for the same claim here to compare.'
    },
    link: { href: 'https://cekfakta.com', label: 'cekfakta.com' }
  },
  {
    id: 'reverse',
    title: { id: 'Cari gambar terbalik', en: 'Reverse image search' },
    body: {
      id: 'Tangkap layar satu adegan video, lalu unggah ke Google Lens. Hasilnya sering menunjukkan kapan dan di mana gambar itu pertama muncul.',
      en: 'Take a screenshot of one scene and upload it to Google Lens. It often shows when and where the image first appeared.'
    },
    link: { href: 'https://lens.google.com', label: 'lens.google.com' }
  },
  {
    id: 'official',
    title: { id: 'Akun dan situs resmi', en: 'Official accounts and sites' },
    body: {
      id: 'Kalau kabar menyebut instansi atau tokoh, cek pengumuman di akun resmi mereka. Jangan lewat tautan di pesan.',
      en: 'If a claim names an agency or a person, check the announcement on their official account. Do not use a link from the message.'
    }
  }
];

export const SCAMS: Scam[] = [
  {
    id: 'bansos',
    title: { id: 'Bansos palsu', en: 'Fake social aid' },
    body: {
      id: 'Pesan bilang Anda terpilih menerima bantuan, lalu meminta Anda mengisi data KTP atau memasukkan OTP lewat tautan.',
      en: 'A message says you were picked for aid, then asks for your ID data or an OTP through a link.'
    },
    act: {
      id: 'Abaikan. Cek hak bantuan hanya lewat situs atau aplikasi resmi.',
      en: 'Ignore it. Check aid eligibility only on the official site or app.'
    }
  },
  {
    id: 'undian',
    title: { id: 'Undian dan hadiah', en: 'Prize draws' },
    body: {
      id: 'Anda menang mobil atau uang, tetapi harus transfer biaya pajak atau administrasi lebih dulu.',
      en: 'You won a car or cash, but must first transfer a tax or admin fee.'
    },
    act: {
      id: 'Undian sungguhan tidak meminta bayar di depan. Putuskan kontak.',
      en: 'A real prize never asks for payment up front. Cut contact.'
    }
  },
  {
    id: 'petugas',
    title: { id: 'Petugas atau pejabat palsu', en: 'Fake officials' },
    body: {
      id: 'Mengaku dari bank, kurir, polisi, atau kantor pemerintah. Biasanya mendesak dan menakut-nakuti supaya Anda panik.',
      en: 'Claims to be from a bank, courier, police or a government office. Usually pushes and frightens you into panic.'
    },
    act: {
      id: 'Tutup telepon, lalu hubungi nomor resmi yang Anda cari sendiri.',
      en: 'Hang up, then call the official number you look up yourself.'
    }
  },
  {
    id: 'deepfake',
    title: { id: 'Video pejabat hasil deepfake', en: 'Deepfake videos of officials' },
    body: {
      id: 'Wajah dan suara tokoh ditiru untuk mengajak berinvestasi, menyumbang, atau mengklaim hadiah.',
      en: 'A public figure’s face and voice are copied to urge you to invest, donate, or claim a prize.'
    },
    act: {
      id: 'Cek di akun resmi tokoh itu dan di berita. Jangan percaya satu video saja.',
      en: 'Check the figure’s official account and the news. Do not trust a single video.'
    }
  }
];

export const QUESTIONS: Question[] = [
  {
    id: 'hands',
    question: {
      id: 'Di sebuah video, tangan seseorang punya enam jari dan jarinya menyatu saat memegang gelas. Apa yang paling mungkin?',
      en: 'In a video a person’s hand has six fingers that merge as they hold a glass. What is most likely?'
    },
    options: [
      { id: 'a', text: { id: 'Kamera ponselnya kurang bagus.', en: 'The phone camera is poor.' } },
      { id: 'b', text: { id: 'Video itu kemungkinan dibuat atau diubah dengan AI.', en: 'The video was probably made or altered with AI.' } },
      { id: 'c', text: { id: 'Video itu pasti asli karena terlihat jelas.', en: 'The video is surely real because it looks clear.' } }
    ],
    answer: 'b',
    explain: {
      id: 'AI masih sering salah menggambar tangan. Ini bukan bukti pasti, tetapi alasan kuat untuk memeriksa sumbernya.',
      en: 'AI still often draws hands wrongly. It is not proof, but a strong reason to check the source.'
    }
  },
  {
    id: 'old',
    question: {
      id: 'Video banjir besar beredar dengan keterangan “terjadi hari ini di kota kita”. Langkah pertama yang paling tepat?',
      en: 'A video of a big flood is shared with the caption “happened today in our city”. Best first step?'
    },
    options: [
      { id: 'a', text: { id: 'Kirim ke grup supaya semua waspada.', en: 'Send it to the group so everyone is alert.' } },
      { id: 'b', text: { id: 'Cari kapan dan di mana video itu pertama muncul.', en: 'Find out when and where the video first appeared.' } },
      { id: 'c', text: { id: 'Percaya karena banyak yang membagikan.', en: 'Believe it because many people share it.' } }
    ],
    answer: 'b',
    explain: {
      id: 'Video lama sering dipasangi keterangan baru. Pencarian gambar terbalik biasanya menunjukkan kemunculan pertamanya.',
      en: 'Old videos often get a new caption. A reverse image search usually shows the first time it appeared.'
    }
  },
  {
    id: 'bansos',
    question: {
      id: 'Pesan WhatsApp: “Anda terpilih menerima bansos Rp 2.000.000. Klik tautan, isi data KTP dan kode OTP.” Sikap yang tepat?',
      en: 'A WhatsApp message: “You were picked for Rp 2,000,000 aid. Click the link, enter your ID data and OTP.” Right response?'
    },
    options: [
      { id: 'a', text: { id: 'Isi secepatnya supaya tidak hangus.', en: 'Fill it in fast so it does not expire.' } },
      { id: 'b', text: { id: 'Abaikan dan cek lewat situs atau aplikasi resmi.', en: 'Ignore it and check on the official site or app.' } },
      { id: 'c', text: { id: 'Teruskan ke keluarga supaya ikut dapat.', en: 'Forward it to family so they get it too.' } }
    ],
    answer: 'b',
    explain: {
      id: 'Bantuan resmi tidak meminta OTP lewat tautan di chat. Kode OTP memberi orang lain akses ke akun Anda.',
      en: 'Official aid never asks for an OTP through a chat link. An OTP gives others access to your account.'
    }
  },
  {
    id: 'deepfake',
    question: {
      id: 'Video seorang pejabat mengajak warga menanam uang di sebuah aplikasi dengan untung besar. Bibirnya sedikit tidak pas dengan suaranya. Apa yang masuk akal?',
      en: 'A video of an official urges people to put money into an app for big profit. The lips do not quite match the voice. What makes sense?'
    },
    options: [
      { id: 'a', text: { id: 'Pejabat itu memang mempromosikan, ikut saja.', en: 'The official really promotes it, join in.' } },
      { id: 'b', text: { id: 'Kemungkinan deepfake. Cek akun resmi dan berita dulu.', en: 'Likely a deepfake. Check the official account and the news first.' } },
      { id: 'c', text: { id: 'Aman, asal yang membagikan teman sendiri.', en: 'Safe, as long as a friend shared it.' } }
    ],
    answer: 'b',
    explain: {
      id: 'Wajah dan suara kini mudah ditiru. Untung besar tanpa risiko adalah tanda penipuan, siapa pun yang tampil di video.',
      en: 'Faces and voices are easy to copy now. Big profit without risk is a scam sign, whoever appears in the video.'
    }
  },
  {
    id: 'undian',
    question: {
      id: 'Telepon: “Anda menang undian mobil. Transfer Rp 500 ribu dulu untuk pajaknya.” Mana yang benar?',
      en: 'A call: “You won a car in a draw. Transfer Rp 500,000 first for the tax.” Which is right?'
    },
    options: [
      { id: 'a', text: { id: 'Undian sungguhan tidak meminta bayar di depan, ini penipuan.', en: 'A real draw never asks for money up front. This is a scam.' } },
      { id: 'b', text: { id: 'Wajar, hadiah memang kena pajak.', en: 'Normal, prizes are taxed.' } },
      { id: 'c', text: { id: 'Boleh bayar kalau nominalnya kecil.', en: 'Fine to pay if the amount is small.' } }
    ],
    answer: 'a',
    explain: {
      id: 'Biaya kecil di depan adalah umpan. Setelah Anda membayar, akan ada biaya berikutnya.',
      en: 'A small fee up front is the bait. After you pay, more fees follow.'
    }
  },
  {
    id: 'context',
    question: {
      id: 'Video gempa bernarasi “terjadi di Jakarta tadi malam”, tetapi papan toko dan plat kendaraannya berbahasa dan berformat negara lain. Artinya?',
      en: 'A quake video says “happened in Jakarta last night”, but shop signs and number plates are from another country. What does it mean?'
    },
    options: [
      { id: 'a', text: { id: 'Narasi tidak cocok dengan isi video. Kemungkinan video lama atau dari negara lain.', en: 'The caption does not match the video. Probably old or from another country.' } },
      { id: 'b', text: { id: 'Tidak penting, yang penting videonya seram.', en: 'Does not matter, the video is scary enough.' } },
      { id: 'c', text: { id: 'Pasti buatan AI.', en: 'Surely made by AI.' } }
    ],
    answer: 'a',
    explain: {
      id: 'Video asli pun bisa menyesatkan kalau keterangannya salah. Cocokkan tempat, waktu, dan tulisan di gambar dengan klaimnya.',
      en: 'A real video can still mislead if the caption is wrong. Match place, time and writing in the picture to the claim.'
    }
  },
  {
    id: 'amba',
    question: {
      id: 'AMBA menampilkan “Tidak ditemukan tanda bahaya”. Apakah itu berarti pasti asli?',
      en: 'AMBA shows “No danger signs found”. Does that mean it is certainly genuine?'
    },
    options: [
      { id: 'a', text: { id: 'Ya, pasti asli.', en: 'Yes, certainly genuine.' } },
      { id: 'b', text: { id: 'Tidak. Artinya tidak ada tanda menonjol pada yang diperiksa. Untuk kabar penting, tetap cek sumber resmi.', en: 'No. It means nothing stood out in what was checked. For important news, still check an official source.' } },
      { id: 'c', text: { id: 'Tidak, berarti pasti palsu.', en: 'No, it means certainly fake.' } }
    ],
    answer: 'b',
    explain: {
      id: 'Hasil AMBA bersifat indikatif. Ia membantu Anda berhenti sejenak, bukan menggantikan sumber resmi.',
      en: 'AMBA’s result is indicative. It helps you pause, it does not replace an official source.'
    }
  },
  {
    id: 'clue',
    question: {
      id: 'Mana yang paling sering menjadi petunjuk video buatan AI?',
      en: 'Which is most often a clue that a video is AI-made?'
    },
    options: [
      { id: 'a', text: { id: 'Resolusinya tinggi.', en: 'It is high resolution.' } },
      { id: 'b', text: { id: 'Tulisan di papan atau kaos berubah-ubah, dan benda saling menembus.', en: 'Writing on signs or shirts keeps changing, and objects pass through each other.' } },
      { id: 'c', text: { id: 'Durasinya pendek.', en: 'It is short.' } }
    ],
    answer: 'b',
    explain: {
      id: 'Tulisan dan hukum alam adalah titik lemah AI. Resolusi dan durasi tidak membuktikan apa pun.',
      en: 'Writing and physics are AI’s weak spots. Resolution and length prove nothing.'
    }
  }
];

export const RULES: Card[] = [
  {
    id: 'ask',
    title: { id: 'Minta kata sandi', en: 'Ask for the word' },
    body: {
      id: 'Kalau ada yang mengaku keluarga lewat nomor baru, minta kata sandi keluarga. Salah jawab berarti bukan mereka.',
      en: 'If someone claims to be family from a new number, ask for the family word. A wrong answer means it is not them.'
    }
  },
  {
    id: 'call',
    title: { id: 'Telepon nomor lama', en: 'Call the old number' },
    body: {
      id: 'Jangan pakai nomor di pesan itu. Hubungi nomor yang sudah lama tersimpan di kontak Anda.',
      en: 'Do not use the number in the message. Call the one long saved in your contacts.'
    }
  },
  {
    id: 'secret',
    title: { id: 'Tidak ada rahasia', en: 'No secrets' },
    body: {
      id: 'Permintaan uang atau OTP yang harus disembunyikan dari keluarga adalah tanda bahaya, bukan alasan untuk diam.',
      en: 'A request for money or an OTP that must be hidden from family is a danger sign, not a reason to stay quiet.'
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
  title: { id: 'Belajar — AMBA', en: 'Learn — AMBA' },
  description: {
    id: 'Cara mengenali video buatan AI dan video hoaks, cara memeriksa sendiri, modus penipuan di Indonesia, dan kuis singkat.',
    en: 'How to spot AI-made and hoax videos, how to check for yourself, scam patterns in Indonesia, and a short quiz.'
  },
  h1: { id: 'Belajar mengenali video palsu dan hoaks', en: 'Learn to spot fake videos and hoaxes' },
  lede: {
    id: 'Kartu singkat, cara memeriksa sendiri, dan kuis delapan pertanyaan. Tidak ada nilai yang disimpan.',
    en: 'Short cards, ways to check for yourself, and an eight-question quiz. No score is saved.'
  },
  jump: { id: 'Isi halaman', en: 'On this page' },
  nav: {
    ai: { id: 'Video AI', en: 'AI video' },
    hoax: { id: 'Video hoaks', en: 'Hoax video' },
    check: { id: 'Cek sendiri', en: 'Check yourself' },
    scam: { id: 'Penipuan', en: 'Scams' },
    quiz: { id: 'Kuis', en: 'Quiz' },
    family: { id: 'Keluarga', en: 'Family' }
  },
  ai: {
    title: { id: 'Mengenali video buatan AI', en: 'Spotting AI-made video' },
    lede: { id: 'Enam hal yang bisa Anda lihat sendiri, tanpa alat apa pun.', en: 'Six things you can see yourself, with no tools.' },
    note: {
      id: 'Satu tanda belum cukup. AI terus membaik, jadi tanda ini bisa hilang. Gabungkan dengan pemeriksaan sumber.',
      en: 'One sign is not enough. AI keeps improving, so these signs can fade. Combine them with a source check.'
    }
  },
  hoax: {
    title: { id: 'Mengenali video hoaks', en: 'Spotting hoax video' },
    lede: { id: 'Videonya asli, tetapi ceritanya dibengkokkan.', en: 'The video is real, but the story is bent.' }
  },
  check: {
    title: { id: 'Cara memeriksa sendiri', en: 'How to check for yourself' },
    lede: { id: 'Empat langkah, urut dari yang paling cepat.', en: 'Four steps, from the quickest.' },
    open: { id: 'Buka', en: 'Open' },
    newTab: { id: 'tab baru', en: 'new tab' }
  },
  scam: {
    title: { id: 'Modus penipuan yang sering terjadi di Indonesia', en: 'Common scam patterns in Indonesia' },
    lede: { id: 'Polanya mirip: mendesak, menjanjikan, lalu meminta data atau uang.', en: 'The pattern is alike: urgent, tempting, then asking for data or money.' },
    act: { id: 'Yang dilakukan', en: 'What to do' },
    rule: {
      id: 'Jangan berikan OTP, PIN, atau kata sandi kepada siapa pun, termasuk yang mengaku petugas bank.',
      en: 'Never give an OTP, PIN or password to anyone, including someone who says they are bank staff.'
    }
  },
  quiz: {
    title: { id: 'Kuis singkat', en: 'Short quiz' },
    lede: { id: 'Pilih satu jawaban, lalu tekan Jawab. Penjelasan muncul setelah setiap jawaban.', en: 'Pick one answer, then press Answer. An explanation follows each answer.' },
    progress: { id: 'Pertanyaan {n} dari {total}', en: 'Question {n} of {total}' },
    answer: { id: 'Jawab', en: 'Answer' },
    next: { id: 'Pertanyaan berikutnya', en: 'Next question' },
    finish: { id: 'Lihat nilai', en: 'See score' },
    pick: { id: 'Pilih satu jawaban dulu.', en: 'Pick an answer first.' },
    right: { id: 'Jawaban tepat.', en: 'Correct.' },
    wrong: { id: 'Kurang tepat.', en: 'Not quite.' },
    correctIs: { id: 'Jawaban yang tepat: {answer}', en: 'The right answer: {answer}' },
    tagRight: { id: 'Tepat', en: 'Correct' },
    tagYours: { id: 'Pilihan Anda', en: 'Your choice' },
    scoreTitle: { id: 'Nilai Anda', en: 'Your score' },
    scoreOf: { id: '{score} dari {total} benar', en: '{score} of {total} correct' },
    high: { id: 'Bagus. Tetap ingat: tanda-tanda ini membantu, tetapi sumber resmi yang menentukan.', en: 'Good. Remember: these signs help, but official sources decide.' },
    mid: { id: 'Cukup baik. Baca kartu di atas sekali lagi, lalu coba ulang.', en: 'Fair. Read the cards above once more, then try again.' },
    low: { id: 'Belum kuat. Tidak apa-apa. Baca kartu di atas, lalu coba ulang.', en: 'Not yet. That is fine. Read the cards above, then try again.' },
    again: { id: 'Ulangi kuis', en: 'Retake quiz' },
    check: { id: 'Periksa tautan sekarang', en: 'Check a link now' }
  },
  family: {
    title: { id: 'Kata sandi keluarga', en: 'Family safe word' },
    lede: {
      id: 'Penipu bisa meniru suara, foto, dan nomor. Mereka tidak bisa menebak kata yang hanya diketahui keluarga Anda.',
      en: 'Scammers can imitate voices, photos and numbers. They cannot guess a word only your family knows.'
    },
    how: {
      id: 'Dua kata yang mudah diingat, tidak berkaitan dengan nama, alamat, atau tanggal lahir. Simpan di kepala, bukan di chat.',
      en: 'Two easy words, unrelated to names, addresses or birthdays. Keep it in your head, not in chat.'
    },
    none: { id: 'Belum dibuat', en: 'Not made yet' },
    make: { id: 'Buat kata sandi', en: 'Make a word' },
    again: { id: 'Buat ulang', en: 'Make another' },
    copy: { id: 'Salin', en: 'Copy' },
    copied: { id: 'Tersalin', en: 'Copied' },
    note: { id: 'Dibuat di perangkat ini saja. AMBA tidak menyimpan atau mengirim kata sandi.', en: 'Made on this device only. AMBA does not store or send the word.' },
    rules: { id: 'Tiga aturan untuk semua orang di rumah', en: 'Three rules for everyone at home' },
    print: { id: 'Cetak kartu', en: 'Print card' },
    cardBrand: { id: 'AMBA: kartu keluarga', en: 'AMBA: family card' },
    cardLabel: { id: 'Kata sandi keluarga kami', en: 'Our family word' },
    cardFoot: { id: 'Hasil pemeriksaan AMBA bersifat indikatif, bukan bukti pasti.', en: 'AMBA results are indicative, not proof.' }
  }
} as const;
