<script>
  let { entries, empty, linkLabel } = $props()
</script>

<div class="content-list">
  {#each entries.filter(entry => entry.published) as entry (entry.id)}
    <article class="project-card">
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
