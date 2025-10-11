import { Injectable } from '@angular/core';
import { combineLatest, debounceTime, distinctUntilChanged, filter, from, Subject, switchMap } from 'rxjs';
import { AnswersService } from './answers.service';
import { LanguageService } from './language.service';

@Injectable({ providedIn: 'root' })
export class WorkerService {
    private readonly _anagram: Subject<{ input: string }> = new Subject();
    private _worker: Worker | null = null;

    constructor(
        private readonly _answers: AnswersService,
        private readonly _language: LanguageService
    ) {
        combineLatest([
            this._anagram.pipe(
                debounceTime(200),
                distinctUntilChanged(deepEqual),
                filter(i => i.input.length > 1)
            ),
            this._language.language$,
        ])
            .pipe(
                switchMap(([next, lang]) => {
                    this._answers.loading.set(true);
                    return from(
                        this._execute('anagram', {
                            input: next.input
                                .toUpperCase()
                                .normalize('NFD')
                                .replace(/[^\w\s]|\d|_/g, '')
                                .replace(/\s+/g, '')
                                .split('')
                                .sort()
                                .join(''),
                            lang,
                        })
                    );
                })
            )
            .subscribe({
                next: (result: string[]) => {
                    this._answers.loading.set(false);
                    this._answers.list.set(result);
                },
                error: () => this._answers.loading.set(false),
            });
    }

    anagram(input: string): void {
        this._anagram.next({ input });
    }

    private _execute(command: string, payload: any): Promise<any> {
        const worker = this._createWorker();

        return new Promise((resolve, reject) => {
            if (worker === null) {
                reject(new Error('worker should not be null at this point'));
                return;
            }

            worker.onmessage = ({ data }) => {
                const { status, result, error } = data;
                if (status === 'completed') {
                    resolve(result);
                } else if (status === 'error') {
                    reject(new Error(error));
                }
            };

            worker.postMessage({ command, payload });
        });
    }

    private _createWorker(): Worker {
        if (this._worker !== null) {
            this._worker.terminate();
        }
        this._worker = new Worker(new URL('../app/app.worker', import.meta.url), { type: 'module' });
        return this._worker;
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
