import { ChangeDetectionStrategy, Component, computed, signal } from '@angular/core';
import { AnswersService } from '../services/answers.service';

@Component({
    selector: 'answers',
    standalone: true,
    imports: [],
    changeDetection: ChangeDetectionStrategy.OnPush,
    templateUrl: './answers.component.html',
})
export class AnswersComponent {
    readonly previewSize = 100;
    readonly showAll = signal<boolean>(false);
    readonly preview = computed(() => [...this.answers.list()].splice(0, this.previewSize));
    constructor(public answers: AnswersService) {}
}
