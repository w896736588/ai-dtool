function stripCodeFenceLines(markdown) {
  const lineList = String(markdown || '').split(/\r?\n/)
  const nextLineList = []
  let inFence = false

  lineList.forEach((line) => {
    const trimmedLine = line.trim()
    if (/^```/.test(trimmedLine) || /^~~~/.test(trimmedLine)) {
      inFence = !inFence
      nextLineList.push('')
      return
    }
    nextLineList.push(inFence ? '' : line)
  })

  return nextLineList
}

function slugifyHeadingText(text) {
  const sourceText = String(text || '').trim().toLowerCase()
  if (sourceText === '') {
    return 'section'
  }

  let slug = ''
  Array.from(sourceText).forEach((char) => {
    if (/[a-z0-9]/.test(char)) {
      slug += char
      return
    }
    if (/\s|-/.test(char)) {
      slug += '-'
      return
    }
    slug += `u${char.codePointAt(0).toString(16)}`
  })

  slug = slug.replace(/-+/g, '-').replace(/^-|-$/g, '')
  return slug || 'section'
}

function buildMarkdownOutline(markdown) {
  const lineList = stripCodeFenceLines(markdown)
  const duplicateCountMap = {}

  return lineList.reduce((outlineList, line, index) => {
    const match = String(line || '').match(/^(#{1,3})\s+(.+?)\s*#*\s*$/)
    if (!match) {
      return outlineList
    }

    const level = match[1].length
    const text = String(match[2] || '').trim()
    if (text === '') {
      return outlineList
    }

    const baseSlug = slugifyHeadingText(text)
    const duplicateCount = (duplicateCountMap[baseSlug] || 0) + 1
    duplicateCountMap[baseSlug] = duplicateCount

    outlineList.push({
      level,
      text,
      slug: duplicateCount === 1 ? baseSlug : `${baseSlug}-${duplicateCount}`,
      lineNumber: index + 1,
    })
    return outlineList
  }, [])
}

module.exports = {
  buildMarkdownOutline,
  slugifyHeadingText,
}
