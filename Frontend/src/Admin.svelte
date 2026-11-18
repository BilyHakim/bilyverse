<script>
  import { contentFields, defaultContent } from './content.js'

  let { onpublish } = $props()
  let key = $state('')
  let draft = $state(null)
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
      } else {
        await request('PUT', JSON.stringify({ revision, fields: draft }))
        revision += 1
        onpublish({ ...draft }, revision)
        status = 'Perubahan sudah diterbitkan.'
      }
    } catch (failure) {
      error = failure.message
    } finally {
      busy = false
    }
  }

  function closeEditor() {
    draft = null
    key = ''
    status = ''
    error = ''
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
        {#each contentFields as [field, label, limit]}
          <label>
            {label}
            {#if limit > 120}
              <textarea rows="4" maxlength={limit} bind:value={draft[field]}></textarea>
            {:else}
              <input maxlength={limit} bind:value={draft[field]} />
            {/if}
          </label>
        {/each}
        <div class="admin-actions">
          <button class="button button-primary" type="submit">Terbitkan perubahan</button>
          <button class="button home-contact" type="button" onclick={closeEditor}>Tutup editor</button>
        </div>
      {/if}
    </fieldset>
    {#if busy}<p role="status">Sedang memproses…</p>{/if}
    {#if status}<p class="form-status" role="status">{status}</p>{/if}
    {#if error}<p class="form-error" role="alert">{error}</p>{/if}
  </form>
</section>
