function buildMdEditorCodeMirrorExtensions(extensions) {
  return (Array.isArray(extensions) ? extensions : []).filter((item) => {
    return item && item.type !== 'linkShortener'
  })
}

module.exports = {
  buildMdEditorCodeMirrorExtensions,
}
