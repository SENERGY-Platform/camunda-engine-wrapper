/*
 * Copyright 2018 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package etree

import "testing"

var testXML = `
<?xml version="1.0" encoding="UTF-8"?>
<bookstore xmlns:p="urn:books-com:prices">

	<!Directive>

	<book category="COOKING">
		<title lang="en">Everyday Italian</title>
		<author>Giada De Laurentiis</author>
		<year>2005</year>
		<p:price>30.00</p:price>
	</book>

	<book category="CHILDREN">
		<title lang="en" sku="150">Harry Potter</title>
		<author>J K. Rowling</author>
		<year>2005</year>
		<p:price>29.99</p:price>
	</book>

	<book category="WEB">
		<title lang="en">XQuery Kick Start</title>
		<author>James McGovern</author>
		<author>Per Bothner</author>
		<author>Kurt Cagle</author>
		<author>James Linn</author>
		<author>Vaidyanathan Nagarajan</author>
		<year>2003</year>
		<p:price>49.99</p:price>
	</book>

	<!-- Final book -->
	<book category="WEB">
		<title lang="en">Learning XML</title>
		<author>Erik T. Ray</author>
		<year>2003</year>
		<p:price>39.95</p:price>
	</book>

</bookstore>
`

type test struct {
	path   string
	result interface{}
}

type errorResult string

var tests = []test{

	// basic queries
	{"./bookstore/book/title", []string{"Everyday Italian", "Harry Potter", "XQuery Kick Start", "Learning XML"}},
	{"./bookstore/book/author", []string{"Giada De Laurentiis", "J K. Rowling", "James McGovern", "Per Bothner", "Kurt Cagle", "James Linn", "Vaidyanathan Nagarajan", "Erik T. Ray"}},
	{"./bookstore/book/year", []string{"2005", "2005", "2003", "2003"}},
	{"./bookstore/book/p:price", []string{"30.00", "29.99", "49.99", "39.95"}},
	{"./bookstore/book/isbn", nil},

	// descendant queries
	{"//title", []string{"Everyday Italian", "Harry Potter", "XQuery Kick Start", "Learning XML"}},
	{"//book/title", []string{"Everyday Italian", "Harry Potter", "XQuery Kick Start", "Learning XML"}},
	{".//title", []string{"Everyday Italian", "Harry Potter", "XQuery Kick Start", "Learning XML"}},
	{".//bookstore//title", []string{"Everyday Italian", "Harry Potter", "XQuery Kick Start", "Learning XML"}},
	{".//book/title", []string{"Everyday Italian", "Harry Potter", "XQuery Kick Start", "Learning XML"}},
	{".//p:price/.", []string{"30.00", "29.99", "49.99", "39.95"}},
	{".//price", []string{"30.00", "29.99", "49.99", "39.95"}},

	// positional queries
	{"./bookstore/book[1]/title", "Everyday Italian"},
	{"./bookstore/book[4]/title", "Learning XML"},
	{"./bookstore/book[5]/title", nil},
	{"./bookstore/book[3]/author[0]", "James McGovern"},
	{"./bookstore/book[3]/author[1]", "James McGovern"},
	{"./bookstore/book[3]/author[3]/./.", "Kurt Cagle"},
	{"./bookstore/book[3]/author[6]", nil},
	{"./bookstore/book[-1]/title", "Learning XML"},
	{"./bookstore/book[-4]/title", "Everyday Italian"},
	{"./bookstore/book[-5]/title", nil},

	// text queries
	{"./bookstore/book[author='James McGovern']/title", "XQuery Kick Start"},
	{"./bookstore/book[author='Per Bothner']/title", "XQuery Kick Start"},
	{"./bookstore/book[author='Kurt Cagle']/title", "XQuery Kick Start"},
	{"./bookstore/book[author='James Linn']/title", "XQuery Kick Start"},
	{"./bookstore/book[author='Vaidyanathan Nagarajan']/title", "XQuery Kick Start"},
	{"//book[p:price='29.99']/title", "Harry Potter"},
	{"//book[price='29.99']/title", "Harry Potter"},

	// attribute queries
	{"./bookstore/book[@category='WEB']/title", []string{"XQuery Kick Start", "Learning XML"}},
	{"./bookstore/book[@category='COOKING']/title[@lang='en']", "Everyday Italian"},
	{"./bookstore/book/title[@lang='en'][@sku='150']", "Harry Potter"},
	{"./bookstore/book/title[@lang='fr']", nil},

	// parent queries
	{"./bookstore/book[@category='COOKING']/title/../../book[4]/title", "Learning XML"},

	// 'Or' queries
	{"./bookstore/book/title|./bookstore/book/title", []string{"Everyday Italian", "Harry Potter", "XQuery Kick Start", "Learning XML"}},
	{"./bookstore/book/title|./bookstore/book/year", []string{"Everyday Italian", "Harry Potter", "XQuery Kick Start", "Learning XML", "2005", "2005", "2003", "2003"}},

	// bad paths
	{"/bookstore", errorResult("etree: paths cannot be absolute.")},
	{"./bookstore/book[]", errorResult("etree: path contains an empty filter expression.")},
	{"./bookstore/book[@category='WEB'", errorResult("etree: path has invalid filter [brackets].")},
	{"./bookstore/book[@category='WEB]", errorResult("etree: path has mismatched filter quotes.")},
	{"./bookstore/book[author]a", errorResult("etree: path has invalid filter [brackets].")},
}

func TestPath(t *testing.T) {
	doc := NewDocument()
	err := doc.ReadFromString(testXML)
	if err != nil {
		t.Error(err)
	}

	for _, test := range tests {
		path, err := CompilePath(test.path)
		if err != nil {
			if r, ok := test.result.(errorResult); !ok || err.Error() != string(r) {
				failError(t, test, err)
			}
			continue
		}

		// Test both FindElementsPath and FindElementPath
		element := doc.FindElementPath(path)
		elements := doc.FindElementsPath(path)

		switch s := test.result.(type) {
		case errorResult:
			failError(t, test, nil)
		case nil:
			if element != nil || len(elements) != 0 {
				failNil(t, test, elements)
			}
		case string:
			if element == nil || element.Text() != s ||
				len(elements) != 1 || elements[0].Text() != s {
				failString(t, test, element)
			}
		case []string:
			if element == nil || element.Text() != s[0] || len(elements) != len(s) {
				failSlice(t, test, elements)
				continue
			}
			for i := 0; i < len(elements); i++ {
				if elements[i].Text() != s[i] {
					failSlice(t, test, elements)
					break
				}
			}
		}

	}
}

func failError(t *testing.T, test test, got error) {
	t.Errorf("etree: failed test '%s'\n", test.path)
	t.Errorf("Expected error: %s\n", string(test.result.(errorResult)))
	if got == nil {
		t.Error("Got error: nil\n")
	} else {
		t.Errorf("Got error: %s\n", got.Error())
	}
}

func failNil(t *testing.T, test test, got []*Element) {
	t.Errorf("etree: failed test '%s'\n", test.path)
	t.Error("Expected: nil")
	t.Error("Got:")
	for _, e := range got {
		t.Errorf("\t%s", e.Text())
	}
}

func failString(t *testing.T, test test, got *Element) {
	t.Errorf("etree: failed test '%s'\n", test.path)
	t.Errorf("Expected: %s\n", test.result.(string))
	t.Errorf("Got: %s\n", got.Text())
}

func failSlice(t *testing.T, test test, got []*Element) {
	t.Errorf("etree: failed test '%s'\n", test.path)
	t.Error("Expected:")
	for _, s := range test.result.([]string) {
		t.Errorf("\t%s", s)
	}
	t.Error("Got:")
	for _, e := range got {
		t.Errorf("\t%s", e.Text())
	}
}
