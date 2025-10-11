package utils

import (
	"sort"
	"sync"
)

// Prefilter dictionary
type wordData struct {
	word   string
	counts [26]uint8
	mask   uint32
	score  float32
}

// FastAnagrams finds multi-word anagrams efficiently.
// It limits depth to maxWords (e.g. 3) to reduce combinatorial explosion.
func FastAnagrams(input string, wordFreq map[string]float32, maxResults int, maxWords int) []Anagram {
	input = normalizeASCIIletters(input)
	if len(input) == 0 {
		return nil
	}

	// Count input letters
	counts, mask, _ := countsFromASCII(input)

	cands := make([]wordData, 0, len(wordFreq))
	for w, f := range wordFreq {
		nw := normalizeASCIIletters(w)
		if nw == "" {
			continue
		}
		c, m, _ := countsFromASCII(nw)
		if m&^mask != 0 {
			continue
		}
		if !countsLE(&c, &counts) {
			continue
		}
		cands = append(cands, wordData{w, c, m, f})
	}

	if len(cands) == 0 {
		return nil
	}

	// Sort: higher frequency + longer words first (better pruning)
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].score == cands[j].score {
			return len(cands[i].word) > len(cands[j].word)
		}
		return cands[i].score > cands[j].score
	})

	// Parallel search
	workers := 8
	var wg sync.WaitGroup
	resultsCh := make(chan Anagram, 1000)
	for wi := 0; wi < workers; wi++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			var rem [26]uint8
			copy(rem[:], counts[:])
			local := make([]Anagram, 0, 128)
			for i := start; i < len(cands); i += workers {
				searchAnagrams(&cands, i, rem, 0, nil, 0, 0, maxWords, &local, maxResults)
			}
			for _, r := range local {
				resultsCh <- r
			}
		}(wi)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	results := make([]Anagram, 0, 512)
	for r := range resultsCh {
		results = append(results, r)
		if maxResults > 0 && len(results) >= maxResults {
			break
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return len(results[i].Words) < len(results[j].Words)
		}
		return results[i].Score > results[j].Score
	})

	if maxResults > 0 && len(results) > maxResults {
		results = results[:maxResults]
	}
	return results
}

// Recursive DFS with max depth (maxWords)
func searchAnagrams(cands *[]wordData, idx int, remaining [26]uint8, usedMask uint32, phrase []string, score float32, depth, maxDepth int, out *[]Anagram, maxResults int) {
	if allZero(&remaining) {
		*out = append(*out, Anagram{
			Words: append([]string(nil), phrase...),
			Score: score,
		})
		return
	}
	if depth >= maxDepth {
		return
	}

	for i := idx; i < len(*cands); i++ {
		c := &(*cands)[i]

		// Fast skip if word doesn’t fit
		canUse := true
		for j := 0; j < 26; j++ {
			if c.counts[j] > remaining[j] {
				canUse = false
				break
			}
		}
		if !canUse {
			continue
		}

		// Subtract
		newRem := remaining
		for j := 0; j < 26; j++ {
			newRem[j] -= c.counts[j]
		}

		// Recurse with word added
		newPhrase := append(phrase, c.word)
		searchAnagrams(cands, i, newRem, usedMask|c.mask, newPhrase, score+c.score, depth+1, maxDepth, out, maxResults)

		if maxResults > 0 && len(*out) >= maxResults {
			return
		}
	}
}

func allZero(c *[26]uint8) bool {
	for _, v := range c {
		if v != 0 {
			return false
		}
	}
	return true
}
