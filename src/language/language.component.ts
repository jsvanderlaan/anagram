import { NgClass } from '@angular/common';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { LanguageService } from '../services/language.service';
import { Language } from '../types';

@Component({
    standalone: true,
    selector: 'language',
    templateUrl: './language.component.html',
    imports: [NgClass],
    changeDetection: ChangeDetectionStrategy.OnPush,
})
export class LanguageComponent {
    readonly languages = Object.values(Language);
    public readonly language = inject(LanguageService);

    onLanguageChange(lang: Language): void {
        this.language.set(lang);
    }
}
