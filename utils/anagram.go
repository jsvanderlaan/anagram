package utils

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Anagram struct {
	Words []string
	Score float32 // sum of word frequencies (higher = more "frequent" words)
}

// FindAnagrams finds all anagrams of `input` that can be formed using the words in wordFreq.
// - input: free text (will be normalized to lower-case ASCII a-z).
// - wordFreq: map[word]frequency (float32).
// - maxResults: if >0, stop after collecting maxResults results (helps with explosion).
//
// Returns results sorted by Score descending (most frequent words first).
//
// NOTE: This function treats words as re-usable if letters allow (a word may appear multiple times,
// e.g. "aa" could be two occurrences of "a-word" if present in the dictionary). It enforces a
// non-decreasing candidate index order to avoid generating different permutations of the same combination.
func FindAnagrams(input string, wordFreq map[string]float32, maxResults int) ([]Anagram, error) {
	// --- normalize input: keep only a-z ASCII letters, lowercase
	normIn := normalizeASCIIletters(input)
	if len(normIn) == 0 {
		return nil, fmt.Errorf("input contains no a-z letters after normalization")
	}
	inputCounts, inputMask, inputTotal := countsFromASCII(normIn)

	// --- prefilter dictionary into candidates
	type cand struct {
		word    string
		counts  [26]uint8
		mask    uint32
		letters int
		freq    float32
	}
	cands := make([]cand, 0, len(wordFreq))
	for w, f := range wordFreq {
		nw := normalizeASCIIletters(w)
		if nw == "" {
			continue
		}
		wCounts, wMask, wTotal := countsFromASCII(nw)
		// quick mask test: if word uses letters outside input, skip
		if wMask & ^inputMask != 0 {
			continue
		}
		// counts must not exceed input counts
		if !countsLE(&wCounts, &inputCounts) {
			continue
		}
		cands = append(cands, cand{
			word:    w,
			counts:  wCounts,
			mask:    wMask,
			letters: wTotal,
			freq:    f,
		})
	}
	if len(cands) == 0 {
		return nil, nil
	}

	// sort candidates: prefer higher frequency and longer words — helps pruning & finds higher-score combos earlier
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].freq == cands[j].freq {
			return cands[i].letters > cands[j].letters
		}
		return cands[i].freq > cands[j].freq
	})

	// --- recursive DP (memoization) to find completions for a remaining counts state
	// key: countsKey(remaining) + "|" + startIndex
	type seq []int
	memo := make(map[string][][]int) // map[key] -> list of completions (each completion is []int indices into cands)

	// helper: remaining => string key (26 bytes)
	countsKey := func(c *[26]uint8) string {
		var b [26]byte
		for i := 0; i < 26; i++ {
			b[i] = byte(c[i])
		}
		return string(b[:])
	}

	var completions func(start int, remaining *[26]uint8, remainingTotal int) [][]int
	completions = func(start int, remaining *[26]uint8, remainingTotal int) [][]int {
		if remainingTotal == 0 {
			// empty completion (no words needed) => single empty sequence
			return [][]int{{}}
		}
		key := countsKey(remaining) + "|" + strconv.Itoa(start)
		if v, ok := memo[key]; ok {
			return v
		}
		var res [][]int
		// iterate candidates from start..end (non-decreasing index ensures no permutations)
		for i := start; i < len(cands); i++ {
			c := &cands[i]
			if c.letters > remainingTotal {
				// candidate longer than remaining letters -> skip (candidates sorted, but not guaranteed strictly by letters)
				continue
			}
			// quick mask subset check (fast)
			if c.mask & ^countsMask(remaining) != 0 {
				continue
			}
			if !countsLE(&c.counts, remaining) {
				continue
			}
			// choose c
			subtractCountsInplace(remaining, &c.counts)
			sub := completions(i, remaining, remainingTotal-c.letters)
			// restore
			addCountsInplace(remaining, &c.counts)
			if len(sub) == 0 {
				continue
			}
			for _, seq := range sub {
				// prepend i to seq
				newSeq := make([]int, 1+len(seq))
				newSeq[0] = i
				copy(newSeq[1:], seq)
				res = append(res, newSeq)
				// optional limit: if global limit hit we can stop early (not strictly 'all' then)
			}
		}
		memo[key] = res
		return res
	}

	// Run DP
	allSeqs := completions(0, &inputCounts, inputTotal)

	// Convert sequences into Anagram results (words + score)
	results := make([]Anagram, 0, len(allSeqs))
	for _, s := range allSeqs {
		var score float32
		words := make([]string, len(s))
		for i, idx := range s {
			words[i] = cands[idx].word
			score += cands[idx].freq
		}
		results = append(results, Anagram{
			Words: words,
			Score: score,
		})
		// optional early stop by maxResults
		if maxResults > 0 && len(results) >= maxResults {
			break
		}
	}

	// sort results by Score desc (most frequent words combinations first).
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			// tie-breaker: fewer words first (prefer more compact phrases)
			if len(results[i].Words) == len(results[j].Words) {
				return strings.Join(results[i].Words, " ") < strings.Join(results[j].Words, " ")
			}
			return len(results[i].Words) < len(results[j].Words)
		}
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// -------------------- helpers --------------------

// normalizeASCIIletters returns a lower-case a-z-only string (drops all other characters).
func normalizeASCIIletters(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b.WriteByte(c - 'A' + 'a')
		} else if c >= 'a' && c <= 'z' {
			b.WriteByte(c)
		}
		// drop everything else (spaces, punctuation, diacritics)
	}
	return b.String()
}

func countsFromASCII(s string) ([26]uint8, uint32, int) {
	var cnt [26]uint8
	var mask uint32
	total := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 'a' || c > 'z' {
			continue
		}
		idx := c - 'a'
		cnt[idx]++
		mask |= 1 << idx
		total++
	}
	return cnt, mask, total
}

func countsLE(a, b *[26]uint8) bool {
	// return true if a[i] <= b[i] for all i
	for i := 0; i < 26; i++ {
		if a[i] > b[i] {
			return false
		}
	}
	return true
}

func subtractCountsInplace(dst, delta *[26]uint8) {
	for i := 0; i < 26; i++ {
		dst[i] -= delta[i]
	}
}

func addCountsInplace(dst, delta *[26]uint8) {
	for i := 0; i < 26; i++ {
		dst[i] += delta[i]
	}
}

func countsMask(c *[26]uint8) uint32 {
	var mask uint32
	for i := 0; i < 26; i++ {
		if c[i] != 0 {
			mask |= 1 << i
		}
	}
	return mask
}
