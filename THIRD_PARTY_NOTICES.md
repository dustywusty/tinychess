# Computer engine notices

Your Move retains its MIT license. The computer opponent uses Arasan under the MIT license.

The build pins Arasan to commit `361840f407b588252a48b25bb3000c751a470c89`.
The same license covers its `arasanv8-20260906.nnue` network.
The build excludes the Windows GUI, opening book, tablebases, and Git submodules.

The build replaces console input with `engine/arasan/input.cpp`.
It also applies `engine/arasan/no-tablebases.patch` to guard optional tablebase settings.
It does not change the search algorithm or network.

Emscripten 4.0.14 compiles the runtime. Its JavaScript runtime uses MIT/NCSA terms.
The bundled C/C++ libraries include musl and LLVM components with their own permissive licenses.
The build copies their full notices into `web/public/arasan/NOTICES.txt` and the mobile engine asset.

Upstream sources:

- [Arasan source and license](https://github.com/jdart1/arasan-chess/tree/361840f407b588252a48b25bb3000c751a470c89).
- [Emscripten 4.0.14 source and license](https://github.com/emscripten-core/emscripten/tree/4.0.14).

The current build does not include Stockfish. Earlier releases and Git history can still contain Stockfish.
Replacing the engine does not remove obligations for those earlier distributions.

This file covers the new engine distribution. It does not replace the licenses of existing application dependencies.
