create table access (
	accessed_at timestamp with time zone default current_timestamp,
	client text,
	query text,
	src text,
	dst text,
	cache_hit bool
);

create table cache (
	updated_at timestamp with time zone,
	query text,
	src text,
	dst text,
	response text,
	unique(query, src, dst)
);
