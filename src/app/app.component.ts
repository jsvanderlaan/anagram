import { Component, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Subscription } from 'rxjs';
import { AnswersComponent } from '../answers/answers.component';
import { LanguageComponent } from '../language/language.component';
import { LanguageService } from '../services/language.service';
import { WorkerService } from '../services/worker.service';

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

    ngOnInit(): void {
        // get query parameter 'q' and set it to searchText
        const params = new URLSearchParams(window.location.search);
        const q = params.get('q');
        if (q) {
            this.searchText = q;
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
