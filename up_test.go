package main

import "testing"

func Test_Editor_insert(test *testing.T) {
	cases := []struct {
		comment   string
		e         Editor
		insert    []rune
		wantValue []rune
	}{
		{
			comment: "prepend ASCII char",
			e: Editor{
				value:  []rune(`abc`),
				cursor: 0,
			},
			insert:    []rune{'X'},
			wantValue: []rune(`Xabc`),
		},
		{
			comment: "prepend UTF char",
			e: Editor{
				value:  []rune(`abc`),
				cursor: 0,
			},
			insert:    []rune{'☃'},
			wantValue: []rune(`☃abc`),
		},
		{
			comment: "insert ASCII char",
			e: Editor{
				value:  []rune(`abc`),
				cursor: 1,
			},
			insert:    []rune{'X'},
			wantValue: []rune(`aXbc`),
		},
		{
			comment: "insert UTF char",
			e: Editor{
				value:  []rune(`abc`),
				cursor: 1,
			},
			insert:    []rune{'☃'},
			wantValue: []rune(`a☃bc`),
		},
		{
			comment: "append ASCII char",
			e: Editor{
				value:  []rune(`abc`),
				cursor: 3,
			},
			insert:    []rune{'X'},
			wantValue: []rune(`abcX`),
		},
		{
			comment: "append UTF char",
			e: Editor{
				value:  []rune(`abc`),
				cursor: 3,
			},
			insert:    []rune{'☃'},
			wantValue: []rune(`abc☃`),
		},
		{
			comment: "insert 2 ASCII chars",
			e: Editor{
				value:  []rune(`abc`),
				cursor: 1,
			},
			insert:    []rune{'X', 'Y'},
			wantValue: []rune(`aXYbc`),
		},
	}

	for _, c := range cases {
		c.e.insert(c.insert...)
		if string(c.e.value) != string(c.wantValue) {
			test.Errorf("%q: bad value\nwant: %q\nhave: %q", c.comment, c.wantValue, c.e.value)
		}
	}
}
