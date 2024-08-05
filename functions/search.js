const languages = ['en', 'de', 'fr', 'pl', 'ja', 'nl', 'it', 'pt', 'es', 'ru'];

export async function onRequest(context) {
  const inparams = new URL(context.request.url).searchParams;
  const src = inparams.get('src');
  if (!languages.includes(src)) {
    return new Response("unsupported src language", {status: 400});
  }
  const dst = inparams.get('dst');
  if (!languages.includes(dst)) {
    return new Response("unsupported dst language", {status: 400});
  }
  const query = inparams.get('query');

  var pages = await search(src, dst, query);
  console.log("pages", pages);
  if (src !== dst) {
    pages = await getAbstracts(dst, pages);
  }
  console.log("abstracts", pages);
  pages = pages.map((p) => ({title: p.title, extract: p.extract, url: p.fullurl}));
  return new Response(JSON.stringify(pages));
}

// available languages: https://en.wikipedia.org/w/api.php?format=json&action=query&meta=languageinfo&liprop=code|autonym

async function search(src, dst, query) {
  const outparams = new URLSearchParams();
  outparams.set('format', 'json');
  outparams.set('action', 'query');
  outparams.set('generator', 'search');
  outparams.set('gsrsearch', query);
  outparams.set('gsrlimit', 5);
  outparams.set('prop', 'extracts|langlinks|info');
  outparams.set('explaintext', true);
  outparams.set('exintro', true);
  outparams.set('exsentences', 2);
  outparams.set('lllang', dst);
  outparams.set('llprop', 'url');
  outparams.set('inprop', 'url');
  const resp = await fetch(`https://${src}.wikipedia.org/w/api.php?${outparams}`);
  const json = await resp.json();
  const pages = Object.values(json.query.pages).sort((a, b) => a.index - b.index);
  return pages;
}

async function getAbstracts(dst, pages) {
  const dst_pages = pages.filter((x) => x.langlinks !== undefined);

  const outparams = new URLSearchParams();
  outparams.set('format', 'json');
  outparams.set('action', 'query');
  outparams.set('titles', dst_pages.map((p) => p.langlinks[0]['*']).join('|'));
  outparams.set('redirects', true);
  outparams.set('prop', 'extracts|info');
  outparams.set('exintro', true);
  outparams.set('explaintext', true);
  outparams.set('exsentences', 2);
  outparams.set('inprop', 'url');
  const resp = await fetch(`https://${dst}.wikipedia.org/w/api.php?${outparams}`);
  const json = await resp.json();
  const url_order = dst_pages.reduce((o, p) => ({...o, [p.langlinks[0].url]: p.index}), {})
  console.log(url_order);
  const results = Object.values(json.query.pages).filter((p) => p.missing === undefined).sort((a, b) => (url_order[a.fullurl] || Infinity) - (url_order[b.fullurl] || Infinity));
  return results;
}
