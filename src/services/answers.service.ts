import { Injectable, signal } from '@angular/core';

@Injectable({ providedIn: 'root' })
export class AnswersService {
    readonly list = signal<string[]>([]);
    readonly loading = signal<boolean>(false);
}
