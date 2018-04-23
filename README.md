# dctnr

A Wikipedia-based dictionary.

`dctnr` provides a UI for searching articles on Wikipedia in one language and
automatically finding the same articles in another language, effectively
working as a dictionary. It fetches and caches article previews using the
MediaWiki API.

## Configuration

For the sake of simplicity, all the parameters are hard-coded. `dctnr` will
listen on port 4000 and connect to a postgres database named `dctnr` at
`localhost:5432` with password `dctnr`. You should create the tables yourself
using `schema.sql`.
