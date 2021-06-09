exports.map = function (doc) {
  if (doc.docType === "file") {
    emit(doc.filename);
  }
}