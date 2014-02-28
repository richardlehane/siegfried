package pronom

import "testing"

type testpattern struct {
	name    string
	pattern string
}

var good = []testpattern{
	{
		"dbf",
		"02{2}[01:1C][01:1F]????[00:03]([41:5A]|[61:7A]){10}(43|4E|4C)",
	},
	{
		"dcx",
		"0004000000000000000000000A00[!00]{1009}03000200FFFFFFFFFFFFFFFF{1}01FFFF{2}0F0F{491}[41:5A]000A{502}2B00[0B:0F]0000002B00{512}[!00]00[!00]00FFFFFFFFFFFFFFFF",
	},
	{
		"igs",
		"53(202020202020|303030303030)31(0D0A|0A){72}(5320202020202032|5330303030303032|4720202020202031|4730303030303031)",
	},
	{
		"mif",
		"56(45|65)(52|72)(53|73)(49|69)??(4E|6E){5-6}(43|63)(48|68)(41|61)(52|72)(53|73)(45|65)(54|74)*43(4F|6F)(4C|6C)(55|75)(4D|6D)(4E|6E)(53|73)",
	},
	{
		"zip",
		"504B01{43-65531}504B0506{18-65531}",
	},
	{
		"ani",
		"52494646{4}41434F4E{0-*}616E69682400000024000000[!00]*4C495354{4}6672616D69636F6E",
	},
	{
		"cel",
		"1991[!4001C80000000000]",
	},
	{
		"notrange",
		"1991[!01:02]", // made this up, haven't seen any in the wild with a not range
	},
	{
		"notchoice",
		"1991(!65|[!01:02])",
	},
	{
		"rangechoice",
		"1991(65|[01:02])",
	},
}

var bad = []testpattern{
	{
		"badchar",
		"1991[!4001C80000000000]y",
	},
	{
		"badrange",
		"1991[:61]",
	},
	{
		"singlewild",
		"1991?ACCD",
	},
	{
		"badwild",
		"1991{ABCD}ABDC",
	},
	{
		"doublenegative",
		"1991[!!ABCD]",
	},
	{
		"unclosed",
		"1991[!4001",
	},
}

var goodContainer = []testpattern{
	{
		"OOXML",
		`'C'00 'o'00 'n'00 't'00 'e'00 'n'00 't'00 'T'00 'y'00 'p'00 'e'00 '='00 '"'00 'a'00 'p'00 'p'00 'l'00 'i'00 'c'00 'a'00 't'00 'i'00 'o'00 'n'00 '/'00 'v'00 'n'00 'd'00 '.'00 'o'00 'p'00 'e'00 'n'00 'x'00 'm'00 'l'00 'f'00 'o'00 'r'00 'm'00 'a'00 't'00 's'00 '-'00 'o'00 'f'00 'f'00 'i'00 'c'00 'e'00 'd'00 'o'00 'c'00 'u'00 'm'00 'e'00 'n'00 't'00 '.'00 'w'00 'o'00 'r'00 'd'00 'p'00 'r'00 'o'00 'c'00 'e'00 's'00 's'00 'i'00 'n'00 'g'00 'm'00 'l'00 '.'00 'd'00 'o'00 'c'00 'u'00 'm'00 'e'00 'n'00 't'00 '.'00 'm'00 'a'00 'i'00 'n'00 '+'00 'x'00 'm'00 'l'00 '"'00`,
	},
	{
		"WORD",
		`10 00 00 00 'Word.Document.' ['6'-'7'] 00`,
	},
	{
		"ODT",
		`'office:version=' [22 27] '1.0' [22 27]`,
	},
	{
		"VISIO",
		`'Visio (TM) Drawing'0D0A`,
	},
}

func TestGood(t *testing.T) {
	for _, v := range good {
		l := sigLex(v.name, v.pattern)
		for i := l.nextItem(); i.typ != itemEOF; i = l.nextItem() {
			if i.typ == itemError {
				t.Error(i)
				break
			}
		}
	}
}

func TestBad(t *testing.T) {
	for _, v := range bad {
		l := sigLex(v.name, v.pattern)
		i := l.nextItem()
		for ; i.typ != itemEOF; i = l.nextItem() {
			if i.typ == itemError {
				break
			}
		}
		if i.typ != itemError {
			t.Error(i)
		}
	}
}

func TestContainerGood(t *testing.T) {
	for _, v := range goodContainer {
		l := conLex(v.name, v.pattern)
		for i := l.nextItem(); i.typ != itemEOF; i = l.nextItem() {
			if i.typ == itemError {
				t.Error(i)
				break
			}
		}
	}
}
