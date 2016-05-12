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
	var url = 'search?phrase=' + encodeURIComponent(phrase);
	url += '&src=' + encodeURIComponent(src);
	url += '&dst=' + encodeURIComponent(dst);
	console.log('querying ' + url);
	get(url, c_ok, c_fail);
}

function handleSubmit(evt) {
	evt.preventDefault();
	var phrase = document.getElementById('input-phrase').value;
	var src = document.getElementById('input-src').value;
	var dst = document.getElementById('input-dst').value;
	output.innerHTML = '';
	veil.style.display = 'block';
	search(phrase, src, dst, function(data) {
		console.log('succeeded');
		output.innerHTML = data;
		header.classList.remove('maximized');
		veil.style.display = 'none';
		output.style.display = 'block';
	}, function(data) {
		console.log('failed');
		output.innerHTML = data;
		header.classList.remove('maximized');
		veil.style.display = 'none';
		output.style.display = 'block';
	});
}

input.addEventListener('submit', handleSubmit, false);
