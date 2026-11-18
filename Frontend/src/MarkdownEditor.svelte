<script>
  import { tick } from 'svelte'
  import { imageMarkdown, renderMarkdown } from './markdown.js'

  let { label, images, value = $bindable('') } = $props()
  const inputId = $props.id()
  const limit = 12000
  const formats = [
    ['Judul', '## ', ''],
    ['Tebal', '**', '**'],
    ['Miring', '*', '*'],
    ['Daftar', '- ', ''],
    ['Tautan', '[', '](https://)'],
    ['Kode', '`', '`'],
  ]
  let input = $state()
  let previewing = $state(false)
  let imageId = $state('')
  let notice = $state('')
  let selectedImage = $derived(images.find(image => image.id === imageId))

  async function replaceSelection(text, start, end, cursor) {
    const updated = value.slice(0, start) + text + value.slice(end)
    if (updated.length > limit) {
      notice = 'Isi sudah mendekati batas 12.000 karakter.'
      return
    }
    value = updated
    notice = ''
    await tick()
    input?.focus()
    input?.setSelectionRange(cursor, cursor)
  }

  function format(name, before, after) {
    const { selectionStart: start, selectionEnd: end } = input
    const text = value.slice(start, end) || name
    const needsLine = (before === '## ' || before === '- ') && start > 0 && value[start - 1] !== '\n'
    const prefix = needsLine ? '\n' + before : before
    return replaceSelection(prefix + text + after, start, end, start + prefix.length + text.length + after.length)
  }

  function insertImage() {
    if (!selectedImage) return
    const { selectionEnd: position } = input
    const text = '\n\n' + imageMarkdown(selectedImage) + '\n\n'
    return replaceSelection(text, position, position, position + text.length)
  }
</script>

<div class="markdown-editor">
  <div class="markdown-editor-header">
    <label for={inputId}>{label} (Markdown)</label>
    <button class="button home-contact" type="button" aria-expanded={previewing} onclick={() => previewing = !previewing}>
      {previewing ? 'Kembali menulis' : 'Pratinjau'}
    </button>
  </div>
  {#if previewing}
    <div class="markdown-copy markdown-preview" aria-label={'Pratinjau ' + label}>
      {#if value.trim()}
        {@html renderMarkdown(value, images)}
      {:else}
        <p>Belum ada isi.</p>
      {/if}
    </div>
  {:else}
    <div class="markdown-editor-tools" role="group" aria-label="Format Markdown">
      {#each formats as [name, before, after]}
        <button type="button" onclick={() => format(name, before, after)}>{name}</button>
      {/each}
    </div>
    <textarea id={inputId} rows="7" maxlength={limit} bind:this={input} bind:value oninput={() => notice = ''}></textarea>
    {#if images.length}
      <div class="markdown-image-picker">
        <select aria-label="Pilih gambar dari galeri" bind:value={imageId}>
          <option value="">Pilih gambar dari galeri</option>
          {#each images as image, index (image.id)}
            <option value={image.id}>{index + 1}. {image.alt || (index === 0 ? 'Sampul' : 'Gambar')}</option>
          {/each}
        </select>
        <button class="button home-contact" type="button" disabled={!selectedImage} onclick={insertImage}>Sisipkan gambar</button>
      </div>
    {/if}
  {/if}
  {#if notice}<p class="form-error" role="status">{notice}</p>{/if}
</div>
