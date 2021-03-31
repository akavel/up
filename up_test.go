package main

import "testing"

func Test_Editor_insert(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		tt.e.insert(tt.insert...)
		if string(tt.e.value) != string(tt.wantValue) {
			t.Errorf("%q: bad value\nwant: %q\nhave: %q", tt.comment, tt.wantValue, tt.e.value)
		}
	}
}

func Test_Editor_unix_word_rubout(t *testing.T) {
	tests := []struct {
		comment       string
		e             Editor
		wantValue     []rune
		wantKillspace []rune
	}{
		{
			comment: "unix-word-rubout at beginning of line",
			e: Editor{
				value:  []rune(`abc`),
				cursor: 0,
			},
			wantValue:     []rune(`abc`),
			wantKillspace: []rune(``),
		},
		{
			comment: "unix-word-rubout at soft beginning of line",
			e: Editor{
				value:  []rune(` abc`),
				cursor: 1,
			},
			wantValue:     []rune(`abc`),
			wantKillspace: []rune(` `),
		},
		{
			comment: "unix-word-rubout until soft beginning of line",
			e: Editor{
				value:  []rune(` abc`),
				cursor: 2,
			},
			wantValue:     []rune(` bc`),
			wantKillspace: []rune(`a`),
		},
		{
			comment: "unix-word-rubout until beginning of line",
			e: Editor{
				value:  []rune(`abc`),
				cursor: 2,
			},
			wantValue:     []rune(`c`),
			wantKillspace: []rune(`ab`),
		},
		{
			comment: "unix-word-rubout in middle of line",
			e: Editor{
				value:  []rune(`lorem ipsum dolor`),
				cursor: 11,
			},
			wantValue:     []rune(`lorem  dolor`),
			wantKillspace: []rune(`ipsum`),
		},
		{
			comment: "unix-word-rubout cursor at beginning of word",
			e: Editor{
				value:  []rune(`lorem ipsum dolor`),
				cursor: 12,
			},
			wantValue:     []rune(`lorem dolor`),
			wantKillspace: []rune(`ipsum `),
		},
		{
			comment: "unix-word-rubout cursor between multiple spaces",
			e: Editor{
				value:  []rune(`a b   c`),
				cursor: 5,
			},
			wantValue:     []rune(`a  c`),
			wantKillspace: []rune(`b  `),
		},
		{
			comment: "unix-word-rubout tab as space char (although is it a realistic case in the context of a command line instruction?)",
			e: Editor{
				value: []rune(`a b		c`),
				cursor: 5,
			},
			wantValue: []rune(`a c`),
			wantKillspace: []rune(`b		`),
		},
	}

	for _, tt := range tests {
		tt.e.unixWordRubout()
		if string(tt.e.value) != string(tt.wantValue) {
			t.Errorf("%q: bad value\nwant: %q\nhave: %q", tt.comment, tt.wantValue, tt.e.value)
		}
		if string(tt.e.killspace) != string(tt.wantKillspace) {
			t.Errorf("%q: bad value in killspace\nwant: %q\nhave: %q", tt.comment, tt.wantKillspace, tt.e.value)
		}
	}
}
