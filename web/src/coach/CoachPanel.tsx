import { useState } from 'react';
import { formatScore, type Lesson } from './controller';
interface Props {
 lesson: Lesson | null; pending: boolean; error: string; moves: string[]; playerColor: string | null;
 feedback: { lesson: Lesson | null; pending: boolean; error: string };
 hidden: boolean; setHidden: (value: boolean) => void;
 openLesson: (lesson: Lesson, branch?: string) => void;
 command: (action: string, args?: Record<string, unknown>) => Promise<void>;
}
function positionLabel(score: { kind: 'cp' | 'mate'; value: number }, ply: number) {
 const side = ply % 2 ? 'Black' : 'White'; const opponent = ply % 2 ? 'White' : 'Black';
 if (score.kind === 'mate') return score.value > 0 ? `${side} has a mating line` : score.value < 0 ? `${opponent} has a mating line` : 'Checkmate';
 if (Math.abs(score.value) < 50) return 'Balanced position';
 return `${score.value > 0 ? side : opponent} is ahead`;
}
export function CoachPanel({ lesson, pending, error, moves, playerColor, feedback, hidden, setHidden, openLesson, command }: Props) {
 const [collapsed, setCollapsed] = useState(false);
 const [text, setText] = useState(''); const [selectedPly, setSelectedPly] = useState(-1);
 const lastOwn = moves.length - 1 - ((moves.length - 1) % 2 === (playerColor === 'black' ? 1 : 0) ? 0 : 1);
 const ply = selectedPly < 0 ? Math.max(0, lastOwn) : Math.min(selectedPly, moves.length - 1);
 const report = lesson ?? (hidden ? null : feedback.lesson);
 const assessment = hidden ? undefined : report?.assessment;
 const recommendation = hidden ? undefined : lesson?.recommendation;
 const stale = lesson?.currentPosition && moves.length !== lesson.ply;
 const run = (action: string, args: Record<string, unknown> = {}) => void command(action, { noSpoilers: hidden, ...args });
 const reviewing = !!lesson;
 const review = () => feedback.lesson && feedback.lesson.ply === ply && !hidden ? openLesson(feedback.lesson) : run('review', { ply });
 return <section className={`coach-panel ${reviewing ? 'is-reviewing' : ''}`} aria-label="Chess coach">
  <header className="move-coach-header">
   <button className="coach-widget-toggle" aria-label="Chess coach" aria-expanded={!collapsed} aria-controls="coach-widget-body" onClick={() => setCollapsed(value => !value)}><span className="eyebrow">A LITTLE INSIGHT</span><span className={`widget-chevron ${collapsed ? 'is-collapsed' : ''}`} aria-hidden="true">⌄</span></button>
  </header>
  <div id="coach-widget-body" hidden={collapsed}>
   {(reviewing || assessment) && <div className="coach-review-heading">
    {reviewing && <h2>{lesson.currentPosition ? `Current position — ${lesson.ply % 2 ? "Black" : "White"} to move` : `Reviewing move ${Math.floor(lesson.ply / 2) + 1} — ${lesson.ply % 2 ? 'Black' : 'White'}`}</h2>}
    {assessment && <span className={`move-grade ${assessment.tone}`}>{assessment.grade}</span>}
    {reviewing && <button className="coach-close" aria-label="Return to game" onClick={() => run('return')}>×</button>}
   </div>}
  {stale && <p className="coach-position-notice" role="status">The game has moved on. This suggestion is for the earlier position. <button className="text-button" disabled={pending} onClick={() => run('best_now')}>Analyze latest position</button></p>}
  {recommendation && <><div className="coach-recommendation">
   <button className="move-option suggested" aria-label="Preview suggested move" disabled={pending} onClick={() => run('alternative')}><span className="move-option-label">SUGGESTED MOVE</span><span className="move-option-value">{recommendation.move}<span className="move-score">{formatScore(recommendation.score)}</span></span><span className="move-option-note">Preview on the board ↗</span></button>
  </div><p className="move-insight">{recommendation.summary}</p></>}
  {assessment && report && <>
   <div className="move-comparison">
    <button className={`move-option ${lesson?.branch === 'played' ? 'active' : ''}`} onClick={() => reviewing ? run('played') : openLesson(report, 'played')} disabled={pending} aria-label="Your move">
     <span className="move-option-label">{report.ply % 2 === (playerColor === 'black' ? 1 : 0) ? 'YOU PLAYED' : 'PLAYED'}</span>
     <span className="move-option-value">{assessment.played}<span className="move-score">{assessment.playedScore.kind === "cp" && assessment.playedScore.value > 0 ? "+" : ""}{formatScore(assessment.playedScore)}</span></span>
     <span className="move-option-note">{positionLabel(assessment.playedScore, report.ply)}</span>
    </button>
    <button className={`move-option suggested ${lesson?.branch === 'alternative' ? 'active' : ''}`} onClick={() => reviewing ? run('alternative') : openLesson(report, 'alternative')} disabled={pending || hidden} aria-label="Alternative">
     <span className="move-option-label">{assessment.played === assessment.best ? 'BEST MOVE' : assessment.tone === 'warning' || assessment.tone === 'caution' ? 'BETTER HERE' : 'ENGINE’S CHOICE'}</span>
     <span className="move-option-value">{assessment.best}<span className="move-score">{assessment.bestScore.kind === "cp" && assessment.bestScore.value > 0 ? "+" : ""}{formatScore(assessment.bestScore)}</span></span>
     <span className="move-option-note">{assessment.played === assessment.best ? 'You found it. Nice move.' : 'See it on the board ↗'}</span>
    </button>
   </div>
   <p className="move-insight" aria-live="polite">{assessment.summary}</p>
  </>}
  {!assessment && !reviewing && <p className="move-coach-empty">{feedback.pending ? 'Taking a look at your last move…' : hidden ? 'Hints only. The answer stays hidden until you choose to reveal it.' : 'Make a move. I’ll show how it stacks up and what you could play instead.'}</p>}
  {reviewing && <>
   <div className="coach-controls coach-playback">
    {!lesson.currentPosition && <button aria-pressed={lesson.branch === 'compare'} disabled={pending || hidden || lesson.reveal < 3} onClick={() => run('compare')}>Compare</button>}
    {lesson.reveal === 3 && lesson.branch !== 'compare' && <button aria-pressed={lesson.step === 0} disabled={pending || hidden} onClick={() => run('before')}>Before</button>}
    <span className="coach-step">{lesson.branch === "compare" ? "Choose a move above to play it" : lesson.step === 0 ? "Starting position" : lesson.step === 1 ? lesson.currentPosition ? "Move preview" : "Move played" : `Reply ${lesson.step - 1}`}</span>
    {lesson.branch !== 'compare' && <><button disabled={pending || lesson.step === 0 || lesson.reveal < 2} aria-label="Back" onClick={() => run('back')}>← Back</button>
    <button disabled={pending || lesson.step >= lesson.steps || lesson.reveal < 3 || hidden} aria-label="Continue" onClick={() => run('next')}>Continue →</button></>}
   </div>
   {lesson.mode === 'exercising' && <p className="move-insight">Try a move on the board. It stays in this exercise.</p>}
   {((!assessment && !recommendation) || lesson.mode === 'exercising') && <p className="move-insight" aria-live="polite">{lesson.caption}</p>}
   {lesson.reveal < 3 && <button className="coach-action" disabled={pending || hidden} onClick={() => run('reveal')}>Reveal next hint</button>}
   {lesson.choices && <div className="coach-controls">{lesson.choices.map(choice => <button key={choice.id} disabled={pending} onClick={() => run('candidate', { candidate: choice.move })}>{choice.label}</button>)}</div>}
  </>}
  <div className="coach-footer">
   <button className="coach-action" disabled={pending} onClick={() => run('best_now')}>{lesson?.currentPosition ? 'Analyze latest' : 'Best move now'}</button>
   {!reviewing && <button className="coach-action" disabled={pending || !moves.length} onClick={review}>Review move <span aria-hidden="true">↗</span></button>}
   <details className="coach-more">
    <summary>{reviewing ? 'Ask & explore' : 'Ask the coach'} <span aria-hidden="true">+</span></summary>
    <div className="coach-expanded">
     {reviewing && <>{(assessment || recommendation) && lesson.mode !== "exercising" && <p className="coach-answer" aria-live="polite">{lesson.caption}</p>}<div className="coach-controls"><button disabled={pending} onClick={() => run('explain')}>Why?</button><button disabled={pending} onClick={() => run('hint')}>Just a hint</button><button disabled={pending} onClick={() => run('exercise')}>Try it</button></div></>}
     <form onSubmit={event => { event.preventDefault(); if (text.trim()) { run('language', { text, ply }); setText(''); } }}>
      <label className="sr-only" htmlFor="coach-question">Ask about this lesson</label>
      <div className="coach-input-row"><input id="coach-question" maxLength={240} value={text} onChange={event => setText(event.target.value)} placeholder={reviewing ? "Why this move? What if…?" : "What should I play next?"} /><button disabled={pending || !text.trim()} aria-label="Ask">↑</button></div>
     </form>
     <div className="coach-preferences"><label className="coach-hide"><input type="checkbox" checked={hidden} onChange={event => { setHidden(event.target.checked); if (event.target.checked && lesson) run('exercise', { noSpoilers: true }); }} />Hints only</label>
      {!reviewing && <label>Review <select aria-label="Move to review" value={ply} onChange={event => setSelectedPly(Number(event.target.value))}>{moves.map((move, index) => <option key={index} value={index}>{Math.floor(index / 2) + 1} · {index % 2 ? 'Black' : 'White'} · {move}</option>)}</select></label>}
     </div>
     {reviewing && <details className="coach-search-details"><summary>About this analysis</summary><p>{lesson.quality}. Scores compare the position for the player who made this move; one pawn is the unit, not a captured piece count.</p></details>}
    </div>
   </details>
  </div>
  {pending && <div className="coach-loading" role="status">Preparing your lesson… {!lesson && <button onClick={() => run('return')}>Cancel review</button>}</div>}
  {(error || feedback.error) && <div role="alert" className="coach-loading">{error || feedback.error}<button disabled={pending} onClick={() => error ? run('retry') : review()}>{error.includes('analysis expired') ? 'Refresh analysis' : 'Retry request'}</button></div>}
  </div>
 </section>;
}
