" Vim syntax file
" Language:          Venice
" Maintainer:        Ian Fisher
" Latest Revision:   3 July 2021
"
" Based on https://vim.fandom.com/wiki/Creating_your_own_syntax_files
"   and /usr/share/vim/vim80/syntax/rust.vim

if exists("b:current_syntax")
	finish
endif
let b:current_syntax = "venice"

syn keyword veniceBuiltin  bool false input int length list map maximum minimum print string true uint
syn keyword veniceKeywords alias and as break case class constructor continue default enum elif else exception export for from func if import in interface let match new not or public private return self throw var while with

syn match veniceLineComment "#.*$"
syn region veniceBlockComment start="^###" end="^###"

syn region veniceString start=+"+ skip=+\\\\\|\\"+ end=+"+

hi def link veniceBuiltin        Function
hi def link veniceLineComment    Comment
hi def link veniceBlockComment   Comment
hi def link veniceKeywords       Statement
hi def link veniceString         String