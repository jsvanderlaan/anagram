package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"syscall/js"

	"anagram.jurre.dev/utils"
)

func main() {
	c := make(chan struct{})
	js.Global().Set("anagram", js.FuncOf(anagram))
	<-c
}

func anagram(this js.Value, args []js.Value) interface{} {
	start := time.Now()
	input := js.ValueOf(args[0]).String()
	words := js.ValueOf(args[1]).String()

	if len(input) < 2 {
		return []string{}
	}

	fmt.Println("Finding anagrams for:", input)
	fmt.Println("Using  word list with", len(words), "characters")

	freq_dict := make(map[string]float32)
	word_freq := strings.Split(words, "\r\n")
	for _, wf := range word_freq {
		parts := strings.Split(wf, ",")

		if len(parts) == 2 {
			word := parts[0]
			freq, err := strconv.ParseFloat(parts[1], 32)
			if err == nil && freq > 3 && len(word) > 1 {
				freq_dict[word] = float32(freq)
			}
		}
	}

	fmt.Println("Loaded", len(freq_dict), "words")

	output := utils.FastAnagrams(input, freq_dict, 1000, 3)

	fmt.Println("Found", len(output), "anagrams")

	uniqueWords := make(map[string]utils.Anagram)
	for i := 0; i < len(output); i++ {
		str := "\"" + strings.Join(output[i].Words, " ") + "\""
		uniqueWords[str] = output[i]
	}
	output = make([]utils.Anagram, 0, len(uniqueWords))
	for _, v := range uniqueWords {
		output = append(output, v)
	}

	fmt.Println("Filtered to", len(output), "unique anagrams")

	for i := 0; i < len(output); i++ {
		var score float32
		for _, word := range output[i].Words {
			score += freq_dict[word]
		}
		output[i].Score = score
	}
	sort.Slice(output, func(i, j int) bool {
		if output[i].Score == output[j].Score {
			if len(output[i].Words) == len(output[j].Words) {
				return strings.Join(output[i].Words, " ") < strings.Join(output[j].Words, " ")
			}
			return len(output[i].Words) < len(output[j].Words)
		}
		return output[i].Score > output[j].Score
	})

	jsArray := js.ValueOf(make([]interface{}, len(output)))
	for i, v := range output {
		str := "\"" + strings.Join(v.Words, " ") + "\""
		jsArray.SetIndex(i, str)
	}

	fmt.Println("Outputting", len(output), "anagrams")

	fmt.Println("Processed in", time.Since(start))
	return jsArray
}
