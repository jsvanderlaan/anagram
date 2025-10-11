package utils

import (
	"runtime"
	"sort"
	"strings"
	"sync"
)

type partial struct {
	remaining [26]uint8
	words     []string
	score     float32
}

// ---- Helper functions ----

func wordCounts(s string) (counts [26]uint8, mask uint32) {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			idx := r - 'a'
			counts[idx]++
			mask |= 1 << idx
		} else if r >= 'A' && r <= 'Z' {
			idx := r - 'A'
			counts[idx]++
			mask |= 1 << idx
		}
	}
	return
}

// ---- Main beam search ----

func beamAnagrams(cands []wordData, input [26]uint8, beamWidth, maxDepth int) []Anagram {
	beams := []partial{{remaining: input}}
	results := []Anagram{}

	for depth := 0; depth < maxDepth; depth++ {
		next := make([]partial, 0, beamWidth*4)

		for _, p := range beams {
			for _, c := range cands {
				if !countsLE(&c.counts, &p.remaining) {
					continue
				}
				newRem := p.remaining
				for j := 0; j < 26; j++ {
					newRem[j] -= c.counts[j]
				}
				newWords := append(append([]string(nil), p.words...), c.word)
				newScore := p.score + c.score

				if allZero(&newRem) {
					results = append(results, Anagram{Words: newWords, Score: newScore})
					continue
				}
				next = append(next, partial{
					remaining: newRem,
					words:     newWords,
					score:     newScore,
				})
			}
		}

		if len(next) == 0 {
			break
		}

		sort.Slice(next, func(i, j int) bool {
			return next[i].score > next[j].score
		})
		if len(next) > beamWidth {
			next = next[:beamWidth]
		}

		beams = next
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	return results
}

// ---- Public API ----

func FindAnagramsBeam(input string, words map[string]float32, maxDepth int, beamWidth int) []Anagram {
	input = strings.ToLower(input)
	inCounts, _ := wordCounts(input)

	// Filter candidates early
	cands := make([]wordData, 0, len(words))
	for w, freq := range words {
		if len(w) > len(input) {
			continue
		}
		c, m := wordCounts(w)
		if !countsLE(&c, &inCounts) {
			continue
		}
		score := freq * float32(len(w))
		cands = append(cands, wordData{word: w, counts: c, mask: m, score: score})
	}

	// Sort candidates descending by score
	sort.Slice(cands, func(i, j int) bool { return cands[i].score > cands[j].score })

	// Parallelize across CPU cores
	numCPU := runtime.NumCPU()
	chunk := (len(cands) + numCPU - 1) / numCPU

	var wg sync.WaitGroup
	resultsChan := make(chan []Anagram, numCPU)

	for i := 0; i < numCPU; i++ {
		start := i * chunk
		if start >= len(cands) {
			break
		}
		end := start + chunk
		if end > len(cands) {
			end = len(cands)
		}
		part := cands[start:end]

		wg.Add(1)
		go func(p []wordData) {
			defer wg.Done()
			res := beamAnagrams(p, inCounts, beamWidth, maxDepth)
			resultsChan <- res
		}(part)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	allResults := []Anagram{}
	for r := range resultsChan {
		allResults = append(allResults, r...)
	}

	sort.Slice(allResults, func(i, j int) bool { return allResults[i].Score > allResults[j].Score })
	return allResults
}
