const allowedSynopsisTags = new Set(['BR', 'I', 'EM', 'B', 'STRONG', 'A'])

function isSafeLink(href: string): boolean {
  return href.startsWith('https://') || href.startsWith('http://')
}

function sanitizeSynopsisNode(source: Node, target: HTMLElement, document: Document) {
  for (const child of source.childNodes) {
    if (child.nodeType === Node.TEXT_NODE) {
      target.appendChild(document.createTextNode(child.textContent ?? ''))
      continue
    }
    if (child.nodeType !== Node.ELEMENT_NODE) {
      continue
    }

    const element = child as HTMLElement
    const tagName = element.tagName

    if (tagName === 'BR') {
      target.appendChild(document.createElement('br'))
      continue
    }

    if (!allowedSynopsisTags.has(tagName)) {
      sanitizeSynopsisNode(element, target, document)
      continue
    }

    const clone = document.createElement(tagName.toLowerCase())
    if (tagName === 'A') {
      const href = element.getAttribute('href') ?? ''
      if (isSafeLink(href)) {
        clone.setAttribute('href', href)
        clone.setAttribute('rel', 'noopener noreferrer')
        clone.setAttribute('target', '_blank')
      }
    }
    sanitizeSynopsisNode(element, clone, document)
    target.appendChild(clone)
  }
}

export function sanitizeAnilistSynopsis(rawHtml: string): string {
  const trimmed = rawHtml.trim()
  if (!trimmed) {
    return ''
  }

  const document = new DOMParser().parseFromString(trimmed, 'text/html')
  const output = document.createElement('div')
  sanitizeSynopsisNode(document.body, output, document)
  return output.innerHTML
}
