/// <reference lib="webworker" />

import { Language } from '../types';

addEventListener('message', async ({ data }) => {
    // time
    console.time('worker');
    const language: Language = data.lang;
    const original = data.input.toUpperCase().normalize('NFD').replace(/[^\w\s]|\d|_/g, '').replace(/\s+/g, '').split('').sort();
    const wordLength = original.length;

    if (wordLength < 2) {
        postMessage([]);
        return;
    }

    const all = (await (await fetch(`./woorden/answers_${language}.json`)).json()) as Record<string, string[]>;
    const anagrams = findMultiWordAnagrams(original, all);
    console.timeStamp('worker');
    const filtered = anagrams.filter(arr => !(arr.length === 1 && arr[0].toUpperCase() === original));
    const unique: string[][] = [];
    const seen = new Set<string>();
    for (const arr of filtered) {
        const sortedArr = [...arr].sort().join(' ');
        if (!seen.has(sortedArr)) {
            seen.add(sortedArr);
            const sorted = arr.sort((a, b) => b.length - a.length || a.localeCompare(b));
            unique.push(sorted);
        }
    }

    const sorted = unique.sort((a, b) => {
        // sort by number of words (ascending)
        if (a.length !== b.length) {
            return a.length - b.length;
        }
        // then by total length of words (descending)
        const aTotalLength = a.reduce((sum, word) => sum + word.length, 0);
        const bTotalLength = b.reduce((sum, word) => sum + word.length, 0);
        if (aTotalLength !== bTotalLength) {
            return bTotalLength - aTotalLength;
        }
        // finally alphabetically
        return a.join(' ').localeCompare(b.join(' '));
    });
    console.timeEnd('worker');

    postMessage(sorted.map(arr => '"' + arr.join(' ') + '"'));
});

function findMultiWordAnagrams(input: string[], words: Record<string, string[]>): string[][] {
    const results: string[][] = [];
    const validWords = getValidWords(input, words);
    const memo = new Map<string, string[][]>();
    const maxWords = 3; //Math.min(5, Math.floor(input.length / 3));

    function backtrack(remainingChars: string[], path: string[]): void {
        const key = remainingChars.join('') + '|' + path.join(' ');
        if (memo.has(key)) {
            results.push(...memo.get(key)!);
            return;
        }

        if (path.length > maxWords) return;
        if (remainingChars.length === 0) {
            results.push([...path]);
            memo.set(key, [[...path]]);
            return;
        }

        const localResults: string[][] = [];
        for (const [length, wordList] of Object.entries(validWords)) {
            if (+length > remainingChars.length) continue;

            for (const word of wordList) {
                const wordChars = word.split('');
                let tempChars = [...remainingChars];
                let canUseWord = true;

                for (const char of wordChars) {
                    const index = tempChars.indexOf(char);
                    if (index === -1) {
                        canUseWord = false;
                        break;
                    }
                    tempChars.splice(index, 1);
                }

                if (canUseWord) {
                    path.push(word);
                    backtrack(tempChars, path);
                    path.pop();
                }
            }
        }

        memo.set(key, localResults);
    }

    backtrack(input, []);

    return results;
}

function getValidWords(inputChars: string[], words: Record<string, string[]>): Record<string, string[]> {
    const inputCharCount = countChars(inputChars);
    const validWords: Record<string, string[]> = {};

    for (const [length, wordList] of Object.entries(words)) {
        validWords[length] = wordList.filter(word => {
            const wordCharCount = countChars(word.split(''));
            return Object.keys(wordCharCount).every(char => (inputCharCount[char] || 0) >= wordCharCount[char]);
        });
    }

    return validWords;
}

function countChars(chars: string[]): Record<string, number> {
    return chars.reduce((count, char) => {
        count[char] = (count[char] || 0) + 1;
        return count;
    }, {} as Record<string, number>);
}
const memo = new Map<string, string[][]>();
