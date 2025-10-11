/// <reference lib="webworker" />

import { setupWasm } from 'src/wasm/wasm_exec';
import { Language } from '../types';

let initialized = false;
let words: Record<Language, string> = {
    nl: '',
    gb: '',
};

addEventListener(
    'message',
    async ({ data: { command, payload } }: { data: { command: string; payload: { lang: Language; input: string } } }) => {
        if (!initialized) {
            if (!initialized) {
                await loadWasm();
                initialized = true;
            }
            postMessage({ status: 'initialized' });
        }

        try {
            if (command !== 'anagram') {
                throw new Error(`Unknown command: ${command}`);
            }
            if (words[payload.lang] === '') {
                words[payload.lang] = await (await fetch(`./woorden/answers_${payload.lang}.txt`)).text();
            }

            const result = (self as any).anagram(payload.input, words[payload.lang]);
            postMessage({ status: 'completed', result });
        } catch (error: any) {
            postMessage({ status: 'error', error: error.message });
        }
        return;
    }
);

async function loadWasm(): Promise<void> {
    setupWasm();
    const go = new Go();
    const result = await WebAssembly.instantiateStreaming(fetch('main.wasm'), go.importObject);
    go.run(result.instance);
}
