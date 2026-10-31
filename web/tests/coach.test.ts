import assert from 'node:assert/strict';
import test from 'node:test';
import { coachReducer, initialCoach, type Lesson } from '../src/coach/controller.ts';
const lesson = { id: 'lesson', revision: 1, mode: 'reviewing' } as Lesson;
test('return and later requests invalidate delayed lesson events', () => {
 const loading = coachReducer(initialCoach, { type: 'request' });
 const live = coachReducer(loading, { type: 'return' });
 assert.deepEqual(coachReducer(live, { type: 'result', revision: loading.revision, lesson }), live);
 const next = coachReducer(loading, { type: 'request' });
 assert.deepEqual(coachReducer(next, { type: 'result', revision: loading.revision, lesson }), next);
});
test('review state has no authoritative move history', () => {
 const loading = coachReducer(initialCoach, { type: 'request' });
 const review = coachReducer(loading, { type: 'result', revision: loading.revision, lesson });
 assert.equal(review.lesson, lesson);
 assert.equal('uci' in review, false);
 assert.equal(coachReducer(review, { type: 'return' }).lesson, null);
});

test('navigation uses validated local frames and never reveals hidden views', async () => {
 const { navigate } = await import('../src/coach/controller.ts');
 const base = { ...lesson, reveal: 3, branch: 'played', step: 1, views: {
  alternative: [{ caption: 'Starting position', fen: 'before', arrows: [], score: { kind: 'cp' as const, value: 10 } }, { caption: 'Alternative: e4', fen: 'after', arrows: [], score: { kind: 'cp' as const, value: 10 } }],
 } };
 assert.equal(navigate(base, 'alternative')?.fen, 'after');
 assert.equal(navigate({ ...base, reveal: 0 }, 'alternative'), null);
 const selected = navigate(base, 'alternative')!;
 assert.equal(navigate(selected, 'back')?.fen, 'before');
 assert.equal(selected.revision, lesson.revision);
 assert.equal(selected.caption, 'Alternative: e4');
});

test('mate distance remains distinct from centipawns in the visible score', async () => {
 const { formatScore } = await import('../src/coach/controller.ts');
 assert.equal(formatScore({ kind: 'mate', value: -3 }), 'Faces mate in 3');
 assert.equal(formatScore({ kind: 'mate', value: 2 }), 'Mate in 2');
 assert.equal(formatScore({ kind: 'cp', value: -120 }), '-1.20 pawns');
});
