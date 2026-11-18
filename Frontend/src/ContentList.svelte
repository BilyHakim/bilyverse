<script>
  let { entries, empty, linkLabel, layout = 'list' } = $props()
</script>

<div class="content-list" class:content-grid={layout !== 'list'} class:certificate-grid={layout === 'certificates'}>
  {#each entries.filter(entry => entry.published) as entry (entry.id)}
    <article class="project-card">
      {#if entry.image && layout !== 'list'}
        <div class="entry-media">
          <img src={entry.image} alt={entry.imageAlt || ''} loading="lazy" decoding="async" />
        </div>
      {/if}
      <div class="project-body">
        {#if entry.date}<time class="eyebrow" datetime={entry.date}>{entry.date}</time>{/if}
        <h3>{entry.title}</h3>
        {#if entry.summary}<p class="published-copy">{entry.summary}</p>{/if}
        {#if entry.url}<a class="entry-link" href={entry.url}>{linkLabel} <span aria-hidden="true">↗</span></a>{/if}
      </div>
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
