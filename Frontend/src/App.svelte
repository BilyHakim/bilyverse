<script>
  import { onMount, tick } from 'svelte'
  import Admin from './Admin.svelte'
  import ContentList from './ContentList.svelte'
  import { copyCollections, defaultContent } from './content.js'

  let content = $state(defaultContent)
  let collections = $state(copyCollections())
  let revision = -1
  function applyContent(fields, nextRevision, publishedCollections) {
    if (nextRevision < revision) return
    revision = nextRevision
    content = { ...defaultContent, ...fields }
    collections = copyCollections(publishedCollections, fields)
  }

  async function loadContent() {
    try {
      const response = await fetch('/api/content', { signal: AbortSignal.timeout(8000) })
      if (response.ok) {
        const published = await response.json()
        applyContent(published.fields, published.revision, published.collections)
      }
    } catch {
      // Keep the bundled content when the API is offline.
    }
  }

  onMount(loadContent)

  const pages = [
    { id: 'tentang', label: 'Tentang' },
    { id: 'proyek', label: 'Portofolio' },
    { id: 'sertifikat', label: 'Sertifikat' },
    { id: 'blog', label: 'Blog' },
  ]
  const pageIds = [...pages.map(({ id }) => id), 'kontak', 'admin']
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

      <p class="home-interests">{content.interests}</p>

      <div class="home-actions">
        <a class="button home-hubby" href="https://bilyhakim.site/">Buka Hub.by <span aria-hidden="true">&#8599;</span></a>
        <a class="button home-contact" href="#kontak">Hubungi saya</a>
      </div>

      <div class="home-role">
        <p>Saat ini</p>
        <strong>{content.role}</strong>
        <p>{content.company ? 'di ' + content.company : ''}</p>
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
    {#if page === 'admin'}
      <Admin onpublish={loadContent} />
    {:else if page === 'tentang'}
      <section id="tentang" class="section split" aria-labelledby="tentang-judul">
        <h2 id="tentang-judul">Tentang saya</h2>
        <div class="about-copy">
          <p class="published-copy">{content.about}</p>
        </div>
      </section>

    {:else if page === 'proyek'}
      <section id="proyek" class="section projects" aria-labelledby="proyek-judul">
        <h2 id="proyek-judul">Proyek</h2>
        <ContentList entries={collections.projects} empty="Belum ada proyek yang ditampilkan." linkLabel="Buka proyek" layout="projects" />
      </section>

    {:else if page === 'kontak'}
      <section id="kontak" class="section split contact" aria-labelledby="kontak-judul">
        <div>
          <h2 id="kontak-judul">Mari bicara<span class="accent">.</span></h2>
          <p class="published-copy">{content.contactIntro}</p>
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
        {#if page === 'sertifikat'}
          <ContentList entries={collections.certificates} empty="Belum ada sertifikat yang ditampilkan." linkLabel="Lihat kredensial" layout="certificates" />
        {:else}
          <ContentList entries={collections.posts} empty="Belum ada tulisan yang diterbitkan." linkLabel="Buka tautan" />
        {/if}
      </section>
    {/if}
  </main>
{/if}
