var veil = document.getElementById('veil');
var input = document.getElementById('input');
var output = document.getElementById('output');
var header = document.getElementById('header');

var xhttp = undefined;

function get(url, c_ok, c_fail) {
	xhttp = new XMLHttpRequest();
	xhttp.onreadystatechange = function() {
		if (xhttp.readyState == 4) {
			if (xhttp.status == 200) {
				c_ok(xhttp.responseText);
			} else {
				c_fail(xhttp.responseText);
			}
			xhttp = undefined;
		}
	};
	xhttp.open("GET", url, true);
	xhttp.send();
}

function search(phrase, src, dst, c_ok, c_fail) {
	console.log('querying ' + url);
	get(url, c_ok, c_fail);
}

async function handleSubmit(evt) {
	evt.preventDefault();

	output.innerHTML = '';
	veil.style.display = 'block';

	const params = new URLSearchParams();
	params.set('src', document.getElementById('input-src').value);
	params.set('dst', document.getElementById('input-dst').value);
	params.set('query', document.getElementById('input-query').value);

	const resp = await fetch(`/search?${params}`);
	if (resp.ok) {
		for (const p of await resp.json()) {
			const a = document.createElement('a');
			a.appendChild(document.createTextNode(p.extract));
			a.title = p.title;
			a.href = p.url;
			output.appendChild(a);
		}
	} else {
		console.log('failed');
	}

	header.classList.remove('maximized');
	veil.style.display = 'none';
	output.style.display = 'block';
}

input.addEventListener('submit', handleSubmit, false);
