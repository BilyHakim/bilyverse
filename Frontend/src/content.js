export const contentFields = [
  ["interests","Bidang minat",100,"Business Intelligence · IT Support · Full Stack Development · Machine Learning"],
  ["role","Posisi",36,"Backend Developer"],
  ["company","Perusahaan",45,"PT. Supranusa Sindata"],
  ["about","Tentang saya",6000,"Backend adalah fokus utama saya. Di luar itu, saya juga tertarik pada pengembangan full-stack, business intelligence, IT support, dan machine learning.\n\nUntuk portofolio ini, saya ingin memperlihatkan bagaimana antarmuka, API, dan database bekerja bersama. Alurnya bisa dicoba langsung lewat formulir kontak."],
  ["projectTitle","Judul proyek",120,"Bilyverse"],
  ["projectSummary","Ringkasan proyek",2000,"Portofolio pribadi dengan formulir kontak yang menyimpan pesan ke PostgreSQL."],
  ["projectDetails","Detail proyek",6000,"Svelte mengirim data formulir ke POST /api/contact. Handler Chi memvalidasi masukan sebelum menyimpan pesan ke PostgreSQL. Migrasi Goose mengatur skema tabel pesan."],
  ["certificates","Sertifikat",6000,""],
  ["blog","Blog",12000,""],
  ["contactIntro","Pengantar kontak",1000,"Punya pertanyaan tentang proyek ini atau ingin berdiskusi? Kirim pesan lewat formulir ini."],
]

export const defaultContent = Object.fromEntries(contentFields.map(([key, , , value]) => [key, value]))
