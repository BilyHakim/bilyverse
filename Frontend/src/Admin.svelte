<script>
  import { collectionLabels, copyCollections, defaultContent, galleryLimit, siteFields } from './content.js'

  let { onpublish } = $props()
  let key = $state('')
  let draft = $state(null)
  let collections = $state(null)
  let newEntryId = $state('')
  let revision = 0
  let busy = $state(false)
  let status = $state('')
  let error = $state('')
  let working = $state('')
  const imageTypes = ['image/jpeg', 'image/png', 'image/webp']

  async function request(method, body) {
    const response = await fetch('/api/admin/content/', {
      method,
      headers: { 'X-CMS-Key': key, 'Content-Type': 'application/json' },
      body,
      signal: AbortSignal.timeout(12000),
    })
    if (!response.ok) {
      const result = await response.json().catch(() => null)
      throw new Error(result?.error || 'Dashboard belum dapat menghubungi server.')
    }
    return response
  }

  async function submit(event) {
    event.preventDefault()
    if (busy) return
    busy = true
    status = ''
    error = ''
    try {
      if (draft === null) {
        const response = await request('GET')
        const content = await response.json()
        revision = content.revision
        draft = { ...defaultContent, ...content.fields }
        collections = copyCollections(content.collections, draft)
      } else {
        const body = JSON.stringify({ revision, fields: draft, collections })
        if (new Blob([body]).size > 32 * 1024 * 1024) {
          throw new Error('Konten melebihi batas penyimpanan 32 MiB. Kurangi jumlah gambar lalu coba lagi.')
        }
        await request('PUT', body)
        revision += 1
        await onpublish()
        status = 'Perubahan tersimpan. Entri berstatus Terbit sudah tampil di website.'
      }
    } catch (failure) {
      error = failure.message
    } finally {
      busy = false
    }
  }

  function closeEditor() {
    draft = null
    collections = null
    newEntryId = ''
    key = ''
    status = ''
    error = ''
  }

  function addEntry(collection) {
    newEntryId = crypto.randomUUID()
    collections[collection].push({
      id: newEntryId,
      title: '', summary: '', body: '', date: '', url: '', published: false, images: [],
    })
  }

  function removeEntry(collection, id) {
    if (window.confirm('Hapus entri ini dari daftar? Perubahan berlaku setelah disimpan.')) {
      collections[collection] = collections[collection].filter(entry => entry.id !== id)
    }
  }

  function jumpTo(section) {
    const heading = document.getElementById(`cms-${section}`)
    heading.focus({ preventScroll: true })
    heading.scrollIntoView({ block: 'start' })
  }

  function setCover(entry, index) {
    const [image] = entry.images.splice(index, 1)
    entry.images.unshift(image)
  }

  async function selectImages(event, entry) {
    const input = event.currentTarget
    const files = [...input.files]
    input.value = ''
    if (!files.length || busy) return
    error = ''
    status = ''
    if (files.length > galleryLimit - entry.images.length) {
      error = `Satu entri dapat memuat hingga ${galleryLimit} gambar.`
      return
    }

    busy = true
    const failures = []
    try {
      for (const [index, file] of files.entries()) {
        working = `Memproses gambar ${index + 1} dari ${files.length}…`
        if (!imageTypes.includes(file.type) || file.size > 10 * 1024 * 1024) {
          failures.push(`${file.name} bukan JPEG, PNG, atau WebP yang berukuran maksimal 10 MB.`)
          continue
        }
        try {
          const src = await optimizeImage(file)
          entry.images.push({ src, alt: '' })
        } catch {
          failures.push(`${file.name} tidak bisa diproses.`)
        }
      }
      const added = files.length - failures.length
      if (added) status = `${added} gambar ditambahkan. Simpan perubahan untuk menyimpannya.`
      error = failures.join(' ')
    } finally {
      working = ''
      busy = false
    }
  }

  async function optimizeImage(file) {
    const bitmap = await createImageBitmap(file)
    const canvas = document.createElement('canvas')
    const context = canvas.getContext('2d')
    const prefix = 'data:image/jpeg;base64,'
    try {
      for (const size of [1200, 1000, 800, 600, 400]) {
        const scale = Math.min(1, size / Math.max(bitmap.width, bitmap.height))
        canvas.width = Math.max(1, Math.round(bitmap.width * scale))
        canvas.height = Math.max(1, Math.round(bitmap.height * scale))
        context.fillStyle = '#fff'
        context.fillRect(0, 0, canvas.width, canvas.height)
        context.drawImage(bitmap, 0, 0, canvas.width, canvas.height)
        const src = canvas.toDataURL('image/jpeg', .8)
        if (src.startsWith(prefix) && src.length <= prefix.length + 131072) return src
      }
      throw new Error('Gambar terlalu besar setelah dioptimalkan.')
    } finally {
      bitmap.close()
    }
  }
</script>

<section class="section admin contact" class:admin-open={draft !== null} aria-labelledby="admin-title">
  <h2 id="admin-title">Kelola konten</h2>
  <form onsubmit={submit}>
    <fieldset disabled={busy}>
      {#if draft === null}
        <label>Kunci admin<input type="password" autocomplete="current-password" spellcheck="false" bind:value={key} required /></label>
        <button class="button button-primary" type="submit">Buka editor</button>
      {:else}
        <div class="admin-workspace">
          <nav class="admin-nav" aria-label="Bagian konten">
            <button type="button" onclick={() => jumpTo('profile')}>Profil dan kontak</button>
            {#each Object.entries(collectionLabels) as [collection, label]}
              <button type="button" onclick={() => jumpTo(collection)}>
                {label} <span>{collections[collection].length}</span>
              </button>
            {/each}
          </nav>
          <div class="admin-panels">
            <section class="admin-collection" aria-labelledby="cms-profile">
              <h3 id="cms-profile" tabindex="-1">Profil dan kontak</h3>
              {#each siteFields as [field, label, limit]}
                <label>
                  {label}
                  {#if limit > 120}
                    <textarea rows="4" maxlength={limit} bind:value={draft[field]}></textarea>
                  {:else}
                    <input maxlength={limit} bind:value={draft[field]} />
                  {/if}
                </label>
              {/each}
            </section>
            {#each Object.entries(collectionLabels) as [collection, label]}
              <section class="admin-collection" aria-labelledby={'cms-' + collection}>
                <div class="collection-heading">
                  <h3 id={'cms-' + collection} tabindex="-1">{label} <span>({collections[collection].length}/50)</span></h3>
                  <button class="button home-contact" type="button" disabled={collections[collection].length >= 50} onclick={() => addEntry(collection)}>Tambah entri</button>
                </div>
                {#each collections[collection] as entry (entry.id)}
                  <details class="admin-entry" open={entry.id === newEntryId} oninvalidcapture={event => event.currentTarget.open = true}>
                    <summary>{entry.title || 'Entri baru'} <span>{entry.published ? 'Terbit' : 'Draf'}</span></summary>
                    <div class="admin-entry-fields">
                      <label>Judul<input required maxlength="120" bind:value={entry.title} /></label>
                      <div class="admin-gallery">
                        <label>
                          {collection === 'certificates' ? 'Unggah gambar sertifikat' : 'Unggah gambar'} ({entry.images.length}/{galleryLimit})
                          <input type="file" multiple accept={imageTypes.join(',')} disabled={entry.images.length >= galleryLimit} onchange={event => selectImages(event, entry)} />
                          <small>JPEG, PNG, atau WebP. Maksimal 10 MB per file.</small>
                        </label>
                        {#if entry.images.length}
                          <div class="admin-gallery-grid">
                            {#each entry.images as image, index (image)}
                              <div class="admin-image">
                                <img src={image.src} alt={image.alt || ''} loading="lazy" />
                                <label>Teks alternatif gambar {index + 1}<input maxlength="180" bind:value={image.alt} /></label>
                                <div class="admin-image-actions">
                                  {#if index === 0}
                                    <span>Sampul</span>
                                  {:else}
                                    <button type="button" aria-label={'Jadikan gambar ' + (index + 1) + ' sampul'} onclick={() => setCover(entry, index)}>Jadikan sampul</button>
                                  {/if}
                                  <button class="remove-entry" type="button" aria-label={'Hapus gambar ' + (index + 1)} onclick={() => entry.images.splice(index, 1)}>Hapus</button>
                                </div>
                              </div>
                            {/each}
                          </div>
                        {/if}
                      </div>
                      <label>{collection === 'certificates' ? 'Penerbit atau ringkasan' : 'Ringkasan'}<textarea rows="2" maxlength="2000" bind:value={entry.summary}></textarea></label>
                      <label>{collection === 'posts' ? 'Isi artikel' : 'Detail'}<textarea rows="4" maxlength="12000" bind:value={entry.body}></textarea></label>
                      <div class="form-row">
                        <label>Tanggal<input type="date" bind:value={entry.date} /></label>
                        <label>{collection === 'certificates' ? 'URL kredensial' : 'URL terkait'}<input type="url" maxlength="2048" placeholder="https://" bind:value={entry.url} /></label>
                      </div>
                      <div class="collection-heading">
                        <label class="publish-entry"><input type="checkbox" bind:checked={entry.published} /> Tampilkan di website</label>
                        <button class="button home-contact remove-entry" type="button" onclick={() => removeEntry(collection, entry.id)}>Hapus entri</button>
                      </div>
                    </div>
                  </details>
                {:else}
                  <p>Belum ada entri.</p>
                {/each}
              </section>
            {/each}
          </div>
        </div>
      {/if}
    </fieldset>
    <div class:admin-toolbar={draft !== null}>
      {#if draft !== null}
        <div class="admin-actions">
          <button class="button button-primary" type="submit" disabled={busy}>Simpan perubahan</button>
          <button class="button home-contact" type="button" disabled={busy} onclick={closeEditor}>Tutup editor</button>
        </div>
      {/if}
      <p class="form-status" role="status">{busy ? working || 'Sedang memproses…' : status}</p>
      <p class="form-error" role="alert">{error}</p>
    </div>
  </form>
</section>
