import MarkdownIt from 'markdown-it'

const markdown = new MarkdownIt({ html: false, breaks: true })
const renderImage = markdown.renderer.rules.image
const galleryPrefix = 'gallery:'

markdown.renderer.rules.image = (tokens, index, options, env, renderer) => {
  const token = tokens[index]
  const source = token.attrGet('src')
  const src = source.startsWith(galleryPrefix)
    ? env.gallery.get(source.slice(galleryPrefix.length))?.src
    : source.startsWith('https://') && source
  if (!src) {
    const alt = renderer.renderInlineAsText(token.children, options, env)
    return markdown.utils.escapeHtml(alt)
  }
  token.attrSet('src', src)
  token.attrSet('loading', 'lazy')
  token.attrSet('decoding', 'async')
  const html = renderImage(tokens, index, options, env, renderer)
  if (!env.zoomImages) return html
  const previousLink = tokens.findLast((token, position) => position < index && (token.type === 'link_open' || token.type === 'link_close'))
  if (previousLink?.type === 'link_open') return html
  const alt = token.attrGet('alt')
  const label = markdown.utils.escapeHtml(alt ? `Perbesar gambar · ${alt}` : 'Perbesar gambar')
  return `<button class="entry-image-button markdown-zoom" type="button" aria-haspopup="dialog" aria-label="${label}">${html}</button>`
}

export function renderMarkdown(body, images, { zoomImages = false } = {}) {
  const gallery = new Map(images.filter(image => image.id).map(image => [image.id, image]))
  return markdown.render(body, { gallery, zoomImages })
}

export function imageMarkdown(image) {
  const alt = (image.alt || '').replace(/[\\[\]]/g, '\\$&').replace(/\s+/g, ' ')
  return `![${alt}](${galleryPrefix}${image.id})`
}

export function markdownText(body) {
  const env = {}
  return markdown.parse(body, env).map(token => token.children
    ? markdown.renderer.renderInlineAsText(token.children, markdown.options, env)
    : token.content).join(' ')
}
