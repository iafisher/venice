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

syn keyword veniceBuiltin  case enum false input integer list map match print string true
syn keyword veniceKeywords elif else fn for if in let return struct while

syn match veniceComment "//.*$"

syn region veniceString start=+"+ skip=+\\\\\|\\"+ end=+"+

hi def link veniceBuiltin    Function
hi def link veniceComment    Comment
hi def link veniceKeywords   Statement
hi def link veniceString     String
