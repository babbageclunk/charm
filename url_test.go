// Copyright 2011, 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package charm_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	gc "gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/yaml.v2"

	"gopkg.in/juju/charm.v6-unstable"
)

type URLSuite struct{}

var _ = gc.Suite(&URLSuite{})

var urlTests = []struct {
	s, err string
	exact  string
	url    *charm.URL
}{{
	s:     "cs:~user/trusty/name",
	exact: "cs:user/name/trusty",
	url:   &charm.URL{"cs", "user", "name", -1, "trusty"},
}, {
	s:     "cs:~user/wily/name-0",
	exact: "cs:user/name/wily/0",
	url:   &charm.URL{"cs", "user", "name", 0, "wily"},
}, {
	s:     "cs:raring/name",
	exact: "cs:name/raring",
	url:   &charm.URL{"cs", "", "name", -1, "raring"},
}, {
	s:     "cs:xenial/name-42",
	exact: "cs:name/xenial/42",
	url:   &charm.URL{"cs", "", "name", 42, "xenial"},
}, {
	s:     "local:precise/name-1",
	exact: "local:name/precise/1",
	url:   &charm.URL{"local", "", "name", 1, "precise"},
}, {
	s:     "local:saucy/name",
	exact: "local:name/saucy",
	url:   &charm.URL{"local", "", "name", -1, "saucy"},
}, {
	s:     "local:utopic/n0-0n-n0",
	exact: "local:n0-0n-n0/utopic",
	url:   &charm.URL{"local", "", "n0-0n-n0", -1, "utopic"},
}, {
	s:     "cs:~user/name",
	exact: "cs:user/name",
	url:   &charm.URL{"cs", "user", "name", -1, ""},
}, {
	s:   "cs:name",
	url: &charm.URL{"cs", "", "name", -1, ""},
}, {
	s:   "local:name",
	url: &charm.URL{"local", "", "name", -1, ""},
}, {
	s:     "http://jujucharms.com/u/user/name/vivid/1",
	url:   &charm.URL{"cs", "user", "name", 1, "vivid"},
	exact: "cs:user/name/vivid/1",
}, {
	s:     "http://www.jujucharms.com/u/user/name/precise/1",
	url:   &charm.URL{"cs", "user", "name", 1, "precise"},
	exact: "cs:user/name/precise/1",
}, {
	s:     "https://www.jujucharms.com/u/user/name/quantal/1",
	url:   &charm.URL{"cs", "user", "name", 1, "quantal"},
	exact: "cs:user/name/quantal/1",
}, {
	s:     "https://jujucharms.com/u/user/name/raring/1",
	url:   &charm.URL{"cs", "user", "name", 1, "raring"},
	exact: "cs:user/name/raring/1",
}, {
	s:     "https://jujucharms.com/u/user/name/saucy",
	url:   &charm.URL{"cs", "user", "name", -1, "saucy"},
	exact: "cs:user/name/saucy",
}, {
	s:     "https://jujucharms.com/u/user/name/1",
	url:   &charm.URL{"cs", "user", "name", 1, ""},
	exact: "cs:user/name/1",
}, {
	s:     "https://jujucharms.com/u/user/name",
	url:   &charm.URL{"cs", "user", "name", -1, ""},
	exact: "cs:user/name",
}, {
	s:     "https://jujucharms.com/name",
	url:   &charm.URL{"cs", "", "name", -1, ""},
	exact: "cs:name",
}, {
	s:     "https://jujucharms.com/name/utopic",
	url:   &charm.URL{"cs", "", "name", -1, "utopic"},
	exact: "cs:name/utopic",
}, {
	s:     "https://jujucharms.com/name/1",
	url:   &charm.URL{"cs", "", "name", 1, ""},
	exact: "cs:name/1",
}, {
	s:     "https://jujucharms.com/name/vivid/1",
	url:   &charm.URL{"cs", "", "name", 1, "vivid"},
	exact: "cs:name/vivid/1",
}, {
	s:     "https://jujucharms.com/u/user/name/wily/1/",
	url:   &charm.URL{"cs", "user", "name", 1, "wily"},
	exact: "cs:user/name/wily/1",
}, {
	s:     "https://jujucharms.com/u/user/name/xenial/",
	url:   &charm.URL{"cs", "user", "name", -1, "xenial"},
	exact: "cs:user/name/xenial",
}, {
	s:     "https://jujucharms.com/u/user/name/1/",
	url:   &charm.URL{"cs", "user", "name", 1, ""},
	exact: "cs:user/name/1",
}, {
	s:     "https://jujucharms.com/u/user/name/",
	url:   &charm.URL{"cs", "user", "name", -1, ""},
	exact: "cs:user/name",
}, {
	s:     "https://jujucharms.com/name/",
	url:   &charm.URL{"cs", "", "name", -1, ""},
	exact: "cs:name",
}, {
	s:     "https://jujucharms.com/name/precise/",
	url:   &charm.URL{"cs", "", "name", -1, "precise"},
	exact: "cs:name/precise",
}, {
	s:     "https://jujucharms.com/name/1/",
	url:   &charm.URL{"cs", "", "name", 1, ""},
	exact: "cs:name/1",
}, {
	s:     "https://jujucharms.com/name/quantal/1/",
	url:   &charm.URL{"cs", "", "name", 1, "quantal"},
	exact: "cs:name/quantal/1",
}, {
	s:   "https://jujucharms.com/",
	err: `URL has invalid charm or bundle name: $URL`,
}, {
	s:   "https://jujucharms.com/bad.wolf",
	err: `URL has invalid charm or bundle name: $URL`,
}, {
	s:   "https://jujucharms.com/u/",
	err: "charm or bundle URL $URL malformed, expected \"/u/<user>/<name>\"",
}, {
	s:   "https://jujucharms.com/u/badwolf",
	err: "charm or bundle URL $URL malformed, expected \"/u/<user>/<name>\"",
}, {
	s:   "https://jujucharms.com/name/raring/badwolf",
	err: "charm or bundle URL has malformed revision: \"badwolf\" in $URL",
}, {
	s:   "https://jujucharms.com/name/badwolf/42",
	err: `charm or bundle URL has invalid series: $URL`,
}, {
	s:   "https://badwolf@jujucharms.com/name/saucy/42",
	err: `charm or bundle URL $URL has unrecognized parts`,
}, {
	s:   "https://jujucharms.com/name/trusty/42#bad-wolf",
	err: `charm or bundle URL $URL has unrecognized parts`,
}, {
	s:   "https://jujucharms.com/name//42?bad=wolf",
	err: `charm or bundle URL $URL has unrecognized parts`,
}, {
	s:   "bs:~user/utopic/name-1",
	err: `charm or bundle URL has invalid schema: $URL`,
}, {
	s:   ":foo",
	err: `cannot parse charm or bundle URL: $URL`,
}, {
	s:   "cs:~1/vivid/name-1",
	err: `charm or bundle URL has invalid user name: $URL`,
}, {
	s:   "cs:~user",
	err: `URL without charm or bundle name: $URL`,
}, {
	s:   "cs:~user/unknown/name-1",
	err: `charm or bundle URL has invalid series: $URL`,
}, {
	s:   "cs:~user/wily/name-1-2",
	err: `URL has invalid charm or bundle name: $URL`,
}, {
	s:   "cs:~user/xenial/name-1-name-2",
	err: `URL has invalid charm or bundle name: $URL`,
}, {
	s:   "cs:~user/precise/name--name-2",
	err: `URL has invalid charm or bundle name: $URL`,
}, {
	s:   "cs:foo-1-2",
	err: `URL has invalid charm or bundle name: $URL`,
}, {
	s:   "cs:~user/quantal/huh/name-1",
	err: `charm or bundle URL has invalid form: $URL`,
}, {
	s:   "cs:~user/production/raring/name-1",
	err: `charm or bundle URL has invalid form: $URL`,
}, {
	s:   "cs:~user/development/saucy/badwolf/name-1",
	err: `charm or bundle URL has invalid form: $URL`,
}, {
	s:   "cs:/name",
	err: `charm or bundle URL has invalid series: $URL`,
}, {
	s:   "local:~user/trusty/name",
	err: `local charm or bundle URL with user name: $URL`,
}, {
	s:   "local:~user/name",
	err: `local charm or bundle URL with user name: $URL`,
}, {
	s:     "precise/wordpress",
	exact: "cs:precise/wordpress",
	url:   &charm.URL{"cs", "", "wordpress", -1, "precise"},
}, {
	s:     "foo",
	exact: "cs:foo",
	url:   &charm.URL{"cs", "", "foo", -1, ""},
}, {
	s:     "foo-1",
	exact: "cs:foo-1",
	url:   &charm.URL{"cs", "", "foo", 1, ""},
}, {
	s:     "n0-n0-n0",
	exact: "cs:n0-n0-n0",
	url:   &charm.URL{"cs", "", "n0-n0-n0", -1, ""},
}, {
	s:     "cs:foo",
	exact: "cs:foo",
	url:   &charm.URL{"cs", "", "foo", -1, ""},
}, {
	s:     "local:foo",
	exact: "local:foo",
	url:   &charm.URL{"local", "", "foo", -1, ""},
}, {
	s:     "vivid/foo",
	exact: "cs:vivid/foo",
	url:   &charm.URL{"cs", "", "foo", -1, "vivid"},
}, {
	s:   "wily/foo/bar",
	err: `charm or bundle URL has invalid form: "wily/foo/bar"`,
}, {
	s:   "cs:foo/~blah",
	err: `URL has invalid charm or bundle name: "cs:foo/~blah"`,
}, {
	s:     "babbageclunk/mysql/xenial/20",
	exact: "cs:babbageclunk/mysql/xenial/20",
	url:   &charm.URL{"cs", "babbageclunk", "mysql", 20, "xenial"},
}, {
	s:     "babbageclunk/mysql/wily",
	exact: "cs:babbageclunk/mysql/wily",
	url:   &charm.URL{"cs", "babbageclunk", "mysql", -1, "wily"},
}, {
	s:     "babbageclunk/mysql/10",
	exact: "cs:babbageclunk/mysql/10",
	url:   &charm.URL{"cs", "babbageclunk", "mysql", 10, ""},
}, {
	s:     "mysql/quantal/15",
	exact: "cs:mysql/quantal/15",
	url:   &charm.URL{"cs", "", "mysql", 15, "quantal"},
}, {
	s:     "babbageclunk/mysql",
	exact: "cs:babbageclunk/mysql",
	url:   &charm.URL{"cs", "babbageclunk", "mysql", -1, ""},
}, {
	s:     "mysql/trusty",
	exact: "cs:mysql/trusty",
	url:   &charm.URL{"cs", "trusty", "mysql", -1, "trusty"},
}, {
	s:     "mysql/15",
	exact: "cs:mysql/15",
	url:   &charm.URL{"cs", "", "mysql", 15, ""},
}, {
	s:     "mysql",
	exact: "cs:mysql",
	url:   &charm.URL{"cs", "", "mysql", -1, ""},
}}

func (s *URLSuite) TestParseURL(c *gc.C) {
	for i, t := range urlTests {
		c.Logf("test %d: %q", i, t.s)

		expectStr := t.s
		if t.exact != "" {
			expectStr = t.exact
		}
		url, uerr := charm.ParseURL(t.s)
		if t.err != "" {
			t.err = strings.Replace(t.err, "$URL", regexp.QuoteMeta(fmt.Sprintf("%q", t.s)), -1)
			c.Assert(uerr, gc.ErrorMatches, t.err)
			c.Assert(url, gc.IsNil)
			continue
		}
		c.Assert(uerr, gc.IsNil)
		c.Assert(url, gc.DeepEquals, t.url)
		c.Assert(url.String(), gc.Equals, expectStr)

		// URL strings are generated as expected.  Reversability is preserved
		// with v1 URLs.
		if t.exact != "" {
			c.Check(url.String(), gc.Equals, t.exact)
		} else {
			c.Check(url.String(), gc.Equals, t.s)
		}
	}
}

var inferTests = []struct {
	vague, exact string
}{
	{"foo", "cs:defseries/foo"},
	{"foo-1", "cs:defseries/foo-1"},
	{"n0-n0-n0", "cs:defseries/n0-n0-n0"},
	{"cs:foo", "cs:defseries/foo"},
	{"local:foo", "local:defseries/foo"},
	{"series/foo", "cs:series/foo"},
	{"cs:series/foo", "cs:series/foo"},
	{"local:series/foo", "local:series/foo"},
	{"cs:~user/foo", "cs:~user/defseries/foo"},
	{"cs:~user/series/foo", "cs:~user/series/foo"},
	{"local:~user/series/foo", "local:~user/series/foo"},
	{"bs:foo", "bs:defseries/foo"},
	{"cs:~1/foo", "cs:~1/defseries/foo"},
	{"cs:foo-1-2", "cs:defseries/foo-1-2"},
}

func (s *URLSuite) TestInferURL(c *gc.C) {
	for i, t := range inferTests {
		c.Logf("test %d", i)
		comment := gc.Commentf("InferURL(%q, %q)", t.vague, "defseries")
		inferred, ierr := charm.InferURL(t.vague, "defseries")
		parsed, perr := charm.ParseURL(t.exact)
		if perr == nil {
			c.Check(inferred, gc.DeepEquals, parsed, comment)
			c.Check(ierr, gc.IsNil)
		} else {
			expect := perr.Error()
			if t.vague != t.exact {
				if colIdx := strings.Index(expect, ":"); colIdx > 0 {
					expect = expect[:colIdx]
				}
			}
			c.Check(ierr.Error(), gc.Matches, expect+".*", comment)
		}
	}
	u, err := charm.InferURL("~blah", "defseries")
	c.Assert(u, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "URL without charm or bundle name: .*")
}

var inferNoDefaultSeriesTests = []struct {
	vague, exact string
	resolved     bool
}{
	{"foo", "", false},
	{"foo-1", "", false},
	{"cs:foo", "", false},
	{"cs:~user/foo", "", false},
	{"series/foo", "cs:series/foo", true},
	{"cs:series/foo", "cs:series/foo", true},
	{"cs:~user/series/foo", "cs:~user/series/foo", true},
}

func (s *URLSuite) TestInferURLNoDefaultSeries(c *gc.C) {
	for i, t := range inferNoDefaultSeriesTests {
		c.Logf("%d: %s", i, t.vague)
		inferred, err := charm.InferURL(t.vague, "")
		if t.exact == "" {
			c.Assert(err, gc.ErrorMatches, fmt.Sprintf("cannot infer charm or bundle URL for %q: charm or bundle url series is not resolved", t.vague))
		} else {
			parsed, err := charm.ParseURL(t.exact)
			c.Assert(err, gc.IsNil)
			c.Assert(inferred, gc.DeepEquals, parsed, gc.Commentf(`InferURL(%q, "")`, t.vague))
		}
	}
}

var validTests = []struct {
	valid  func(string) bool
	string string
	expect bool
}{

	{charm.IsValidName, "", false},
	{charm.IsValidName, "wordpress", true},
	{charm.IsValidName, "Wordpress", false},
	{charm.IsValidName, "word-press", true},
	{charm.IsValidName, "word press", false},
	{charm.IsValidName, "word^press", false},
	{charm.IsValidName, "-wordpress", false},
	{charm.IsValidName, "wordpress-", false},
	{charm.IsValidName, "wordpress2", true},
	{charm.IsValidName, "wordpress-2", false},
	{charm.IsValidName, "word2-press2", true},

	{charm.IsValidSeries, "", false},
	{charm.IsValidSeries, "precise", true},
	{charm.IsValidSeries, "Precise", false},
	{charm.IsValidSeries, "pre cise", false},
	{charm.IsValidSeries, "pre-cise", false},
	{charm.IsValidSeries, "pre^cise", false},
	{charm.IsValidSeries, "prec1se", true},
	{charm.IsValidSeries, "-precise", false},
	{charm.IsValidSeries, "precise-", false},
	{charm.IsValidSeries, "precise-1", false},
	{charm.IsValidSeries, "precise1", true},
	{charm.IsValidSeries, "pre-c1se", false},
}

func (s *URLSuite) TestValidCheckers(c *gc.C) {
	for i, t := range validTests {
		c.Logf("test %d: %s", i, t.string)
		c.Assert(t.valid(t.string), gc.Equals, t.expect, gc.Commentf("%s", t.string))
	}
}

func (s *URLSuite) TestMustParseURL(c *gc.C) {
	url := charm.MustParseURL("cs:series/name")
	c.Assert(url, gc.DeepEquals, &charm.URL{"cs", "", "name", -1, "series"})
	f := func() { charm.MustParseURL("local:@@/name") }
	c.Assert(f, gc.PanicMatches, "charm or bundle URL has invalid series: .*")
	f = func() { charm.MustParseURL("cs:~user") }
	c.Assert(f, gc.PanicMatches, "URL without charm or bundle name: .*")
	f = func() { charm.MustParseURL("cs:~user") }
	c.Assert(f, gc.PanicMatches, "URL without charm or bundle name: .*")
}

func (s *URLSuite) TestWithRevision(c *gc.C) {
	url := charm.MustParseURL("cs:series/name")
	other := url.WithRevision(1)
	c.Assert(url, gc.DeepEquals, &charm.URL{"cs", "", "name", -1, "series"})
	c.Assert(other, gc.DeepEquals, &charm.URL{"cs", "", "name", 1, "series"})

	// Should always copy. The opposite behavior is error prone.
	c.Assert(other.WithRevision(1), gc.Not(gc.Equals), other)
	c.Assert(other.WithRevision(1), gc.DeepEquals, other)
}

var codecs = []struct {
	Name      string
	Marshal   func(interface{}) ([]byte, error)
	Unmarshal func([]byte, interface{}) error
}{{
	Name:      "bson",
	Marshal:   bson.Marshal,
	Unmarshal: bson.Unmarshal,
}, {
	Name:      "json",
	Marshal:   json.Marshal,
	Unmarshal: json.Unmarshal,
}, {
	Name:      "yaml",
	Marshal:   yaml.Marshal,
	Unmarshal: yaml.Unmarshal,
}}

func (s *URLSuite) TestURLCodecs(c *gc.C) {
	for i, codec := range codecs {
		c.Logf("codec %d: %v", i, codec.Name)
		type doc struct {
			URL *charm.URL `json:",omitempty" bson:",omitempty" yaml:",omitempty"`
		}
		url := charm.MustParseURL("cs:series/name")
		v0 := doc{url}
		data, err := codec.Marshal(v0)
		c.Assert(err, gc.IsNil)
		var v doc
		err = codec.Unmarshal(data, &v)
		c.Assert(v, gc.DeepEquals, v0)

		// Check that the underlying representation
		// is a string.
		type strDoc struct {
			URL string
		}
		var vs strDoc
		err = codec.Unmarshal(data, &vs)
		c.Assert(err, gc.IsNil)
		c.Assert(vs.URL, gc.Equals, "cs:series/name")

		data, err = codec.Marshal(doc{})
		c.Assert(err, gc.IsNil)
		v = doc{}
		err = codec.Unmarshal(data, &v)
		c.Assert(err, gc.IsNil)
		c.Assert(v.URL, gc.IsNil, gc.Commentf("data: %q", data))
	}
}

func (s *URLSuite) TestJSONGarbage(c *gc.C) {
	// unmarshalling json gibberish
	for _, value := range []string{":{", `"cs:{}+<"`, `"cs:~_~/f00^^&^/baaaar$%-?"`} {
		err := json.Unmarshal([]byte(value), new(struct{ URL *charm.URL }))
		c.Check(err, gc.NotNil)
	}
}

type QuoteSuite struct{}

var _ = gc.Suite(&QuoteSuite{})

func (s *QuoteSuite) TestUnmodified(c *gc.C) {
	// Check that a string containing only valid
	// chars stays unmodified.
	in := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-"
	out := charm.Quote(in)
	c.Assert(out, gc.Equals, in)
}

func (s *QuoteSuite) TestQuote(c *gc.C) {
	// Check that invalid chars are translated correctly.
	in := "hello_there/how'are~you-today.sir"
	out := charm.Quote(in)
	c.Assert(out, gc.Equals, "hello_5f_there_2f_how_27_are_7e_you-today.sir")
}
