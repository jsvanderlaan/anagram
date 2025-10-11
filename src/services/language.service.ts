import { Injectable } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { BehaviorSubject, map, skip } from 'rxjs';
import { Language } from '../types';

@Injectable({ providedIn: 'root' })
export class LanguageService {
    static readonly defaultLang = Language.NL;
    readonly language$ = new BehaviorSubject<Language>(LanguageService.defaultLang);
    readonly language = toSignal(this.language$, {
        initialValue: LanguageService.defaultLang,
    });
    readonly onLanguageChange = this.language$.pipe(
        skip(1),
        map(() => null)
    );

    set(lang: Language): void {
        this.language$.next(lang);
    }
}
