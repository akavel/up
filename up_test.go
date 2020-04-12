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
