package urlshort

import (
	"encoding/json"
	"net/http"

	"gopkg.in/yaml.v2"
)

type pathMapping struct {
	Path string
	URL  string
}

type Mapper interface {
	Map() (map[string]string, error)
}

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, ok := pathsToUrls[r.URL.Path]
		if !ok {
			fallback.ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//     - path: /some-path
//       url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	parsedYaml, err := parseYAML(yml)
	if err != nil {
		return nil, err
	}
	pathMap := buildMap(parsedYaml)
	return MapHandler(pathMap, fallback), nil
}

func JSONHandler(jsonData []byte, fallback http.Handler) (http.HandlerFunc, error) {
	parsedJSON, err := parseJSON(jsonData)
	if err != nil {
		return nil, err
	}
	pathMap := buildMap(parsedJSON)
	return MapHandler(pathMap, fallback), nil
}

func DBHandler(db Mapper, fallback http.Handler) (http.HandlerFunc, error) {
	pathMap, err := db.Map()
	if err != nil {
		return nil, err
	}
	return MapHandler(pathMap, fallback), nil
}

func parseYAML(yml []byte) ([]pathMapping, error) {
	var p []pathMapping
	err := yaml.Unmarshal(yml, &p)
	if err != nil {
		return nil, err
	}
	return p, err
}

func parseJSON(jsonData []byte) ([]pathMapping, error) {
	var p []pathMapping
	err := json.Unmarshal(jsonData, &p)
	if err != nil {
		return nil, err
	}
	return p, err
}

func buildMap(pathMaps []pathMapping) map[string]string {
	mapping := make(map[string]string)
	for _, m := range pathMaps {
		mapping[m.Path] = m.URL
	}
	return mapping
}
