<script>
  import { collectionLabels, copyCollections, defaultContent, siteFields } from './content.js'

  let { onpublish } = $props()
  let key = $state('')
  let draft = $state(null)
  let collections = $state(null)
  let newEntryId = $state('')
  let revision = 0
  let busy = $state(false)
  let status = $state('')
  let error = $state('')

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
        await request('PUT', JSON.stringify({ revision, fields: draft, collections }))
        revision += 1
        onpublish({ ...draft }, revision, copyCollections(collections))
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
      title: '', summary: '', body: '', date: '', url: '', published: false,
    })
  }

  function removeEntry(collection, id) {
    if (window.confirm('Hapus entri ini dari daftar? Perubahan berlaku setelah disimpan.')) {
      collections[collection] = collections[collection].filter(entry => entry.id !== id)
    }
  }
</script>

<section class="section admin contact" aria-labelledby="admin-title">
  <h2 id="admin-title">Kelola konten</h2>
  <form onsubmit={submit}>
    <fieldset disabled={busy}>
      {#if draft === null}
        <label>Kunci admin<input type="password" autocomplete="current-password" spellcheck="false" bind:value={key} required /></label>
        <button class="button button-primary" type="submit">Buka editor</button>
      {:else}
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
        {#each Object.entries(collectionLabels) as [collection, label]}
          <section class="admin-collection" aria-label={label}>
            <div class="collection-heading">
              <h3>{label} <span>({collections[collection].length}/50)</span></h3>
              <button class="button home-contact" type="button" disabled={collections[collection].length >= 50} onclick={() => addEntry(collection)}>Tambah entri</button>
            </div>
            {#each collections[collection] as entry (entry.id)}
              <details class="admin-entry" open={entry.id === newEntryId} oninvalidcapture={event => event.currentTarget.open = true}>
                <summary>{entry.title || 'Entri baru'} <span>{entry.published ? 'Terbit' : 'Draf'}</span></summary>
                <div class="admin-entry-fields">
                  <label>Judul<input required maxlength="120" bind:value={entry.title} /></label>
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
        <div class="admin-actions">
          <button class="button button-primary" type="submit">Simpan perubahan</button>
          <button class="button home-contact" type="button" onclick={closeEditor}>Tutup editor</button>
        </div>
      {/if}
    </fieldset>
    {#if busy}<p role="status">Sedang memproses…</p>{/if}
    {#if status}<p class="form-status" role="status">{status}</p>{/if}
    {#if error}<p class="form-error" role="alert">{error}</p>{/if}
  </form>
</section>
