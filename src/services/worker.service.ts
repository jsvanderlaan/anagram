import { Injectable } from '@angular/core';
import { combineLatest, debounceTime, distinctUntilChanged, Subject } from 'rxjs';
import { AnswersService } from './answers.service';
import { LanguageService } from './language.service';

@Injectable({ providedIn: 'root' })
export class WorkerService {
    private readonly _anagram: Subject<{ input: string }> = new Subject();
    private _worker: Worker | null = null;

    constructor(private readonly _answers: AnswersService, private readonly _language: LanguageService) {
        combineLatest([
            this._anagram.pipe(debounceTime(200), distinctUntilChanged(deepEqual)),
            this._language.language$,
        ]).subscribe(([next, lang]) => this._process({ ...next, lang }));
    }

    anagram(input: string): void {
        this._anagram.next({ input });
    }

    private _process(d: any): void {
        this._answers.loading.set(true);
        if (this._worker !== null) {
            this._worker.terminate();
        }

        this._worker = new Worker(new URL('../app/app.worker', import.meta.url), { type: 'module' });
        this._worker.onmessage = ({ data }) => {
            this._answers.loading.set(false);
            this._answers.list.set(data);
        };
        this._worker.postMessage(d);
    }
}

var deepEqual = function (x: any, y: any): boolean {
    if (x === y) {
        return true;
    } else if (typeof x == 'object' && x != null && typeof y == 'object' && y != null) {
        if (Object.keys(x).length != Object.keys(y).length) return false;

        for (var prop in x) {
            if (y.hasOwnProperty(prop)) {
                if (!deepEqual(x[prop], y[prop])) return false;
            } else return false;
        }

        return true;
    } else return false;
};
