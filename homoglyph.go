package main

import "fmt"

func getAlikeNames(name string) []string {
	res := []string{}
	// `ᅟᅠ         　ㅤǃ！״″＂＄％＆＇﹝（﹞）⁎＊＋‚，‐𐆑－٠۔܁܂․‧。．｡⁄∕╱⫻⫽／ﾉΟοОоՕ𐒆ＯｏΟοОоՕ𐒆Ｏｏا
	//１２３４５６𐒇７Ց８９։܃܄∶꞉：;；‹＜𐆐＝›＞？＠［＼］＾＿｀
	// ÀÁÂÃÄÅàáâãäåɑΑαаᎪＡａßʙΒβВЬᏴᛒＢｂϲϹСсᏟⅭⅽ𐒨ＣｃĎďĐđԁժᎠⅮⅾＤｄÈÉÊËéêëĒēĔĕĖėĘĚěΕЕеᎬＥｅϜＦｆɡɢԌնᏀＧｇʜΗНһᎻＨｈɩΙІіاᎥᛁⅠⅰ𐒃ＩｉϳЈјյᎫＪｊΚκКᏦᛕKＫｋʟιاᏞⅬⅼＬｌΜϺМᎷᛖⅯⅿＭｍɴΝＮｎΟοОоՕ𐒆ＯｏΟοОоՕ𐒆ＯｏΡρРрᏢＰｐႭႳＱｑʀԻᏒᚱＲｒЅѕՏႽᏚ𐒖ＳｓΤτТᎢＴｔμυԱՍ⋃ＵｕνѴѵᏙⅤⅴＶｖѡᎳＷｗΧχХхⅩⅹＸｘʏΥγуҮＹｙΖᏃＺｚ｛ǀا｜｝⁓～ӧӒӦ`
	glyph := map[rune][]string{
		'1': {"１"},
		'2': {"２"},
		'3': {"３"},
		'4': {"４"},
		'5': {"５"},
		'6': {"６"},
		'7': {"𐒇", "７"},
		'8': {"Ց", "&", "８"},
		'9': {"９"},
		'0': {"Ο", "ο", "О", "о", "Օ", "𐒆", "Ｏ", "ｏ", "Ο", "ο", "О", "о", "Օ", "𐒆", "Ｏ", "ｏ"},
		'a': {"à", "á", "â", "ã", "ä", "å", "а", "ɑ", "α", "а", "ａ"},
		'A': {"À", "Á", "Â", "Ã", "Ä", "Å", "Ꭺ", "Ａ"},
		'b': {"d", "lb", "ib", "ʙ", "Ь", "ｂ", "ß"},
		'B': {"ß", "Β", "β", "В", "Ь", "Ᏼ", "ᛒ", "Ｂ"},
		'c': {"ϲ", "с", "ⅽ"},
		'C': {"Ϲ", "С", "с", "Ꮯ", "Ⅽ", "ⅽ", "𐒨", "Ｃ"},
		'd': {"b", "cl", "dl", "di", "ԁ", "ժ", "ⅾ", "ｄ"},
		'e': {"é", "ê", "ë", "ē", "ĕ", "ė", "ｅ", "е"},
		'f': {"Ϝ", "Ｆ", "ｆ"},
		'g': {"q", "ɢ", "ɡ", "Ԍ", "Ԍ", "ｇ"},
		'h': {"lh", "ih", "һ", "ｈ"},
		'i': {"1", "l", "Ꭵ", "ⅰ", "ｉ"},
		'j': {"ј", "ｊ"},
		'k': {"lk", "ik", "lc", "κ", "ｋ"},
		'l': {"1", "i", "ⅼ", "ｌ", "ӏ"},
		'm': {"n", "nn", "rn", "rr", "ⅿ", "ｍ"},
		'n': {"m", "r", "ｎ"},
		'o': {"0", "Ο", "ο", "О", "о", "Օ", "𐒆", "Ｏ", "ｏ", "Ο", "ο", "О", "о", "Օ", "𐒆", "Ｏ", "ｏ"},
		'p': {"ρ", "р", "ｐ", "р"},
		'q': {"g", "ｑ"},
		'r': {"ʀ", "ｒ"},
		's': {"Ⴝ", "Ꮪ", "Ｓ", "ｓ"},
		't': {"τ", "ｔ"},
		'u': {"μ", "υ", "Ս", "Ｕ", "ｕ"},
		'v': {"ｖ", "ѵ", "ⅴ"},
		'w': {"vv", "ѡ", "ｗ"},
		'x': {"ⅹ", "ｘ"},
		'y': {"ʏ", "γ", "у", "Ү", "ｙ"},
		'z': {"ｚ"},
	}
	dom := []rune(name)

	for ws := range dom {
		for i := 0; i < (len(dom) - ws); i++ {
			win := dom[i : i+ws]

			j := 0
			for j < ws {
				c := rune(win[j])
				if repList, ok := glyph[c]; ok {
					for _, rep := range repList {
						g := []rune(rep)

						win = []rune(fmt.Sprintf("%s%s%s", string(win[:j]), string(g), string(win[j+1:])))
						if len(g) > 1 {
							j++
						}
						fuzzed := fmt.Sprintf("%s%s%s", string(dom[:i]), string(win), string(dom[i+ws:]))
						res = append(res, fuzzed)
					}
				}
				j++
			}
		}
	}

	return Dedup(res)
}

// Dedup a slice of string
func Dedup(s []string) []string {
	for i := 0; i < len(s); i++ {
		for i2 := i + 1; i2 < len(s); i2++ {
			if s[i] == s[i2] {
				// delete
				s = append(s[:i2], s[i2+1:]...)
				i2--
			}
		}
	}
	return s
}
