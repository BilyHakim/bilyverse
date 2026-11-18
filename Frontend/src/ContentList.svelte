<script>
  let { entries, empty, linkLabel, layout = 'list' } = $props()
</script>

<div class="content-list" class:content-grid={layout !== 'list'} class:certificate-grid={layout === 'certificates'}>
  {#each entries.filter(entry => entry.published) as entry (entry.id)}
    <article class="project-card">
      {#if entry.images.length}
        <a class="entry-media" href={entry.images[0].src} aria-label={'Buka gambar sampul ' + entry.title}>
          <img src={entry.images[0].src} alt={entry.images[0].alt || ''} loading="lazy" decoding="async" />
        </a>
      {/if}
      <div class="project-body">
        {#if entry.date}<time class="eyebrow" datetime={entry.date}>{entry.date}</time>{/if}
        <h3>{entry.title}</h3>
        {#if entry.summary}<p class="published-copy">{entry.summary}</p>{/if}
        {#if entry.url}<a class="entry-link" href={entry.url}>{linkLabel} <span aria-hidden="true">↗</span></a>{/if}
      </div>
      {#if entry.images.length > 1}
        <ul class="entry-gallery" aria-label={'Galeri ' + entry.title}>
          {#each entry.images.slice(1) as image, index}
            <li>
              <a href={image.src} aria-label={image.alt ? 'Buka gambar ' + (index + 2) + ' · ' + image.alt : 'Buka gambar ' + (index + 2) + ' untuk ' + entry.title}>
                <img src={image.src} alt={image.alt || ''} loading="lazy" decoding="async" />
              </a>
            </li>
          {/each}
        </ul>
      {/if}
      {#if entry.body}
        <details class="project-details">
          <summary>Baca selengkapnya <span aria-hidden="true"></span></summary>
          <p class="published-copy">{entry.body}</p>
        </details>
      {/if}
    </article>
  {:else}
    <p class="empty-page">{empty}</p>
  {/each}
</div>
