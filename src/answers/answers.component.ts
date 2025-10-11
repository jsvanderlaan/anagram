import { NgClass } from '@angular/common';
import { ChangeDetectionStrategy, Component, computed, effect, inject, input, signal } from '@angular/core';
import { AnswersService } from '../services/answers.service';
import { LanguageService } from '../services/language.service';

@Component({
    selector: 'answers',
    standalone: true,
    imports: [NgClass],
    changeDetection: ChangeDetectionStrategy.OnPush,
    templateUrl: './answers.component.html',
})
export class AnswersComponent {
    readonly previewSize = 100;
    readonly inputText = input<string>('');
    readonly highlight = input<string>('');
    readonly sa = input<boolean>(false);
    readonly clickedWord = signal<string | null>(null);
    readonly wordToHighlight = computed(() => this.clickedWord() || this.highlight());
    readonly showAll = signal<boolean>(false);
    readonly answers = inject(AnswersService);
    readonly language = inject(LanguageService);
    readonly preview = computed(() => [...this.answers.list()].splice(0, this.previewSize));
    readonly copiedUrl = signal<string | null>(null);

    constructor() {
        effect(() => this.showAll.set(this.sa()));
    }

    copyUrl(word: string): void {
        this.clickedWord.set(word);
        const url = `${window.location.origin}${window.location.pathname}?q=${encodeURIComponent(
            this.inputText()
        )}&lang=${this.language.language()}&highlight=${encodeURIComponent(word)}&sa=${this.showAll() ? '1' : '0'}`;
        navigator.clipboard.writeText(url);
        this.showCopiedMessage(url);
    }

    showCopiedMessage(url: string) {
        this.copiedUrl.set(url);
        setTimeout(() => {
            this.copiedUrl.set(null);
        }, 2000);
    }
}
