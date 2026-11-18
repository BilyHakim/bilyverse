<script>
  import { tick } from 'svelte'

  const pages = [
    { id: 'tentang', label: 'Tentang' },
    { id: 'proyek', label: 'Portofolio' },
    { id: 'sertifikat', label: 'Sertifikat' },
    { id: 'blog', label: 'Blog' },
  ]
  const pageIds = [...pages.map(({ id }) => id), 'kontak']
  function currentPage() {
    const hash = window.location.hash.slice(1)
    return pageIds.includes(hash) ? hash : 'home'
  }

  let page = $state(currentPage())
  let main = $state()

  async function navigate() {
    const nextPage = currentPage()
    if (nextPage === page) return
    page = nextPage
    await tick()
    window.scrollTo(0, 0)
    main?.focus({ preventScroll: true })
  }

  const stack = ['Svelte', 'Go / Chi', 'PostgreSQL', 'Goose']
  let pending = $state(false)
  let status = $state('')
  let error = $state('')

  async function sendMessage(event) {
    event.preventDefault()
    if (pending) return

    const form = event.currentTarget
    const message = Object.fromEntries(new FormData(form))
    pending = true
    status = ''
    error = ''

    try {
      const response = await fetch('/api/contact', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(message),
        signal: AbortSignal.timeout(12000),
      })
      if (!response.ok) {
        if (response.status === 400) {
          const result = await response.json()
          error = result.error
          return
        }
        throw new Error('Failed to save contact message')
      }
      form.reset()
      status = 'Pesan sudah terkirim. Terima kasih!'
    } catch {
      error = 'Pesan belum terkirim. Silakan coba lagi.'
    } finally {
      pending = false
    }
  }
</script>

<svelte:window onhashchange={navigate} />

{#if page === 'home'}
  <main class="home" tabindex="-1" bind:this={main}>
    <div class="home-content">
      <h1 aria-label="Bily Hakim Erlangga">
        <span class="home-name-mobile" aria-hidden="true">Bily Hakim<br />Erlangga<span class="accent">.</span></span>
        <svg viewBox="0 0 1000 215" aria-hidden="true">
          <path id="name-arc" d="M 28,195 Q 500,-104 972,195" fill="none" />
          <text text-anchor="middle" fill="currentColor">
            <textPath href="#name-arc" startOffset="50%">BILY HAKIM ERLANGGA</textPath>
          </text>
        </svg>
      </h1>

      <nav class="home-nav" aria-label="Navigasi utama">
        {#each pages as { id, label }}
          <a href={'#' + id}>{label}</a>
        {/each}
      </nav>

      <p class="home-interests">Business Intelligence &middot; IT Support &middot; Full Stack Development &middot; Machine Learning</p>

      <a class="button home-contact" href="#kontak">Hubungi saya <span aria-hidden="true">&#8599;</span></a>

      <div class="home-role">
        <p>Saat ini</p>
        <strong>Backend Developer</strong>
        <p>di PT. Supranusa Sindata</p>
      </div>
    </div>
  </main>
{:else}
  <header class="site-header">
    <div class="shell header-inner">
      <a class="brand" href="#home">Bilyverse<span class="accent">.</span></a>
      <nav aria-label="Navigasi utama">
        {#each pages as { id, label }}
          <a href={'#' + id} aria-current={page === id ? 'page' : undefined}>{label}</a>
        {/each}
        <a href="#kontak" aria-current={page === 'kontak' ? 'page' : undefined}>Kontak</a>
      </nav>
    </div>
  </header>

  <main class="shell" tabindex="-1" bind:this={main}>
    {#if page === 'tentang'}
      <section id="tentang" class="section split" aria-labelledby="tentang-judul">
        <h2 id="tentang-judul">Tentang saya</h2>
        <div class="about-copy">
          <p>Backend adalah fokus utama saya. Di luar itu, saya juga tertarik pada pengembangan full-stack, business intelligence, IT support, dan machine learning.</p>
          <p>Untuk portofolio ini, saya ingin memperlihatkan bagaimana antarmuka, API, dan database bekerja bersama. Alurnya bisa dicoba langsung lewat formulir kontak.</p>
        </div>
      </section>

    {:else if page === 'proyek'}
      <section id="proyek" class="section projects" aria-labelledby="proyek-judul">
        <h2 id="proyek-judul">Proyek</h2>
        <article class="project-card">
          <div class="project-body">
            <p class="eyebrow">Portofolio pribadi</p>
            <h3>Bilyverse</h3>
            <p>Portofolio pribadi dengan formulir kontak yang menyimpan pesan ke PostgreSQL.</p>
            <ul class="stack" aria-label="Teknologi proyek">
              {#each stack as technology}
                <li>{technology}</li>
              {/each}
            </ul>
          </div>
          <details class="project-details">
            <summary>Detail teknis <span aria-hidden="true"></span></summary>
            <p>Svelte mengirim data formulir ke <code>POST /api/contact</code>. Handler Chi memvalidasi masukan sebelum menyimpan pesan ke PostgreSQL. Migrasi Goose mengatur skema tabel pesan.</p>
          </details>
        </article>
      </section>

    {:else if page === 'kontak'}
      <section id="kontak" class="section split contact" aria-labelledby="kontak-judul">
        <div>
          <h2 id="kontak-judul">Mari bicara<span class="accent">.</span></h2>
          <p>Punya pertanyaan tentang proyek ini atau ingin berdiskusi? Kirim pesan lewat formulir ini.</p>
        </div>
        <form onsubmit={sendMessage}>
          <div class="form-row">
            <label>Nama<input name="name" autocomplete="name" maxlength="100" required /></label>
            <label>Email<input name="email" type="email" autocomplete="email" maxlength="254" required /></label>
          </div>
          <label>Pesan<textarea name="message" rows="4" maxlength="3000" required></textarea></label>
          <button class="button button-primary" disabled={pending} type="submit">
            {pending ? 'Mengirim…' : 'Kirim pesan'} <span aria-hidden="true">↗</span>
          </button>
          <p class="form-status" role="status">{status}</p>
          <p class="form-error" role="alert">{error}</p>
        </form>
      </section>
    {:else}
      <section class="section">
        <h2>{page === 'sertifikat' ? 'Sertifikat' : 'Blog'}</h2>
        <p class="empty-page">{page === 'sertifikat' ? 'Belum ada sertifikat yang ditampilkan.' : 'Belum ada tulisan yang diterbitkan.'}</p>
      </section>
    {/if}
  </main>
{/if}
