import { Component, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Subscription } from 'rxjs';
import { AnswersComponent } from '../answers/answers.component';
import { LanguageComponent } from '../language/language.component';
import { LanguageService } from '../services/language.service';
import { WorkerService } from '../services/worker.service';
import { Language } from '../types';

@Component({
    selector: 'app-root',
    templateUrl: './app.component.html',
    imports: [LanguageComponent, AnswersComponent, FormsModule],
})
export class AppComponent {
    private readonly _sub = new Subscription();
    private readonly _worker = inject(WorkerService);
    private readonly _lang = inject(LanguageService);

    searchText = '';
    highlight = '';
    sa: boolean | null = null;

    ngOnInit(): void {
        const params = new URLSearchParams(window.location.search);
        const q = params.get('q');
        if (q) {
            this.searchText = q;
        }
        const lang = params.get('lang') as any;
        if (lang && Object.values<string>(Language).includes(lang)) {
            this._lang.set(lang);
        }
        const highlight = params.get('highlight');
        if (highlight) {
            this.highlight = highlight;
        }
        const sa = params.get('sa');
        if (sa === '1') {
            this.sa = true;
        } else if (sa === '0') {
            this.sa = false;
        }
        this._sub.add(this._lang.onLanguageChange.subscribe(() => this.onSearch()));
        this.onSearch();
    }

    ngOnDestroy(): void {
        this._sub.unsubscribe();
    }

    onSearch(): void {
        this._worker.anagram(this.searchText);
    }
}
