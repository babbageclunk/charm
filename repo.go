package charm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"launchpad.net/juju/go/log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// InfoResponse is sent by the charm store in response to charm-info requests.
type InfoResponse struct {
	Revision int      `json:"revision"` // Zero is valid. Can't omitempty.
	Sha256   string   `json:"sha256,omitempty"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// Repo respresents a collection of charms.
type Repo interface {
	Get(curl *URL) (Charm, error)
	Latest(curl *URL) (int, error)
}

// store is a Repo that talks to the juju charm server (in ../store).
type store struct {
	baseURL   string
	cachePath string
}

const (
	STORE_URL  = "https://store.juju.ubuntu.com"
	CACHE_PATH = "$HOME/.juju/cache"
)

// Store returns a Repo that provides access to the juju charm store.
func Store() Repo {
	return &store{STORE_URL, os.ExpandEnv(CACHE_PATH)}
}

// info returns the revision and SHA256 digest of the charm referenced by curl.
func (s *store) info(curl *URL) (rev int, digest string, err error) {
	key := curl.String()
	resp, err := http.Get(s.baseURL + "/charm-info?charms=" + url.QueryEscape(key))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	infos := make(map[string]*InfoResponse)
	if err = json.Unmarshal(body, &infos); err != nil {
		return
	}
	info, found := infos[key]
	if !found {
		err = fmt.Errorf("missing info for charm: %q", key)
		return
	}
	for _, w := range info.Warnings {
		log.Printf("WARNING: info for %q: %s", key, w)
	}
	if info.Errors != nil {
		err = fmt.Errorf(
			"charm info errors for %q: %s", key, strings.Join(info.Errors, "; "),
		)
		return
	}
	return info.Revision, info.Sha256, nil
}

// Latest returns the latest revision of the charm referenced by curl, regardless
// of the revision set on curl itself.
func (s *store) Latest(curl *URL) (int, error) {
	rev, _, err := s.info(curl.WithRevision(-1))
	return rev, err
}

// verify returns an error unless a file exists at path with a hex-encoded
// SHA256 matching digest.
func verify(path, digest string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	h := sha256.New()
	h.Write(b)
	if hex.EncodeToString(h.Sum(nil)) != digest {
		return fmt.Errorf("bad SHA256 of %q", path)
	}
	return nil
}

// Get returns the charm referenced by curl.
func (s *store) Get(curl *URL) (Charm, error) {
	if err := os.MkdirAll(s.cachePath, 0755); err != nil {
		return nil, err
	}
	rev, digest, err := s.info(curl)
	if err != nil {
		return nil, err
	}
	if curl.Revision == -1 {
		curl = curl.WithRevision(rev)
	} else if curl.Revision != rev {
		return nil, fmt.Errorf("bad revision info for %q", curl.String())
	}
	path := filepath.Join(s.cachePath, Quote(curl.String())+".charm")
	if verify(path, digest) != nil {
		resp, err := http.Get(s.baseURL + "/charm/" + url.QueryEscape(curl.Path()))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		f, err := ioutil.TempFile("", "juju-charm-download")
		if err != nil {
			return nil, err
		}
		dlPath := f.Name()
		_, err = io.Copy(f, resp.Body)
		f.Close()
		if err != nil {
			os.Remove(dlPath)
			return nil, err
		}
		if err := os.Rename(dlPath, path); err != nil {
			return nil, err
		}
	}
	if err := verify(path, digest); err != nil {
		return nil, err
	}
	return ReadBundle(path)
}