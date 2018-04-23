package db

import (
	"fmt"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

var conn *sql.DB

func Init() error {
	var err error
	conn, err = sql.Open("postgres", "user=dctnr password=dctnr dbname=dctnr sslmode=disable")
	if err != nil {
		fmt.Println("could not connect to db")
		return err
	}
	return nil
}

func GetCached(phrase, src, dst string, maxage time.Duration) (string, bool) {
	var response string
	row := conn.QueryRow(`
		select response
		from cache
		where
			query = $1
			and src = $2
			and dst = $3
			and updated_at > now() - interval '1 second' * $4
		`, phrase, src, dst, maxage.Seconds(),
	)
	err := row.Scan(&response)
	return response, err == nil
}

func PutCached(phrase, src, dst, response string) {
	_, err := conn.Exec(`
		insert into cache
		values (now(), $1, $2, $3, $4)
		on conflict (query, src, dst) do update
			set updated_at = excluded.updated_at,
			response = excluded.response
		`, phrase, src, dst, response,
	)
	if err != nil {
		fmt.Println("error during PutCached")
		fmt.Println(err)
	}
}

func LogAccess(client, phrase, src, dst string, cache_hit bool) {
	_, err := conn.Exec(`
		insert into access (client, query, src, dst, cache_hit)
		values ($1, $2, $3, $4, $5)
		`, client, phrase, src, dst, cache_hit,
	)
	if err != nil {
		fmt.Println("error during LogAccess")
		fmt.Println(err)
	}
}

type Stats struct {
	Cache struct {
		Count         int
		QueryUsage    int
		ResponseUsage int
		DiskUsage     int
	}
	Access struct {
		Count int
		Hits  int
	}
}

func GetStats() (*Stats, error) {
	var stats Stats

	row := conn.QueryRow(`
		select
			count(*),
			sum(length(query)),
			sum(length(response))
		from cache
	`)
	err := row.Scan(&stats.Cache.Count, &stats.Cache.QueryUsage, &stats.Cache.ResponseUsage)
	if err != nil {
		return nil, err
	}

	row = conn.QueryRow("select pg_total_relation_size('cache')")
	err = row.Scan(&stats.Cache.DiskUsage)
	if err != nil {
		return nil, err
	}

	row = conn.QueryRow("select count(*) from access")
	err = row.Scan(&stats.Access.Count)
	if err != nil {
		return nil, err
	}

	row = conn.QueryRow("select count(*) from access where cache_hit")
	err = row.Scan(&stats.Access.Hits)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func Ping() error {
	return conn.Ping()
}

func Finalize() {
	conn.Close()
	fmt.Println("db connection finalized")
}
