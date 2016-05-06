var veil = document.getElementById('veil');
var input = document.getElementById('input');
var output = document.getElementById('output');

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

function search(phrase, c_ok, c_fail) {
	var url = '/search?phrase=' + encodeURI(phrase);
	console.log('querying ' + url);
	get(url, c_ok, c_fail);
}

function handleSubmit(evt) {
	evt.preventDefault();
	var phrase = document.getElementById('input-phrase').value;
	output.innerHTML = '';
	veil.style.display = 'block';
	search(phrase, function(data) {
		console.log('succeeded');
		output.innerHTML = data;
		veil.style.display = 'none';
	}, function(data) {
		console.log('failed');
		output.innerHTML = data;
		veil.style.display = 'none';
	});
}

input.addEventListener('submit', handleSubmit, false);
