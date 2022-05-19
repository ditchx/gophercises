package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/boltdb/bolt"
	"github.com/ditchx/gophercises/urlshort"
)

type BoltDBStore struct {
	db         *bolt.DB
	bucketName string
}

func NewBoltDBStore(d *bolt.DB) *BoltDBStore {
	return &BoltDBStore{db: d, bucketName: "pathMap"}
}

func (s *BoltDBStore) Map() (map[string]string, error) {
	pathMap := make(map[string]string)
	db := s.db
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucketName))

		b.ForEach(func(k, v []byte) error {
			pathMap[string(k)] = string(v)
			return nil
		})
		return nil
	})
	return pathMap, nil
}

func (s *BoltDBStore) Populate() error {
	var updateErr error
	db := s.db
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucketName))

		if b != nil {
			updateErr = nil
			return updateErr
		}

		b, err := tx.CreateBucket([]byte(s.bucketName))
		if err != nil {
			updateErr = fmt.Errorf("failed to create bucket %s: %w", s.bucketName, err)
			return updateErr
		}

		pathMap := map[string]string{
			"/fb":   "https://facebook.com",
			"/twtr": "https://twitter.com",
			"/ig":   "https://instagram.com",
		}

		for k, v := range pathMap {
			updateErr = b.Put([]byte(k), []byte(v))
			if updateErr != nil {
				return updateErr
			}
		}

		return updateErr
	})
	return updateErr

}

func (s *BoltDBStore) AddPath(path, url string) error {
	var updateErr error
	db := s.db
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucketName))
		updateErr = b.Put([]byte(path), []byte(url))
		return updateErr
	})
	return updateErr
}

func (s *BoltDBStore) AddMap(pathMap map[string]string) error {
	var updateErr error
	db := s.db
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucketName))

		for k, v := range pathMap {
			updateErr = b.Put([]byte(k), []byte(v))
			if updateErr != nil {
				return updateErr
			}
		}

		return updateErr
	})
	return updateErr
}

func main() {
	var yamlFile string

	flag.StringVar(&yamlFile, "yaml", "", "yaml file containing path name to url mappings")
	flag.Parse()

	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := urlshort.MapHandler(pathsToUrls, mux)

	// Build the YAMLHandler using the mapHandler as the
	// fallback

	yaml := []byte(`
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort-final
  url: https://github.com/gophercises/urlshort/tree/solution
`)
	var err error
	if yamlFile != "" {
		log.Printf("Loading path to URL maps from YAML file: %s\n", yamlFile)
		yaml, err = os.ReadFile(yamlFile)
	}

	if err != nil {
		log.Fatalf("failed reading specified YAML file %s: %s", yamlFile, err)
	}

	yamlHandler, err := urlshort.YAMLHandler(yaml, mapHandler)
	if err != nil {
		panic(err)
	}

	jsonData := []byte(`[{"path": "/g", "url" : "https://google.com"}, {"path": "/yt", "url" : "https://youtube.com"}]`)
	jsonHandler, err := urlshort.JSONHandler(jsonData, yamlHandler)
	if err != nil {
		panic(err)
	}

	db, err := bolt.Open("pathMap.db", 0600, nil)
	if err != nil {
		panic(err)
	}

	boltDB := NewBoltDBStore(db)
	boltDB.Populate()
	dbHandler, err := urlshort.DBHandler(boltDB, jsonHandler)
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", dbHandler)

}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
