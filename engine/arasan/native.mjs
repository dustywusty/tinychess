// Build the same pinned Arasan search for the server. No upstream files are edited.
import { execFileSync } from 'node:child_process';
import { mkdirSync, copyFileSync, existsSync } from 'node:fs';
import { resolve } from 'node:path';
const source = resolve('engine/arasan/vendor');
const out = resolve('engine/arasan/generated/native');
const commit = '361840f407b588252a48b25bb3000c751a470c89';
if (!existsSync(source)) {
 execFileSync('git', ['clone', '--no-checkout', '--filter=blob:none', 'https://github.com/jdart1/arasan-chess.git', source], { stdio: 'inherit' });
 execFileSync('git', ['-C', source, 'checkout', '--detach', commit], { stdio: 'inherit' });
}
if (execFileSync('git', ['-C', source, 'rev-parse', 'HEAD'], { encoding: 'utf8' }).trim() !== commit) throw new Error('Arasan source revision mismatch');
if (execFileSync('git', ['-C', source, 'status', '--porcelain', '--untracked-files=no'], { encoding: 'utf8' }).trim()) throw new Error('Use clean pinned Arasan sources');
mkdirSync(out, { recursive: true });
execFileSync('patch', ['-o', `${out}/protocol.cpp`, `${source}/src/protocol.cpp`, 'engine/arasan/no-tablebases.patch'], { stdio: 'inherit' });
const names = 'arasanx attacks bench bhash bitutil bitprobe board boardio bookread bookwrit calctime chess chessio eco epdrec globals hash input learn legal material movearr movegen notation options protocol scoring searchc search see stats tester threadc threadp evaluate'.split(' ');
execFileSync(process.env.CXX || 'c++', [
 ...names.map(name => name === 'protocol' ? `${out}/protocol.cpp` : `${source}/src/${name}.cpp`),
 `-I${source}/src`, `-I${source}/src/nnue`, ...(process.env.ARASAN_STATIC === '1' ? ['-static'] : []), '-std=c++17', '-O2', '-DNDEBUG', '-D_64BIT', '-pthread',
 ...(process.arch === 'arm64' ? ['-DSIMD', '-DNEON'] : ['-DSIMD', '-DSSE2', '-msse2']),
 '-DARASAN_VERSION=26.0', '-DNETWORK=arasanv8-20260906.nnue', '-o', `${out}/arasan`,
], { stdio: 'inherit' });
copyFileSync(`${source}/network/arasanv8-20260906.nnue`, `${out}/arasanv8-20260906.nnue`);
copyFileSync(`${source}/LICENSE`, `${out}/LICENSE.arasan`);
console.log(`${out}/arasan`);
