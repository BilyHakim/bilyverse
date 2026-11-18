<script>
  import { tick } from 'svelte'
  import { markdownText, renderMarkdown } from './markdown.js'

  let { entries, empty, linkLabel, layout = 'list' } = $props()
  const readerTitle = $props.id()
  let reader = $state()
  let readingEntry = $state(null)
  let imageDialog = $state()
  let displayedImage = $state(null)

  function preview(text) {
    const characters = Array.from(text.trim().replace(/\s+/g, ' '))
    if (characters.length <= 160) return characters.join('')
    const excerpt = characters.slice(0, 160).join('')
    return (excerpt.replace(/\s+\S*$/, '') || excerpt) + '…'
  }

  async function readEntry(entry) {
    readingEntry = entry
    await tick()
    reader?.showModal()
  }

  async function openImage(src, alt) {
    displayedImage = { src, alt }
    await tick()
    imageDialog?.showModal()
  }

  function openArticleImage(event) {
    const image = event.target.closest('.markdown-zoom')?.querySelector('img')
    if (image) void openImage(image.src, image.alt)
  }
</script>

{#snippet imageButton(image, label, style = '')}
  <button class={'entry-image-button ' + style} type="button" aria-label={label} aria-haspopup="dialog" onclick={() => openImage(image.src, image.alt || '')}>
    <img src={image.src} alt={image.alt || ''} loading="lazy" decoding="async" />
  </button>
{/snippet}

{#snippet gallery(entry)}
  {#if entry.images.length > 1}
    <ul class="entry-gallery" aria-label={'Galeri ' + entry.title}>
      {#each entry.images.slice(1) as image, index}
        <li>
          {@render imageButton(image, `Perbesar gambar ${index + 2} · ${image.alt || entry.title}`)}
        </li>
      {/each}
    </ul>
  {/if}
{/snippet}

<div class="content-list" class:content-grid={layout !== 'list'} class:certificate-grid={layout === 'certificates'}>
  {#each entries.filter(entry => entry.published) as entry (entry.id)}
    {@const previewText = entry.summary.trim() || markdownText(entry.body).trim()}
    <article class="project-card">
      {#if entry.images.length}
        {@render imageButton(entry.images[0], 'Perbesar sampul ' + entry.title, 'entry-media')}
      {/if}
      <div class="project-body">
        {#if entry.date}<time class="eyebrow" datetime={entry.date}>{entry.date}</time>{/if}
        <h3>{entry.title}</h3>
        {#if previewText}<p class="entry-preview">{preview(previewText)}</p>{/if}
        {#if entry.url}<a class="entry-link" href={entry.url}>{linkLabel} <span aria-hidden="true">↗</span></a>{/if}
      </div>
      {@render gallery(entry)}
      {#if previewText || entry.body.trim() || entry.images.length}
        <button class="entry-read-more" type="button" aria-label={'Read more: ' + entry.title} aria-haspopup="dialog" onclick={() => readEntry(entry)}>
          Read more <span aria-hidden="true">→</span>
        </button>
      {/if}
    </article>
  {:else}
    <p class="empty-page">{empty}</p>
  {/each}
</div>

<dialog class="entry-reader" bind:this={reader} aria-labelledby={readerTitle} onclose={() => readingEntry = null}>
  {#if readingEntry}
    <header class="entry-reader-heading">
      <h2 id={readerTitle}>{readingEntry.title}</h2>
      <button class="button home-contact" type="button" onclick={() => reader.close()}>Tutup</button>
    </header>
    <div class="entry-reader-body">
      {#if readingEntry.images.length}
        {@render imageButton(readingEntry.images[0], 'Perbesar sampul ' + readingEntry.title, 'entry-reader-cover')}
      {/if}
      {#if readingEntry.summary.trim() && readingEntry.summary.trim() !== readingEntry.body.trim()}
        <p class="published-copy">{readingEntry.summary}</p>
      {/if}
      {#if readingEntry.body.trim()}
        <div class="markdown-copy" role="presentation" onclick={openArticleImage}>
          {@html renderMarkdown(readingEntry.body, readingEntry.images, { zoomImages: true })}
        </div>
      {/if}
      {@render gallery(readingEntry)}
      {#if readingEntry.url}<a class="entry-link" href={readingEntry.url}>{linkLabel} <span aria-hidden="true">↗</span></a>{/if}
    </div>
  {/if}
</dialog>

<dialog class="entry-image-dialog" bind:this={imageDialog} aria-label="Pratinjau gambar">
  <button class="entry-image-close" type="button" aria-label="Tutup gambar" onclick={() => imageDialog.close()}>
    <span aria-hidden="true">×</span>
  </button>
  {#if displayedImage}<img src={displayedImage.src} alt={displayedImage.alt} decoding="async" />{/if}
</dialog>
